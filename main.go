package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 토이 프로젝트용 CORS 허용
	},
}

var urlStore sync.Map

func main() {
	// 1. 환경변수 로드
	err := godotenv.Load()
	if err != nil {
		log.Println("경고: .env 파일을 찾을 수 없습니다.")
	}

	// 2. 동시 편집 워크스페이스
	hub := NewHub()
	go hub.Run()

	// 3. 라우터 설정
	// 초대 링크 생성 API (프론트엔드에서 생성한 ID를 받도록 수정)
	http.HandleFunc("/api/invite", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST만 허용됩니다", http.StatusMethodNotAllowed)
			return
		}

		// 프론트엔드가 만들어서 보낸 ID를 해독
		var req struct {
			ShortID string `json:"shortId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "잘못된 요청입니다.", http.StatusBadRequest)
			return
		}

		// 메모리에 매핑 저장 (단축ID -> "main-room")
		urlStore.Store(req.ShortID, "main-room")

		// 성공적으로 저장했다고 응답만 제공
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	})

	// 단축 URL 리다이렉트 라우터 (/r/...)
	http.HandleFunc("/r/", func(w http.ResponseWriter, r *http.Request) {
		// URL에서 "/r/a1b2c"의 "a1b2c" 부분만 추출
		shortID := strings.TrimPrefix(r.URL.Path, "/r/")

		// 메모리 DB에서 조회
		if _, ok := urlStore.Load(shortID); ok {
			// 찾았다면 원래 워크스페이스(루트 주소)로 리다이렉트
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// 못 찾았다면 에러 처리
		http.Error(w, "유효하지 않거나 만료된 초대 링크입니다.", http.StatusNotFound)
	})

	// 프론트엔드 화면 서빙
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// AI 스트리밍 담당 (ai_handler.go 의 함수 호출)
	http.HandleFunc("/ask/stream", HandleAIStream)

	// 웹소켓 워크스페이스 담당
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println("웹소켓 업그레이드 실패:", err)
			return
		}

		client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
		client.hub.register <- client

		go client.writePump()
		go client.readPump()
	})

	// 리포트 생성 및 조회 API
	http.HandleFunc("/report/generate", GenerateReportHandler)
	http.HandleFunc("/api/report", GetReportAPIHandler)

	// /report 페이지로 접속하면 report.html 화면을 띄워줌
	http.HandleFunc("/report", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "report.html")
	})

	// 4. 서버 실행
	fmt.Println("🚀 서버가 8080 포트에서 실행 중입니다... (http://localhost:8080)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("서버 에러: %v\n", err)
	}
}
