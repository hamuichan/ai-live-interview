package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

func main() {
	// ⭐️ 서버가 켜지자마자 가장 먼저 .env 파일을 읽어옵니다!
	err := godotenv.Load()
	if err != nil {
		log.Println("경고: .env 파일을 찾을 수 없습니다. OS 환경변수를 사용합니다.")
	}

	// 1. 기존 기본 라우터
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello! AI Live Interview 서버가 성공적으로 켜졌습니다 🚀")
	})

	// 2. 새로운 AI 대화 라우터 추가!
	http.HandleFunc("/ask", func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background() // Go에서 외부 통신을 할 때 생명주기를 관리하는 객체입니다.

		// 환경변수에서 API 키 읽어오기 (하드코딩은 위험하니까요!)
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

		// 가볍고 빠른 모델인 gemini-1.5-flash 선택
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

	fmt.Println("서버가 8080 포트에서 실행 중입니다... (http://localhost:8080)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("서버 에러: %v\n", err)
	}
}
