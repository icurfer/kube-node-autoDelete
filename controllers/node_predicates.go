// controllers/node_predicates.go
package controllers

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func NodePredicate(enableGPUTaint bool) predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode := e.ObjectOld.(*corev1.Node)
			newNode := e.ObjectNew.(*corev1.Node)

			// GPU taint 복구 트리거: 기능이 켜졌고 GPU 노드일 때만 검사한다.
			// allocatable 등록 / taint 유실 순간을 잡되, label 전체 변경은
			// 통과시키지 않는다 — GFD 가 타임스탬프 라벨을 자주 갱신해
			// reconcile 폭증을 유발하므로 정밀 감지만 한다. 기능이 꺼졌거나
			// 일반 노드면 taint 슬라이스 스캔 자체를 건너뛴다.
			if enableGPUTaint && (hasGPUPresentLabel(oldNode) || hasGPUPresentLabel(newNode)) {
				if gpuAllocatable(oldNode) != gpuAllocatable(newNode) ||
					hasGPUTaint(oldNode) != hasGPUTaint(newNode) {
					return true
				}
			}

			oldCond := getReadyCondition(oldNode)
			newCond := getReadyCondition(newNode)

			if oldCond == nil || newCond == nil {
				return true
			}

			return oldCond.Status != newCond.Status ||
				!oldCond.LastTransitionTime.Equal(&newCond.LastTransitionTime)
		},
	}
}
