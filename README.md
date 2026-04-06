# 🤖 AI Live Interview (AI 기반 실시간 모의 면접 & 워크스페이스)

> **Go 언어의 강력한 동시성 처리와 SSE를 활용한 실시간 AI 모의 면접 플랫폼입니다.**
> 지원자와 AI 면접관이 실시간으로 대화하며, 작성 중인 코드를 동시에 편집하고 리뷰할 수 있는 환경을 제공합니다.

## 🌟 기획 배경 및 목표
- 기존의 단방향 AI 챗봇의 한계를 넘어, **실제 면접관과 대화하는 듯한 실시간 인터랙션**을 구현하고자 했습니다.
- Go 언어의 `Goroutine`을 활용하여 수많은 실시간 커넥션(SSE, WebSocket)을 안정적으로 처리하는 **고성능 동시성 서버 아키텍처**를 설계하는 것을 목표로 합니다.

## 🚀 핵심 기능 (Core Features)
1. **AI 실시간 모의 면접 (SSE 통신)**
   - OpenAI/Gemini API를 활용한 페르소나 기반 면접관 챗봇
   - Server-Sent Events(SSE)를 통한 끊김 없는 실시간 텍스트 스트리밍
2. **실시간 동시 편집 워크스페이스 (WebSocket)**
   - 면접자와 AI(또는 스터디원)가 함께 코드를 작성하고 수정하는 실시간 에디터
   - 상태 동기화 및 충돌 해결 아키텍처 적용
3. **면접 결과 리포트 & 커스텀 URL 단축기**
   - 면접 종료 후 분석 리포트 생성 및 Redis 기반 캐싱
   - 공유하기 쉬운 커스텀 단축 URL 생성 및 트래픽 분산 처리

## 🛠 기술 스택 (Tech Stack)
- **Backend:** Go (Golang)
- **Database / Cache:** Redis, (추후 RDB 추가 예정)
- **Real-time Comm:** SSE (Server-Sent Events), WebSockets
- **Infrastructure:** Docker, Nginx (추후 배포 시 적용)
- **External API:** OpenAI API / Gemini API

## 🏛 아키텍처 (Architecture)
*(추후 시스템 아키텍처 다이어그램이 추가될 예정입니다.)*

## 💡 트러블슈팅 및 기술적 의사결정 (ADR)
> 개발 과정에서 마주친 깊이 있는 고민과 해결 과정을 기록합니다.
- [Go 언어를 메인 서버 기술로 선택한 이유](Issue/Wiki 링크 예정)
- [SSE vs WebSocket: AI 스트리밍 통신에 SSE를 채택한 배경](Issue/Wiki 링크 예정)
- (진행하며 마주치는 트러블슈팅 링크 추가)

## ⚙️ 로컬 실행 방법 (How to Run)
```bash
# 레포지토리 클론
$ git clone [https://github.com/username/ai-live-interview.git](https://github.com/username/ai-live-interview.git)

# 프로젝트 폴더 이동
$ cd ai-live-interview

# Go 모듈 다운로드
$ go mod tidy

# 서버 실행
$ go run main.go
