---
description: version 파일 PATCH +1, helm/values.yaml 동기화 후 chore 커밋
---

# /bump-version

루트의 `version` 파일과 `helm/values.yaml` 의 `image.tag` 를 한 단계 올린 뒤 커밋한다.

## 동작

1. `version` 파일 읽기 — 1번줄만 사용.
2. `MAJOR.MINOR.PATCH` 파싱 → PATCH +1 (`X.Y.Z` → `X.Y.(Z+1)`).
3. `version` 파일 갱신 — **1번줄에만 새 버전, 빈 줄·끝 개행 없음**.
4. `helm/values.yaml` 의 `image.tag` 를 새 버전으로 갱신.
5. 두 파일을 stage 후 커밋:
   ```
   chore: bump version to X.Y.Z
   ```

## 주의

- MINOR/MAJOR 증가가 필요한 경우 이 명령을 쓰지 말고 직접 수정 후 별도 커밋.
- 변경 사항이 다른 작업과 섞이지 않게 **버전 증가 단독 커밋**으로 분리.
- push 직전 호출이 기본. 호출 후 `git push` → CI 가 이미지 build/push.
- `helm/values.yaml` 을 함께 갱신하므로 CI 의 `update-helm` job 은 변경사항이 없어 자동 커밋이 발생하지 않는다 (의도된 동작).
- 커밋·push 전에 `.claude/rules/local-verify.md` 의 검증 묶음 (`go build`, `golangci-lint`, `helm lint`) 통과를 확인할 것.

상세 형식 계약은 `.claude/skills/version-management/SKILL.md` 참고.
