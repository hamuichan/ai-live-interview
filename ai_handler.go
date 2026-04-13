package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// HandleAIStream: AI 면접관의 답변을 SSE로 스트리밍해 주는 전담 핸들러
func HandleAIStream(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "스트리밍을 지원하지 않는 브라우저입니다.", http.StatusInternalServerError)
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

	iter := model.GenerateContentStream(ctx, prompt)

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("스트리밍 에러 발생: %v", err)
			break
		}

		for _, part := range resp.Candidates[0].Content.Parts {
			text := fmt.Sprintf("%v", part)
			text = strings.ReplaceAll(text, "\n", "<br>") // 멀티라인 버그 방지

			fmt.Fprintf(w, "data: %v\n\n", text)
			flusher.Flush()
		}
	}
}
