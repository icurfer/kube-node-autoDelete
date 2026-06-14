// controllers/gpu_taint.go
package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	// gpuPresentLabel: GPU 노드 식별 라벨 (NVIDIA GPU Feature Discovery 가 부여)
	gpuPresentLabel = "nvidia.com/gpu.present"
	// gpuResourceName: device plugin 이 status.allocatable 에 등록하는 리소스명
	gpuResourceName = corev1.ResourceName("nvidia.com/gpu")

	// 복구 대상 taint
	gpuTaintKey    = "nvidia.com/gpu"
	gpuTaintValue  = "present"
	gpuTaintEffect = corev1.TaintEffectNoSchedule
)

// reconcileGPUTaint 는 Ready=True 인 GPU 노드에 NoSchedule taint 가 유실된 경우 복구한다.
// device plugin 이 allocatable 에 GPU 를 등록한 이후에만 taint 를 부여해
// GPU operator 초기화 완료 전 일반 워크로드가 스케줄링되는 오염을 방지한다.
func (r *NodeReconciler) reconcileGPUTaint(ctx context.Context, node *corev1.Node) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 기능 off
	if !r.EnableGPUTaint {
		return ctrl.Result{}, nil
	}

	// GPU 노드가 아니면 대상 아님
	if !hasGPUPresentLabel(node) {
		return ctrl.Result{}, nil
	}

	// GPU 리소스가 아직 allocatable 에 등록되지 않음 → operator 초기화 대기.
	// 별도 폴링 없이 종료한다 — device plugin 이 allocatable 을 등록하는 순간
	// Update 이벤트가 발생하고, predicate(allocatable 변경)가 이를 통과시켜
	// reconcile 이 다시 깨어난다. 고정 주기 requeue 는 불필요하며, allocatable 이
	// 영영 안 붙는 노드(operator 미설치 등)를 무한 폴링하는 부작용만 만든다.
	if !gpuAllocatable(node) {
		logger.Info("GPU node not ready yet (allocatable empty), waiting for device plugin", "node", node.Name)
		return ctrl.Result{}, nil
	}

	// 이미 taint 가 있으면 멱등 종료
	if hasGPUTaint(node) {
		return ctrl.Result{}, nil
	}

	// taint 복구. Spec.Taints 는 merge key 없는 atomic 리스트라 일반 MergeFrom 은
	// 전체 배열을 덮어써 동시 taint 쓰기를 유실시킨다. OptimisticLock 으로
	// resourceVersion 을 실어 충돌 시 Conflict 를 받아 reconcile 이 최신 상태로
	// 재시도하도록 한다.
	patch := client.MergeFromWithOptions(node.DeepCopy(), client.MergeFromWithOptimisticLock{})
	node.Spec.Taints = append(node.Spec.Taints, buildGPUTaint())
	if err := r.Patch(ctx, node, patch); err != nil {
		return ctrl.Result{}, err
	}
	logger.Info("GPU taint restored",
		"node", node.Name,
		"taint", gpuTaintKey+"="+gpuTaintValue+":"+string(gpuTaintEffect),
	)

	return ctrl.Result{}, nil
}

// hasGPUPresentLabel 은 노드가 GPU 식별 라벨을 가졌는지 확인한다.
func hasGPUPresentLabel(node *corev1.Node) bool {
	return node.Labels[gpuPresentLabel] == "true"
}

// gpuAllocatable 은 status.allocatable 에 nvidia.com/gpu 가 0 보다 크게 등록됐는지 확인한다.
// device plugin 이 정상 기동해 GPU 를 노출했음을 의미한다.
func gpuAllocatable(node *corev1.Node) bool {
	q, ok := node.Status.Allocatable[gpuResourceName]
	return ok && !q.IsZero()
}

// hasGPUTaint 은 복구 대상 taint 가 이미 부여돼 있는지 확인한다.
func hasGPUTaint(node *corev1.Node) bool {
	for i := range node.Spec.Taints {
		t := &node.Spec.Taints[i]
		if t.Key == gpuTaintKey && t.Value == gpuTaintValue && t.Effect == gpuTaintEffect {
			return true
		}
	}
	return false
}

// buildGPUTaint 은 복구할 taint 객체를 만든다.
func buildGPUTaint() corev1.Taint {
	return corev1.Taint{
		Key:    gpuTaintKey,
		Value:  gpuTaintValue,
		Effect: gpuTaintEffect,
	}
}
