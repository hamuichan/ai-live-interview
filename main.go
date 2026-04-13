package main

import (
	"fmt"
	"log"
	"net/http"

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

	// 4. 서버 실행
	fmt.Println("🚀 서버가 8080 포트에서 실행 중입니다... (http://localhost:8080)")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("서버 에러: %v\n", err)
	}
}
