// controllers/drain.go
package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *NodeReconciler) DrainNode(ctx context.Context, node *corev1.Node) (bool, error) {
	logger := log.FromContext(ctx)

	var podList corev1.PodList
	if err := r.List(
		ctx,
		&podList,
		client.MatchingFields{"spec.nodeName": node.Name},
	); err != nil {
		return false, err
	}

	// ✅ 실제 eviction 대상만 필터링
	evictablePods := filterEvictablePods(podList.Items)

	if len(evictablePods) == 0 {
		logger.Info("Drain complete: no evictable pods left", "node", node.Name)
		return true, nil
	}

	logger.Info("Evictable pods found",
		"node", node.Name,
		"count", len(evictablePods),
	)

	for _, pod := range evictablePods {
		eviction := &policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}

		logger.Info("Evicting pod",
			"pod", pod.Name,
			"namespace", pod.Namespace,
		)

		err := r.SubResource("eviction").Create(ctx, &pod, eviction)
		if err != nil && !errors.IsNotFound(err) {
			return false, err
		}
	}

	return false, nil
}
