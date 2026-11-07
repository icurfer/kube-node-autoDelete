---
name: git-workflow
description: 커밋 시점 및 version 파일 관리·증가 규칙
---

# Git 워크플로

## 커밋 시점

- 한 가지 코드 작업이 **완료**되면 그 시점에 커밋한다 (= 컴파일 통과 + 의미 있는 단위 종료).
- 미완성·컴파일 실패 상태로 커밋하지 않는다.
- 커밋 메시지 prefix 권장: `feat:` / `fix:` / `refactor:` / `chore:` / `docs:` / `ci:` / `test:`.
- CI 자동 커밋은 `ci: ... [skip ci]` prefix 를 그대로 둔다.

## version 파일

- 저장소 루트에 `version` 파일을 둔다.
- **1번줄에만 버전 값**이 있어야 하며, 그 뒤에 빈 줄/추가 줄이 있어선 안 된다.
- **끝 개행 없음** (POSIX 줄바꿈도 추가하지 않음).
- 형식: `MAJOR.MINOR.PATCH` — **시작 값은 `0.0.0`**, 현재 값은 `version` 파일 참조.
- 이미지 태그 = `version` 파일 내용 그대로. 즉 `0.0.4` 면 이미지 태그도 `0.0.4`.
- `helm/values.yaml` 의 `image.tag` 와 항상 동기화. 로컬 bump 시 함께 변경하면 CI 검증을 통과한다.

## 버전 단계별 의미

### `0.x.y` (안정화 전, 현재 단계)

| 자리 | 언제 올리나 |
|---|---|
| PATCH | 버그 수정·리팩터·작은 동작 개선 — 외부 동작 변화 거의 없음 |
| MINOR | **새 기능 단위 완료**. 예: DRY_RUN 모드, label 기반 노드 제외, Prometheus metrics, annotation 기반 first-seen 로직 등. |
| MAJOR | `0 → 1` 전환은 **첫 안정 릴리스 시점에만** (`1.0.0`). |

### `1.x.y` (안정 릴리스 이후, SemVer)

| 자리 | 기준 |
|---|---|
| PATCH | 동작 호환성 유지 버그 수정 |
| MINOR | 동작 호환성 유지 기능 추가 (RBAC 권한 추가 포함 — 단, 기존 기능 호환 시) |
| MAJOR | 동작 호환성 깨짐: 기본 grace 기간 변경, RBAC 권한 제거, 노드 삭제 기준 변경, CRD 도입 등 운영자가 인지해야 하는 변경 |

## push 시 자동/수동 처리

- 코드·이미지 변경을 릴리스할 때 PATCH를 올린다. 문서만 바뀌면 올리지 않는다.
- `version`, Helm `image.tag`, 차트 `appVersion`을 맞춘 뒤 동일한 `vX.Y.Z` 태그를 push한다.
- **MINOR/MAJOR 는 자동화에 맡기지 말 것.** 기능 단위 완료(MINOR), 안정 전환(MAJOR)은 사람이 판단해 별도 커밋으로 직접 올린다.

상세 절차는 `.claude/skills/version-management/SKILL.md` 참고.
