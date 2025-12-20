package controllers

import corev1 "k8s.io/api/core/v1"

func filterEvictablePods(pods []corev1.Pod) []corev1.Pod {
	result := []corev1.Pod{}

	for _, pod := range pods {

		// 이미 terminating
		if pod.DeletionTimestamp != nil {
			continue
		}

		// DaemonSet
		if isDaemonSetPod(&pod) {
			continue
		}

		// mirror pod (static pod)
		if isMirrorPod(&pod) {
			continue
		}

		// kube-system static / control-plane 보호
		if pod.Namespace == "kube-system" {
			continue
		}

		result = append(result, pod)
	}

	return result
}

func isDaemonSetPod(pod *corev1.Pod) bool {
	for _, owner := range pod.OwnerReferences {
		if owner.Kind == "DaemonSet" {
			return true
		}
	}
	return false
}

func isMirrorPod(pod *corev1.Pod) bool {
	_, ok := pod.Annotations["kubernetes.io/config.mirror"]
	return ok
}
