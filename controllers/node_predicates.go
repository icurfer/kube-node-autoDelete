// controllers/node_predicates.go
package controllers

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

func NodePredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode := e.ObjectOld.(*corev1.Node)
			newNode := e.ObjectNew.(*corev1.Node)

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
