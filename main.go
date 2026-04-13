package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// HTTP 연결을 WebSocket 연결로 업그레이드해 주는 객체
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CORS 에러 방지: 일단 모든 도메인에서의 접속을 허용 (토이 프로젝트용)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	// .env 파일 로드
	err := godotenv.Load()
	if err != nil {
		log.Println("경고: .env 파일을 찾을 수 없습니다. OS 환경변수를 사용합니다.")
	}

	hub := NewHub()
	go hub.Run()

	// 기존 기본 라우터
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// AI 대화 라우터
	http.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background() // Go에서 외부 통신을 할 때 생명주기를 관리하는 객체

		// 환경변수에서 API 키 읽어오기
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			http.Error(w, "앗! API 키가 설정되지 않았습니다.", http.StatusInternalServerError)
			return
		}

		// Gemini 클라이언트 생성
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			http.Error(w, fmt.Sprintf("클라이언트 생성 실패: %v", err), http.StatusInternalServerError)
			return
		}
		defer client.Close() // 함수가 끝나면 클라이언트를 깔끔하게 닫아줍니다.

		// 사용할 모델 선택 (Gemini 3.1 Flash Lite 프리뷰 모델)
		model := client.GenerativeModel("gemini-3.1-flash-lite-preview")

		// AI에게 던질 첫 프롬프트!
		prompt := genai.Text("안녕? 너는 10년 차 시니어 백엔드 개발자 면접관이야. 나에게 짧고 친근하게 첫인사를 건네줘.")

		// 답변 생성 (이 부분이 실제 API를 호출하는 곳입니다)
		resp, err := model.GenerateContent(ctx, prompt)
		if err != nil {
			http.Error(w, fmt.Sprintf("AI 응답 에러: %v", err), http.StatusInternalServerError)
			return
		}

		// 브라우저 화면에 AI의 응답 출력
		for _, part := range resp.Candidates[0].Content.Parts {
			fmt.Fprintf(w, "🤖 AI 면접관: %v", part)
		}
	})

	// SSE 스트리밍 라우터
	http.HandleFunc("/ask/stream", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()

		// SSE를 위한 필수 HTTP 헤더 세팅
		w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		// Flusher 확인
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "이 브라우저는 스트리밍을 지원하지 않습니다.", http.StatusInternalServerError)
			return
		}

		apiKey := os.Getenv("GEMINI_API_KEY")
		client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
		if err != nil {
			http.Error(w, "AI 클라이언트 생성 실패", http.StatusInternalServerError)
			return
		}
		defer client.Close()

		model := client.GenerativeModel("gemini-3.1-flash-lite-preview")
		prompt := genai.Text("안녕? 너는 10년 차 시니어 백엔드 개발자 면접관이야. 나에게 짧고 친근하게 첫인사를 건네줘.")

		// 기존 GenerateContent 대신 GenerateContentStream 호출
		iter := model.GenerateContentStream(ctx, prompt)

		for {
			resp, err := iter.Next()  // AI가 만든 단어를 한 뭉치씩 가져옵니다.
			if err == iterator.Done { // AI가 말을 다 끝냈다면?
				break // 반복문을 탈출합니다.
			}
			if err != nil {
				log.Printf("스트리밍 에러 발생: %v", err)
				break
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				// AI가 준 답변(part)을 문자열로 확실하게 변환
				text := fmt.Sprintf("%v", part)

				// 문자열 속의 엔터(\n)를 모두 HTML의 <br> 태그로 바꿈
				text = strings.ReplaceAll(text, "\n", "<br>")

				// SSE 통신 규격에 맞게 "data: 내용\n\n" 형태로 출력
				fmt.Fprintf(w, "data: %v\n\n", text)

				// 모아두지 말고 브라우저로 즉시 발사 (이게 스트리밍의 핵심)
				flusher.Flush()
			}
		}
	})

	// 웹소켓 라우터
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("웹소켓 업그레이드 실패:", err)
			return
		}

		// 새로운 참여자(Client) 생성
		client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}

		// 방장(Hub)의 입장 우편함에 나를 등록
		client.hub.register <- client

		// 데이터를 읽고 쓰는 두 명의 요정(고루틴)을 백그라운드에 띄움
		go client.writePump() // 서버 -> 브라우저
		go client.readPump()  // 브라우저 -> 서버
	})

	fmt.Println("서버가 8080 포트에서 실행 중입니다... (http://localhost:8080)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("서버 에러: %v\n", err)
	}
}
