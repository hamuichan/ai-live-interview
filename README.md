# 🤖 AI Live Interview

> **Go 언어의 강력한 동시성 처리와 최신 프론트엔드 통신 기법을 결합한 실시간 AI 협업 플랫폼입니다.**
> 단순한 챗봇을 넘어, 지원자가 작성 중인 코드를 실시간으로 읽고(Context-Aware) 피드백을 스트리밍하는 혁신적인 면접 경험을 제공합니다.

## 🌟 기획 배경 및 목표
- 기존 단방향 AI 챗봇의 한계를 넘어, "내 코드를 실시간으로 보고 피드백해 주는 진짜 면접관"과의 인터랙션을 구현하고자 했습니다.
- Go 언어의 `Goroutine`과 `Channel`을 활용하여 다중 클라이언트의 실시간 커넥션(WebSocket, SSE)을 안정적으로 처리하는 **고성능 동시성 서버 아키텍처** 설계를 목표로 했습니다.

## 🚀 핵심 기능 (Core Features)

### 1. 🧠 맥락 인지(Context-Aware) AI 스트리밍
- **코드 인지형 채팅:** 프론트엔드의 `fetch` API(ReadableStream)를 활용해 현재 작성 중인 긴 코드와 질문을 `POST`로 전송, AI가 문맥을 파악해 답변을 제공.
- **실시간 마크다운 렌더링:** Server-Sent Events(SSE) 파이프라인을 구축하여 AI의 응답을 타자 치듯 스트리밍하고, `marked.js`로 코드 블록과 스타일을 즉각 렌더링.

### 2. ⚡ 실시간 동시 편집 워크스페이스
- **WebSocket Hub 아키텍처:** 다수의 사용자가 동일한 워크스페이스에서 충돌 없이 코드를 동시 편집할 수 있는 Thread-safe 통신 환경 구축.
- **다국어 Syntax Highlighting:** `CodeMirror` 엔진을 연동하여 Go, Python, JavaScript 등 다국어 문법 하이라이팅 및 에디터 환경 제공.

### 3. 📊 종합 평가 리포트 & 🔗 URL 단축 라우팅
- **리포트 생성기:** 면접 종료 시 전체 대화 Context를 모아 AI에게 전달하여 종합 마크다운 리포트 생성 및 인메모리(`sync.Map`) 저장.
- **커스텀 URL 단축기 (초대 기능):** 무작위 5자리 난수(Short ID)를 생성하여 동적 라우터(`/r/{id}`)를 통해 워크스페이스 접속 리다이렉션 트래픽 처리.

## 🛠 기술 스택 (Tech Stack)
- **Backend:** Go (Golang), `sync.Map` (In-memory DB)
- **Frontend:** HTML5, CSS3 (CSS Variables, Flexbox), Vanilla JavaScript (ES6+), CodeMirror 5, Marked.js
- **Network / API:** WebSocket, `fetch` ReadableStream, Google Gemini API
- **Architecture:** SoC (Separation of Concerns), Hub & Client Pattern

---

## 🔥 트러블슈팅 및 기술적 의사결정 (ADR)
> 개발 과정에서 마주친 깊이 있는 고민과 아키텍처 개선 과정을 기록합니다.

### 1. AI 스트리밍: `EventSource (GET)` 한계 극복 및 `Fetch POST` 전환
- **문제:** 기존 `EventSource`는 GET 요청만 지원하여, 길이가 긴 워크스페이스 코드를 AI에게 보낼 때 URL 길이 제한(URL Too Long) 및 보안 문제가 발생.
- **해결:** 프론트엔드 통신 방식을 `fetch` API 기반의 `ReadableStream`으로 전면 개편. 코드를 POST Body에 담아 안전하게 전송하면서도, 청크(Chunk) 단위로 쪼개어 들어오는 데이터를 디코딩해 스트리밍 UI를 완벽하게 유지함.

### 2. WebSocket 상태 동기화 누락 버그 해결 (Late-joiner 문제)
- **문제:** 신규 사용자가 워크스페이스에 늦게 접속(Join)할 경우, 클라이언트의 초기 '빈 문자열' 상태가 서버를 통해 브로드캐스트되어 기존 사용자의 코드가 모두 날아가는 치명적 버그 발생.
- **해결:** 단순 메시지 브로드캐스트 릴레이 역할만 하던 Hub 구조체에 `content` 상태 변수를 추가(메모리 캐싱). 신규 접속 이벤트(`register` 채널) 발생 시, 서버가 들고 있던 최신 상태를 해당 클라이언트에게 즉시 전송(`sync`)하도록 동기화 파이프라인 재설계.

### 3. Safari 비동기 클립보드 API 차단 우회 (Fallback Strategy)
- **문제:** '초대 링크 생성' 시 서버에서 단축 URL을 받아와 클립보드에 복사할 때, Safari 브라우저가 "비동기(fetch 대기) 직후의 복사"를 악성 스크립트로 간주하여 `NotAllowedError` 발생.
- **해결:** 서버 응답을 기다리지 않고 프론트엔드에서 난수 ID를 즉시(동기적으로) 생성하여 복사를 선행한 뒤, 백엔드에 해당 매핑 정보를 비동기(POST)로 통보하는 역방향 아키텍처로 선회하여 크로스 브라우징 완벽 대응.

### 4. 반응형 UI Layout Shift 및 텍스트 오버플로우 방어
- **문제:** Flexbox 기반 레이아웃에서 채팅이 길어지거나 화면이 좁아질 때 버튼들이 워크스페이스를 침범하거나, 마크다운 코드 블록이 말풍선을 뚫고 나가는 현상 발생.
- **해결:** `min-height: 0` 속성을 활용해 Flex 자식 요소의 무한 팽창 억제. 코드 블록에 `word-break` 및 `overflow-x: auto`를 적용하여 말풍선 내부 가로 스크롤 생성. 버튼 그룹에는 `flex-wrap: wrap`과 Graceful Degradation 설계를 적용해 우아한 UI 강등 구현.

---

## ⚙️ 로컬 실행 방법 (How to Run)

**1. 레포지토리 클론 및 폴더 이동**
```bash
$ git clone [https://github.com/username/ai-live-interview.git](https://github.com/username/ai-live-interview.git)
$ cd ai-live-interview
```

**2. 환경변수 설정 (`.env`)**
프로젝트 루트 디렉토리에 `.env` 파일을 생성하고 Gemini API 키를 입력합니다.
```env
GEMINI_API_KEY=your_gemini_api_key_here
```

**3. Go 모듈 다운로드 및 서버 실행**
```bash
$ go mod tidy
# main.go 뿐만 아니라 분리된 핸들러 파일들을 함께 실행해야 합니다.
$ go run *.go 
```

**4. 접속**
브라우저를 열고 `http://localhost:8080` 에 접속하여 실시간 면접을 시작하세요!

---
