---
name: version-management
description: kube-node-autodelete 저장소의 version 파일과 helm/values.yaml image.tag 를 형식 계약에 맞게 갱신·동기화한다. 버전 증가가 필요한 모든 작업에서 호출.
---

# version 관리 스킬

## 세 가지 버전의 분리

이 저장소에는 **독립된 cadence 로 움직이는 세 가지 버전** 이 있다. 절대 서로 자동 동기화시키지 않는다.

| 대상 | 파일 / 필드 | 무엇을 가리키는가 | 언제 bump 하는가 |
|---|---|---|---|
| 컨트롤러 이미지 태그 | `version` (루트) ↔ `helm/values.yaml` `image.tag` | 빌드된 컨테이너 이미지 식별자 | 코드 변경 시 (`/bump-version`) |
| Helm 차트 버전 | `helm/Chart.yaml` `version` | 차트 자체의 버전 (templates, values 기본값, RBAC) | 차트 구조 변경 시 수동 |
| Helm 차트 appVersion | `helm/Chart.yaml` `appVersion` | 이 차트가 동봉·검증한 "기준 앱 버전" 메타 | 차트-앱 계약이 의미 있게 변할 때만 수동, 매 이미지 bump 마다 동기화하지 않음 |

이 분리의 이유:
- 코드만 변해도 차트는 그대로일 수 있다 → 차트 사용자가 매 이미지 bump 마다 새 차트 버전을 받지 않게 함.
- 차트만 변해도 이미지는 그대로일 수 있다 (예: RBAC 수정, replica 기본값 변경) → 이미지 재빌드 없이 차트 버전만 올린다.

## 파일 형식 계약

### `version` (루트)

- 1번줄에 `MAJOR.MINOR.PATCH` 한 줄.
- 그 외 줄 (빈 줄 포함) 금지.
- **끝 개행 없음** (POSIX 기본 줄바꿈도 추가하지 않음).

### `helm/values.yaml`

- `image.tag` 필드의 값이 `version` 파일과 **완전히 동일**.
- 따옴표 없이 단순 스칼라 (`tag: 0.0.4`).

### `helm/Chart.yaml`

- `version`: 차트 SemVer. 템플릿/values 기본값/RBAC/Chart 메타 변경 시 bump.
- `appVersion`: 차트와 결합 보장된 앱 버전 메타. 매 이미지 bump 와 무관하게 별도 cadence.

### 이미지 태그 (CI 산출물)

- `example.com/open/kube-node-autodelete:<version>` — `<version>` 은 `version` 파일 그대로.
- prefix(`dev_`, `release_` 등) 붙이지 않는다 — version 과 1:1 매칭이 추적·롤백에 가장 명확.

## 시작 값

- 최초 `version` 파일은 `0.0.0`.
- 최초 `helm/values.yaml` 의 `image.tag` 도 `0.0.0`.

## 증가 규칙 — 0.x.y 단계 (안정화 전, 현재 단계)

| 변경 종류 | MAJOR | MINOR | PATCH |
|---|---|---|---|
| 매 push (기본) | - | - | +1 |
| 새 기능 단위 완료 (DRY_RUN, 라벨 제외, metrics 등) | - | +1 | 0 |
| 첫 안정 릴리스 | 0→1 | 0 | 0 |

## 증가 규칙 — 1.x.y 단계 (안정 릴리스 이후, SemVer)

| 변경 종류 | MAJOR | MINOR | PATCH |
|---|---|---|---|
| 동작 호환성 깨짐 (grace 기본값 변경, RBAC 권한 제거 등) | +1 | 0 | 0 |
| 호환 유지 기능 추가 | - | +1 | 0 |
| 호환 유지 버그 수정 | - | - | +1 |
| 매 push (기본) | - | - | +1 |

자동화(`/bump-version`)는 **PATCH 만** 자동 처리. MINOR/MAJOR 는 사람이 직접 결정.

## 절차

1. `version` 읽기 → 한 줄 보장 확인.
2. 새 버전 계산.
3. `version` 갱신 (한 줄만, 끝 개행 없이).
4. `helm/values.yaml` 의 `image.tag` 갱신 (값만 새 버전으로).
5. 두 파일을 stage 후 단독 커밋.

## CI 와의 관계

- `.gitea/workflows/build.yaml` 은 `version` 변경을 트리거로 이미지 build/push 후 `helm/values.yaml` 의 `image.tag` 를 sed 로 갱신해 `[skip ci]` 자동 커밋한다.
- 로컬 `/bump-version` 이 `helm/values.yaml` 을 함께 갱신해두면 CI 의 sed 결과가 동일해 자동 커밋이 발생하지 않는다 (정상). 한쪽만 갱신해도 결과적으로 동기화는 보장되지만 커밋이 1개 더 생긴다.

## 자기검증 체크리스트

- [ ] `wc -c version` 결과 = `MAJOR.MINOR.PATCH` 길이 (한 줄, 끝 개행 없음).
- [ ] `version` 파일에 빈 줄·공백 줄 없음.
- [ ] `helm/values.yaml` 의 `image.tag` 가 `version` 파일과 동일.
- [ ] `helm template ./helm | grep "image:"` 결과의 태그가 새 버전.
