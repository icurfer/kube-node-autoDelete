---
name: local-verify
description: 로컬 및 GitHub Actions 검증
---

# 검증

- 로컬과 PR/main CI: `go test ./...`, `go build ./...`, `golangci-lint run`, `helm lint ./helm`, Helm 렌더, 버전 일치 확인.
- Go 1.22.12와 golangci-lint 1.59.1을 사용한다. 새 도구 버전은 별도 검증 후 갱신한다.
- `vX.Y.Z` 태그도 검증을 통과한 뒤 GHCR 이미지를 게시한다.
- 이미지 게시에는 `GITHUB_TOKEN`만 사용한다. CI에 클러스터 인증정보를 넣지 않는다.
- 클러스터 배포는 검증된 이미지 버전으로 수동 수행한다.
