// controllers/logger.go
package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// logNodeSummary is retained for manual diagnostics; normal reconciliation does not call it.
//
//nolint:unused
func logNodeSummary(ctx context.Context, c client.Client) {
	logger := log.FromContext(ctx)

	var nodes corev1.NodeList
	if err := c.List(ctx, &nodes); err != nil {
		logger.Error(err, "Failed to list nodes")
		return
	}

	logger.Info("Current node status summary")

	for _, n := range nodes.Items {
		cond := getReadyCondition(&n)
		status := "Unknown"
		if cond != nil {
			status = string(cond.Status)
		}

		logger.Info("Node",
			"name", n.Name,
			"status", status,
			"role", nodeRole(&n),
		)
	}
}
