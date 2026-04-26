---
name: subagents
description: 작업 단계별로 반드시 호출해야 하는 서브에이전트 규칙
---

# 서브에이전트 사용 규칙

다음 상황에서는 해당 서브에이전트를 **반드시** 호출한다.

| 상황 | 호출할 에이전트 |
|---|---|
| 코드 수정 완료 후 | `code-reviewer` (셀프 리뷰) |
| RBAC 변경, 새 K8s API 호출 추가, eviction/finalizer 관련 변경 | `security-auditor` |
| 버그 수정 시 | (1) `debugger` 로 원인 분석 → (2) 수정 → (3) `test-writer` 로 회귀 테스트 추가 |
| 새 컨트롤러/리컨실러 추가 시 | `code-reviewer` + `test-writer` |

## 호출 안 해도 되는 경우

- 오타 수정
- 주석만 추가/수정
- 포맷팅만 변경 (의미 변경 없음)
- 문서(`docs/`, `*.md`)만 변경
- `version` 파일 단독 bump

## 에이전트 가용성

- `code-reviewer` — Claude Code 공식 플러그인 (`pr-review-toolkit`, `feature-dev`)에 포함. 보통 사용 가능.
- `security-auditor`, `test-writer`, `debugger` — 외부 에이전트. 시스템에 설치되어 있지 않으면 사용자에게 설치 또는 대체 절차를 보고할 것.

호출 시도 후 에이전트 미설치 오류가 나면, 작업을 중단하지 말고 사용자에게 상황을 보고하고 수동 점검 또는 대체 에이전트 사용을 제안한다.
