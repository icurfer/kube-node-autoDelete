---
name: local-verify
description: CI 는 build/push 전용. 코드 검증은 로컬에서 처리하는 정책
---

# 로컬 검증 정책

## 책임 분리

- **CI (`.gitea/workflows/build.yaml`)**: 이미지 build/push, `helm/values.yaml` 태그 자동 갱신.
- **로컬**: 코드/설정 검증 (lint, compile, helm chart 문법). CI 는 검증을 수행하지 않는다.

이 분리는 CI 러너 비용·시간 절약과 푸시 전 빠른 피드백을 위함이다.

## 권장 verify 묶음

push 전에 다음을 모두 통과해야 한다.

```bash
go build ./...
golangci-lint run
helm lint ./helm
```

- `go build ./...` — 컴파일 가능 여부.
- `golangci-lint run` — `.golangci.yml` 정의된 linter 묶음 (errcheck, govet, staticcheck, ineffassign 등).
- `helm lint ./helm` — 차트 문법, 필수 필드, 템플릿 렌더 가능 여부.

테스트 코드가 추가되면 `go test ./...` 도 묶음에 포함시킨다.

## 통과 후 push

- 위 3개 통과 → `/bump-version` (또는 수동 `version` 갱신) → 커밋 → push.
- 깨진 코드는 push 하지 않는다 — CI 가 잡아주지 않는다.

## 자동화 옵션

현재 도입되어 있지 않음. 필요해지면 다음 중 선택:

- **Makefile**: `make verify` 한 줄로 묶음 실행 (수동, 명시적).
- **pre-push hook**: 푸시 시 자동 실행, 우회는 `--no-verify`.
- **pre-commit**: 커밋 시 자동, 더 빈번. 가벼운 항목만 권장.
