// controllers/node_controller.go
package controllers

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const notReadyGrace = 2 * time.Minute

type NodeReconciler struct {
	client.Client
	// EnableGPUTaint: GPU 노드 재등록 시 NoSchedule taint 자동 복구 활성화
	EnableGPUTaint bool
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// 🔴 Pod spec.nodeName index (drain 필수)
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&corev1.Pod{},
		"spec.nodeName",
		func(rawObj client.Object) []string {
			pod := rawObj.(*corev1.Pod)
			if pod.Spec.NodeName == "" {
				return nil
			}
			return []string{pod.Spec.NodeName}
		},
	); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		WithEventFilter(NodePredicate(r.EnableGPUTaint)).
		Complete(r)
}

func (r *NodeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var node corev1.Node
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name}, &node); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 🚫 control-plane 제외
	if isControlPlaneNode(&node) {
		logger.Info("Skip control-plane node", "node", node.Name)
		return ctrl.Result{}, nil
	}

	cond := getReadyCondition(&node)
	if cond == nil {
		return ctrl.Result{}, nil
	}

	// ✅ Ready 노드 → GPU taint 복구 책임 (삭제 로직과 무관)
	if cond.Status == corev1.ConditionTrue {
		return r.reconcileGPUTaint(ctx, &node)
	}

	// 🔒 전원 OFF 상태만
	if cond.Reason != "NodeStatusUnknown" {
		return ctrl.Result{}, nil
	}

	// ⏱ grace time
	if time.Since(cond.LastTransitionTime.Time) < notReadyGrace {
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, nil
	}

	logger.Info("Node NotReady too long → drain & delete",
		"node", node.Name,
		"since", cond.LastTransitionTime.Time,
	)

	// 1️⃣ cordon
	if !node.Spec.Unschedulable {
		patch := client.MergeFrom(node.DeepCopy())
		node.Spec.Unschedulable = true
		if err := r.Patch(ctx, &node, patch); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("Node cordoned", "node", node.Name)
	}

	// 2️⃣ drain
	done, err := r.DrainNode(ctx, &node)
	if err != nil {
		return ctrl.Result{}, err
	}
	if !done {
		logger.Info("Drain in progress", "node", node.Name)
		return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
	}

	// 3️⃣ delete node
	logger.Info("Deleting node", "node", node.Name)
	if err := r.Delete(ctx, &node); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
