package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// 메모리 DB: 생성된 리포트를 서버 메모리에 안전하게 저장
// 여러 사용자가 동시에 접속해도 꼬이지 않도록 Go 언어의 동시성 안전(Thread-safe) Map을 사용
var reportStore sync.Map

// 프론트엔드에서 받을 데이터 구조체
type ReportRequest struct {
	Code    string `json:"code"`
	ChatLog string `json:"chat_log"`
}

// 리포트 생성 핸들러 (AI 호출 및 저장)
func GenerateReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST 요청만 지원합니다.", http.StatusMethodNotAllowed)
		return
	}

	var req ReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "데이터 파싱 에러", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		http.Error(w, "AI 클라이언트 에러", http.StatusInternalServerError)
		return
	}
	defer client.Close()

	// 리포트는 스트리밍이 아니라 한 번에 받아옴 (GenerateContent)
	model := client.GenerativeModel("gemini-3-flash-preview")
	promptText := fmt.Sprintf(`너는 엄격하지만 친절한 시니어 백엔드 면접관이야.
다음은 지원자의 [최종 코드]와 [면접 대화 내역]이야. 이를 바탕으로 마크다운 형식의 평가 리포트를 작성해줘.
반드시 다음 3가지 항목을 포함해야 해:
1. 총평 및 예상 점수 (100점 만점)
2. 잘한 점 (강점)
3. 아쉬운 점 (개선 포인트)

[최종 코드]
%s

[대화 내역]
%s`, req.Code, req.ChatLog)

	resp, err := model.GenerateContent(ctx, genai.Text(promptText))
	if err != nil {
		log.Println("리포트 생성 실패:", err)
		http.Error(w, "리포트 생성 실패", http.StatusInternalServerError)
		return
	}

	// 응답 텍스트 추출
	var reportMarkdown string
	for _, part := range resp.Candidates[0].Content.Parts {
		reportMarkdown += fmt.Sprintf("%v", part)
	}

	// 고유 ID 생성 (무작위 8자리 헥사코드 생성, 예: "8f3a9b2c")
	b := make([]byte, 4)
	rand.Read(b)
	id := fmt.Sprintf("%x", b)

	// 메모리 DB에 저장 (Key: id, Value: 마크다운 텍스트)
	reportStore.Store(id, reportMarkdown)

	// 프론트엔드에 성공적으로 생성된 ID 반환
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": id})
}

// 리포트 데이터 조회 API (새 페이지에서 호출할 용도)
func GetReportAPIHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if report, ok := reportStore.Load(id); ok {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"markdown": report.(string)})
		return
	}
	http.Error(w, "리포트를 찾을 수 없습니다.", http.StatusNotFound)
}
