# kube_node_autoDelete
> kube-node-autodelete는 전원이 꺼졌거나 kubelet이 완전히 단절된 Kubernetes 노드를 자동으로 감지하여
안전하게 cordon → drain → delete 처리하는 컨트롤러입니다. ChatGPT를 이용하여 구현되었습니다.

> control-plane(master) 노드는 자동 삭제 대상에서 제외됩니다.
## 개발 목적
- kubeadm으로 구성된 kubernetes 테스트환경에서 Node를 drain없이 shutdown하는 경우가 있는데 Node는 NotReady상태인데,
pod는 Running으로 동작해서 발생하는 이상현상해소를 위함.

## 주요 기능
- NodeReady=Unknown 상태 노드 감지
- configurable grace period 후 자동 처리
- control-plane / master 노드 자동 제외
- Pod eviction 기반 안전한 drain
- Controller 재시작에도 중복 처리 방지 가능 (annotation 확장 가능)
    - 확인 필요.

### 삭제 대상 기준

> 다음 조건을 모두 만족해야 노드가 삭제됩니다:

- NodeReady 상태가 Unknown
- Reason == NodeStatusUnknown
- (전원 OFF / kubelet 완전 단절 상황)
- control-plane / master 노드 아님
- 단순 NotReady (네트워크 지연, 일시적 장애)는 삭제하지 않음
    - 확인 필요. 디테일하고 코드 검토하지 않음.

### 동작 흐름
```
Node 상태 변경
   ↓
NodeReady == Unknown
   ↓
(grace period 대기)
   ↓
cordon
   ↓
drain (eviction)
   ↓
node delete
```

### 주요 구성 요소
controllers/
├── node_controller.go     # 메인 Reconciler
├── drain.go               # Pod eviction 기반 drain 로직
├── node_predicates.go     # Node 상태 변경 필터
├── node_utils.go          # Ready condition / role 유틸
├── node_logger.go         # Node 상태 요약 로그

### 설정 값
- grace period
- const notReadyGrace = 2 * time.Minute
- 운영 환경에 따라 5~10분 권장

### 제외되는 Pod 유형
- 확인필요.

### 제외되는 노드
- 다음 label이 있는 노드는 절대 삭제되지 않습니다. > 확인필요.
    - node-role.kubernetes.io/control-plane
    - node-role.kubernetes.io/master

## RBAC 및 Controller 배포 (Helm)
> Helm 3 기준. 차트는 `helm/` 디렉토리에 포함되어 있습니다.

```bash
# 차트 렌더 확인 (선택)
helm template kube-node-autodelete ./helm --namespace kube-node-autodelete-system

# 설치 (--create-namespace 필수: 차트가 NS 매니페스트를 갖지 않음)
helm install kube-node-autodelete ./helm \
  --namespace kube-node-autodelete-system \
  --create-namespace

# 업그레이드
helm upgrade kube-node-autodelete ./helm \
  --namespace kube-node-autodelete-system

# 제거 (네임스페이스는 남음 — helm 이 만든 게 아니라 helm CLI 가 만든 것)
helm uninstall kube-node-autodelete \
  --namespace kube-node-autodelete-system

# 네임스페이스까지 정리하려면
kubectl delete ns kube-node-autodelete-system
```

> 운영상 주의: `helm uninstall` 은 컨트롤러를 즉시 종료시키므로 진행 중이던 drain/eviction 이 중단되고
> cordon 된 노드가 그대로 남을 수 있다. 운영 노드 정리 작업이 진행 중이 아닐 때 수행 권장.

주요 values (전체 항목은 `helm/values.yaml` 참고):

| key | 기본값 | 설명 |
| --- | --- | --- |
| `image.repository` | `ghcr.io/icurfer/kube-node-autodelete` | 컨트롤러 이미지 |
| `image.tag` | `0.1.1` | 이미지 태그 (루트 `version` 파일과 일치해야 함) |
| `replicaCount` | `1` | Deployment replica 수 |
| `namespace.create` | `false` | 차트가 namespace 매니페스트를 포함할지 여부 (기본 false — `helm install --create-namespace` 사용 권장) |
| `namespace.name` | `kube-node-autodelete-system` | 배포 namespace |
| `rbac.create` | `true` | ClusterRole / RoleBinding 생성 여부 |
| `leaderElection.enabled` | `false` | leader election 사용 여부 |
| `args.metricsBindAddress` | `:8080` | metrics 바인딩 주소 |
| `args.healthProbeBindAddress` | `:8081` | health probe 바인딩 주소 |

### 로그 예시

> 정상 동작 시 다음과 같은 로그를 확인할 수 있습니다.

- Node NotReady too long → drain & delete
- Node cordoned
- Drain complete: no evictable pods left
- Deleting node

## 주의 사항

컨트롤러 배포 이전에 이미 NotReady 상태인 노드는
grace period를 초과한 것으로 판단되어 즉시 삭제될 수 있습니다
이를 방지하려면:
annotation 기반 first-seen 로직 추가를 권장합니다



⏱ annotation 기반 grace timer

🧪 DRY_RUN 모드

🏷 특정 label이 있는 노드는 삭제 금지

📊 Prometheus metrics export

📦 Helm Chart 제공


## Author
- icurfer
- Kubernetes / Cloud Native Engineer

## GitHub CI와 이미지 배포

원본 저장소: https://github.com/icurfer/kube-node-autoDelete

- PR과 `main` push: Go 테스트·빌드·lint, Helm 검증, 컨테이너 빌드.
- `vX.Y.Z` 태그 push: 같은 검증 후 `ghcr.io/icurfer/kube-node-autodelete:X.Y.Z` 게시.
- `version`, Helm `image.tag`, 차트 `appVersion`은 같은 값이어야 한다.
- 이미지에는 소스 저장소 URL, 버전, 커밋 SHA를 OCI label로 기록한다.
- GHCR 패키지는 Public으로 설정해야 클러스터에서 인증 없이 pull할 수 있다.
- GitHub Actions에는 클러스터 인증정보를 등록하지 않는다. 배포는 수동으로 수행한다.

```bash
helm upgrade --install kube-node-autodelete ./helm \
  --namespace kube-node-autodelete-system --create-namespace \
  --set image.repository=ghcr.io/icurfer/kube-node-autodelete \
  --set-string image.tag=0.1.1
kubectl -n kube-node-autodelete-system rollout status deployment/kube-node-autodelete
```

배포 전에 `helm history`와 기존 values를 보관한다. 문제가 있으면 확인해 둔 revision으로
`helm rollback kube-node-autodelete <revision> -n kube-node-autodelete-system`을 실행한다.

문서에 개인 인프라 도메인이나 인증정보를 넣지 않는다. 예시 주소는 `example.com` 또는
`example.org`를 사용한다. 과거 검토 문서의 버전·운영 상태는 당시 기록이다.
