package main

import (
	"github.com/gorilla/websocket"
)

// ==========================================
// 1. Client: 웹소켓에 접속한 참여자 1명을 의미
// ==========================================
type Client struct {
	hub  *Hub            // 자기가 속한 방(Hub)의 주소
	conn *websocket.Conn // 브라우저와 연결된 실제 웹소켓 파이프
	send chan []byte     // 이 사람에게 보낼 메시지를 담아두는 '개인 우편함(채널)'
}

// ==========================================
// 2. Hub: 클라이언트들을 관리하고 메시지를 방송(Broadcast)하는 관리자
// ==========================================
type Hub struct {
	clients    map[*Client]bool // 현재 접속 중인 사람들 명단 (출석부)
	broadcast  chan []byte      // 누군가 전체 공지를 날릴 때 쓰는 '방송용 우편함'
	register   chan *Client     // 새로운 사람이 입장할 때 쓰는 '입장 우편함'
	unregister chan *Client     // 사람이 퇴장할 때 쓰는 '퇴장 우편함'
	content    []byte           // 현재 공유되는 코드 내용
}

// 처음 방(Hub)을 만들 때 초기화해 주는 함수
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		content:    []byte(""),
	}
}

// ==========================================
// 3. Hub의 핵심 뇌: 쉬지 않고 우편함을 감시하는 무한 루프
// ==========================================
func (h *Hub) Run() {
	for {
		// select는 여러 채널(우편함) 중 "어느 우편함에 편지가 왔나?"를 동시에 감시
		select {

		// 1) 누군가 입장 우편함에 들어왔다면?
		case client := <-h.register:
			h.clients[client] = true // 출석부에 이름 적기
			client.send <- h.content // 현재 공유되는 코드 내용 보내주기

		// 2) 누군가 퇴장 우편함에 들어왔다면?
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok { // 출석부에 있는 사람인지 확인
				delete(h.clients, client) // 출석부에서 지우기
				close(client.send)        // 그 사람의 개인 우편함 폐기
			}

		// 3) 누군가 방송용 우편함에 메시지를 넣었다면? (코드 수정 발생!)
		case message := <-h.broadcast:
			h.content = message // 현재 공유되는 코드 내용 업데이트
			// 출석부에 있는 "모든 사람"에게 편지를 복사해서 쫙 돌림
			for client := range h.clients {
				select {
				case client.send <- message: // 각자의 개인 우편함에 메시지 넣기
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// ==========================================
// 4. ReadPump: 브라우저에서 오는 메시지를 계속 읽어서 방장에게 넘김
// ==========================================
func (c *Client) readPump() {
	defer func() {
		// 에러가 나거나 브라우저가 닫히면 방장에게 '저 퇴장합니다'라고 알림
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break // 연결이 끊기면 무한 루프 탈출
		}
		// 읽은 메시지를 방장의 방송용 우편함에 넣기
		c.hub.broadcast <- message
	}
}

// ==========================================
// 5. WritePump: 방장이 개인 우편함에 넣어준 메시지를 브라우저로 쏴줌
// ==========================================
func (c *Client) writePump() {
	defer c.conn.Close()

	for {
		// 내 우편함(c.send)에 편지가 올 때까지 대기하다가 꺼냄
		message, ok := <-c.send
		if !ok {
			// 방장이 우편함을 닫아버렸다면 연결 종료
			c.conn.WriteMessage(websocket.CloseMessage, []byte{})
			return
		}

		// 꺼낸 편지를 브라우저 화면(웹소켓)으로 전송
		c.conn.WriteMessage(websocket.TextMessage, message)
	}
}
