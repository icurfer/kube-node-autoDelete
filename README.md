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

## RBAC 및 Controller 배포
```bash
cd ~/config/default
```
```bash
kubectl apply -k .
```

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