# ic-praxis 도입 검토 — 2026-08-27

대상: `node-auto-del` (`kube-node-autodelete`)
상태: **미적용.** 아래 후처리가 필요해 이번에는 문서만 남긴다.

## 0. TL;DR

- 이 저장소에 ic-praxis 는 **설치되어 있지 않다** (훅·스크립트·`AGENTS.md` 전부 없음).
- 규칙은 `.claude/rules/` 에 6개 문서로 잘 정리돼 있지만 **강제 수단이 없다.**
  사람이 읽고 지켜야 하는 상태다.
- 실제로 그 때문에 오판이 발생했다 (§2).
- 적용은 저장소에 파일을 추가하고 git 훅을 활성화하는 변경이므로, `install.sh` 정독 →
  추가 파일 목록 확인 → 적용 → 기존 규칙 문서와의 중복 정리 순으로 진행한다.

## 1. 현재 상태 (확인 결과)

| 검사 항목 | 결과 |
|---|---|
| `.githooks/` | 없음 |
| `scripts/check-conventions.sh` | 없음 |
| `AGENTS.md` | 없음 (`CLAUDE.md` 만 있음) |
| `.git/hooks/pre-commit` | 없음 (샘플 파일만) |
| `core.hooksPath` | 미설정 |
| 저장소 내 `praxis` 문자열 | 없음 |

ic-praxis 본체 위치: `/data/github-icurfer/ic-praxis`
(origin: `github.com/icurfer/ic-praxis`)

## 2. 도입 동기 — 실제로 겪은 오판

`docs/review-2026-07-31.md` 를 커밋할 때, `version` 파일을 bump 해야 하는지 판단이
엇갈렸다.

- `.claude/rules/git-workflow.md` 에는 **"`git push` 직전에 PATCH +1"** 이라고
  조건 없이 적혀 있다.
- 그래서 문서만 바뀐 커밋에도 bump 가 필요하다고 잘못 안내했다.
- 실제로는 불필요하다. `.gitea/workflows/build.yaml` 이 `paths: ['version', '.gitea/workflows/**']`
  로 걸려 있어서, `version` 을 올리면 **내용이 동일한 이미지를 새 태그로 재빌드·재푸시**한다.

즉 규칙 문서에 "코드가 변경된 경우" 라는 전제가 빠져 있었고, 문서만 읽으면 무조건으로
해석된다. ic-praxis README 가 존재 이유로 첫 번째로 드는 문제가 이것이다.

> 문서는 나쁜 커밋을 막지 못한다. 좋은 커밋이 어떤 모양인지 서술할 뿐이다.

pre-commit 게이트라면 판단이 필요 없다. 코드 파일이 스테이징됐는데 `version` 이 그대로면
차단, 문서만 바뀌었으면 통과 — 기계가 결정한다.

## 3. 기계로 검사 가능한 규칙 (게이트 후보)

`.claude/rules/` 와 `.claude/skills/version-management/` 에 이미 서술된 것 중,
자동 검사가 가능한 항목.

| 규칙 | 출처 | 검사 방법 |
|---|---|---|
| `version` 은 1번줄만, 끝 개행 없음 | `git-workflow.md` | 줄 수 + 마지막 바이트 검사 |
| `version` == `helm/values.yaml` 의 `image.tag` | `git-workflow.md` | 두 값 파싱 후 비교 |
| **코드 변경 시 `version` bump 필수 / 문서만이면 불필요** | §2 의 교훈 | staged 파일 확장자 분기 |
| push 전 `go build ./...` 통과 | `local-verify.md` | pre-push 훅 |
| push 전 `golangci-lint run` 통과 | `local-verify.md` | pre-push 훅 |
| push 전 `helm lint ./helm` 통과 | `local-verify.md` | pre-push 훅 |
| RBAC 변경 시 `clusterrole.yaml` 동반 수정 | `CLAUDE.md` | `controllers/` 변경 + clusterrole 미변경 감지 (**정확도 낮음, 경고 수준 권장**) |

마지막 항목은 오탐이 많을 것이라 차단(block)이 아니라 경고로 두는 편이 낫다고 본다.
검증하지 않은 추정이다.

## 4. 적용 시 필요한 후처리

1. **`install.sh` 정독** — 저장소에 무엇이 추가되는지 파일 목록을 먼저 확정한다.
   현재 미확인. `templates/` 에 `AGENTS.md`, `CLAUDE.md`, `docs/`, `scripts/`,
   `optional/` 이 있는 것만 확인했다.
2. **`CLAUDE.md` 충돌 정리** — ic-praxis 가 자체 `CLAUDE.md` 템플릿을 넣는다.
   이 저장소에는 이미 내용이 있는 `CLAUDE.md` 가 있으므로 덮어쓰기 여부를 확인해야 한다.
   `AGENTS.md` 도 새로 생긴다.
3. **`.claude/rules/` 와의 중복 제거** — 게이트로 강제되는 규칙이 문서에도 남으면
   두 곳을 따로 관리하게 된다. 게이트가 진실 소스가 되고 문서는 이유만 남기는 방향.
4. **`git-workflow.md` 의 "push 직전 PATCH +1" 문장 수정** — §2 의 전제("코드 변경 시")를
   명시한다. 게이트를 넣더라도 문서가 틀린 상태로 남으면 안 된다.
5. **훅 활성화 방식 결정** — `core.hooksPath` 를 쓸지 `.git/hooks/` 에 직접 넣을지.
   전자는 저장소에 커밋되어 공유되지만 1인 프로젝트라 실익은 작다.
6. **우회 경로 확인** — `--no-verify` 로 넘어갈 수 있음을 인지하고, CI 가 검증을 하지
   않는다는 `local-verify.md` 의 책임 분리와 모순되지 않는지 본다.

## 5. 커밋되지 않은 채 남은 변경

- `CLAUDE.md` — 저장소 로컬 경로 이동(§6) 반영 + GitHub 저장소가 폐기본이라는 경고 추가.
  의도적으로 커밋하지 않았다.

## 6. 참고: 저장소 경로 이동 (2026-08-27)

```
/data/github-icurfer/gitea-kube-tools/node-auto-del
  → /data/gitea-kube-tools/node-auto-del
```

origin 이 gitea 인데 상위 폴더 이름이 `github-icurfer` 여서 GitHub 저장소를 현행 소스로
착각하기 쉬웠다. 이름을 origin 과 맞췄다. 같은 파일시스템이라 rename 으로 처리됐고
git 히스토리·remote·워킹트리는 온전하다.

비게 된 `/data/github-icurfer/gitea-kube-tools/` 는 혼란만 유발하므로 삭제했다.
이제 `/data/github-icurfer/` 하위는 전부 실제 GitHub 저장소만 남아, 폴더 이름과 origin 이
일치한다 (`PDFMathTranslate`, `ic-crawl`, `ic-nvlog`, `ic-praxis`, `naver-mcp`).
