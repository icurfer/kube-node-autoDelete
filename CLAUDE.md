# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 프로젝트 컨텍스트

- **1인 운영 프로젝트**. 협업자 다수 가정 안 함 — 결정·승인 절차는 가벼움.
- **원본 저장소는 GitHub**: `github.com/icurfer/kube-node-autoDelete`.
- 로컬 작업 경로 이름은 원격 저장소 위치를 뜻하지 않는다.
- 이미지: `ghcr.io/icurfer/kube-node-autodelete`. 버전 태그로 게시하며 클러스터 배포는 수동 Helm 작업이다.

## 프로젝트 규칙 (반드시 먼저 확인)

- 협업 스타일 (한국어 / 간결 / 짧은 응답 = 승인): `.claude/rules/communication.md`
- 커밋 시점, version 파일 형식, 이미지 태그 정책: `.claude/rules/git-workflow.md`
- 빌드/런타임/K8s 호환 범위: `.claude/rules/platform.md`
- CI 책임 경계 (build/push 전용) + 로컬 검증 묶음: `.claude/rules/local-verify.md`
- 코드 변경 시 호출할 서브에이전트: `.claude/rules/subagents.md`
- 블로그 초안 관리: `.claude/rules/blog-drafts.md`
- version + helm/values.yaml 동기화 절차: `.claude/skills/version-management/SKILL.md`
- 슬래시 커맨드: `.claude/commands/bump-version.md`

## Project

`kube-node-autodelete` 는 전원이 꺼졌거나 kubelet 단절로 `NodeReady=Unknown` 이 된 Kubernetes 노드를 자동으로 cordon → drain → delete 처리하는 controller-runtime 기반 컨트롤러다. Kubebuilder v4 layout, Go 1.22, controller-runtime v0.17, k8s.io/* v0.29.

## 자주 쓰는 명령

```bash
# 빌드 (정적 바이너리)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o manager ./cmd

# Lint
golangci-lint run

# 컨테이너 이미지 빌드
docker build -t ghcr.io/icurfer/kube-node-autodelete:<tag> .

# Helm 차트 검증 / 설치
helm lint ./helm
helm template kna ./helm --namespace kube-node-autodelete-system
helm install kube-node-autodelete ./helm -n kube-node-autodelete-system --create-namespace
```

GPU taint helper 단위 테스트가 있다. 통합 테스트 추가 시 envtest(controller-runtime)를 사용한다. 단일 테스트 실행은 `go test ./controllers -run TestName -count=1`.

## 아키텍처 핵심

**진입점**: `cmd/main.go` — controller-runtime Manager 생성 후 `NodeReconciler` 등록. 플래그: `--leader-elect`(기본 true), `--metrics-bind-address=:8080`, `--health-probe-bind-address=:8081`. LeaderElectionID는 코드에 하드코딩(`kube-node-autodelete.example.org`).

**Reconcile 트리거 (`controllers/node_predicates.go`)**: 모든 Node 이벤트가 아니라 `Ready` condition 의 status / lastTransitionTime 변경만 통과시킴. UpdateFunc 외 다른 이벤트(Create/Delete)는 기본값으로 처리되며, 이 필터를 변경하면 reconcile 빈도가 폭증할 수 있다.

**Reconcile 흐름 (`controllers/node_controller.go`)**:
1. control-plane / master 라벨이 있으면 즉시 종료 (`isControlPlaneNode`)
2. `Ready != Unknown` 또는 `Reason != "NodeStatusUnknown"` 이면 종료 — 단순 `NotReady`(네트워크 지연 등)는 대상 아님
3. `lastTransitionTime` 이후 `notReadyGrace`(상수, `2 * time.Minute`) 미만이면 1분 후 requeue
4. cordon → drain → delete. drain 미완 시 30초 후 requeue
5. `notReadyGrace` 는 컴파일타임 상수다. 운영 grace 기간 변경은 코드 수정 + 재빌드 필요 (helm values 로 노출되어 있지 않음)

**Drain 의미론 (`controllers/drain.go`, `controllers/pod_utils.go`)**:
- Pod 조회는 `spec.nodeName` field index 기반 — `SetupWithManager` 에서 인덱서 등록
- Eviction API(`policy/v1`) 사용. PDB 위반 시 individual eviction 실패는 발생하지만 컨트롤러가 errorscheme 으로 다시 reconcile 함
- 제외 대상: 이미 terminating, DaemonSet 소속, mirror pod(static), `kube-system` namespace 전체. 이 기준을 바꾸려면 `filterEvictablePods` 수정

**상태 저장 없음**: 컨트롤러는 어떤 annotation 도 쓰지 않는다. 따라서 컨트롤러 재시작 직후 이미 `NodeStatusUnknown` 상태로 grace 기간을 넘긴 노드는 즉시 삭제 후보가 된다 — README 의 "주의 사항" 참고. annotation 기반 first-seen 로직은 미구현.

## 배포 (helm)

차트 루트는 `helm/`. kustomize 기반 `config/` 는 제거됨.

- 템플릿: `Namespace`, `ServiceAccount`, `ClusterRole`/`ClusterRoleBinding`(node/pod/eviction 권한), leader-election용 `Role`/`RoleBinding`(configmaps/leases/events), `Deployment`
- 주요 values: `image.{repository,tag}`, `replicaCount`, `namespace.{create,name}`, `rbac.create`, `leaderElection.enabled`, `args.{metricsBindAddress,healthProbeBindAddress,extra}`. 전체 항목은 `helm/values.yaml`
- `app.kubernetes.io/component` 라벨은 helpers에서 제외하고 각 템플릿에서 명시적으로 부여한다(controller / leader-election 구분). 헬퍼에 다시 넣으면 leader-election 리소스에서 라벨이 중복됨

## RBAC 변경 시 주의

컨트롤러가 추가로 호출하는 API 가 생기면 `helm/templates/clusterrole.yaml` 의 rules 와 `controllers/` 코드가 동시에 수정돼야 한다. 현재 권한:
- `nodes`: get/list/watch/patch/delete
- `pods`: get/list/watch
- `pods/eviction` (core group, subresource): create
- (leader-election) `configmaps`, `coordination.k8s.io/leases`, `events`: 전체 R/W
