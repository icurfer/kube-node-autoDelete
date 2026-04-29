---
name: blog-drafts
description: blog/ 디렉토리는 .gitignore 등재. mermaid 환영, 1인 프로젝트라 sanitize 불필요
---

# 블로그 초안 관리

블로그 글은 저장소 메타가 아닌 사용자 개인 작업이다.

## 위치 / 형식

- 경로: `blog/YYYY-MM-DD-<slug>.md`.
- frontmatter 포함: `title`, `date`, `tags`.
- `.gitignore` 의 `blog/` 패턴으로 git 트래킹되지 않음. 외부 플랫폼(velog/티스토리/Hashnode 등)으로 옮겨 게시.

## 다이어그램

- mermaid 다이어그램 활용 권장 — 구조도 / Reconcile 흐름 / 배포 파이프라인 등 시각화 시 글 가치 ↑.
- 코드 블록(```` ```mermaid ```` ) 은 GitHub / Gitea / Obsidian / VSCode(`bierner.markdown-mermaid` 확장) 모두에서 렌더됨.

## sanitize 정책

1인 프로젝트라 다음은 **sanitize 하지 않아도 무방** (사용자 명시적 결정):

- 내부 호스트: `example.com`, `example.com/kube-tools/...`
- 노드 이름, Pod 이름, 네임스페이스
- 개인 이메일 (어차피 git 커밋 author 로 노출됨)

호스트명 노출 ≠ 접근 권한 부여. Harbor/Gitea 자체 인증이 게이트키퍼이므로 별도 위장 불필요.
