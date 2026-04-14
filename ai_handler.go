package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// 프론트엔드에서 보낼 JSON 데이터를 받을 구조체
type AIRequest struct {
	Question string `json:"question"`
	Code     string `json:"code"`
}

func HandleAIStream(w http.ResponseWriter, r *http.Request) {
	// 1. POST 요청만 허용
	if r.Method != http.MethodPost {
		http.Error(w, "POST 요청만 지원합니다.", http.StatusMethodNotAllowed)
		return
	}

	// 2. 브라우저가 보낸 질문과 코드를 Decode
	var req AIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "잘못된 데이터 형식입니다.", http.StatusBadRequest)
		return
	}

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

	model := client.GenerativeModel("gemini-3-flash-preview")

	// 3. '맥락 인지' 프롬프트 조립
	// 지원자의 코드와 질문을 합쳐서 AI 면접관에게 쥐어줌.
	promptText := fmt.Sprintf(`너는 10년 차 시니어 백엔드 개발자 면접관이야.
다음은 지원자가 현재 워크스페이스에서 작성 중인 코드야:
---
%s
---
지원자의 질문: %s

이 코드의 문맥을 파악해서, 지원자의 질문에 친절하고 핵심만 짚어서 답변해 줘.`, req.Code, req.Question)

	prompt := genai.Text(promptText)
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
			text = strings.ReplaceAll(text, "\n", "<br>")
			fmt.Fprintf(w, "data: %v\n\n", text)
			flusher.Flush()
		}
	}
}
