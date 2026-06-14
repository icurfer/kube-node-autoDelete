// controllers/gpu_taint_test.go
package controllers

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestHasGPUPresentLabel(t *testing.T) {
	cases := []struct {
		name   string
		labels map[string]string
		want   bool
	}{
		{"present true", map[string]string{gpuPresentLabel: "true"}, true},
		{"present false", map[string]string{gpuPresentLabel: "false"}, false},
		{"missing", map[string]string{"foo": "bar"}, false},
		{"nil labels", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			node := &corev1.Node{}
			node.Labels = c.labels
			if got := hasGPUPresentLabel(node); got != c.want {
				t.Errorf("hasGPUPresentLabel = %v, want %v", got, c.want)
			}
		})
	}
}

func TestGPUAllocatable(t *testing.T) {
	cases := []struct {
		name  string
		alloc corev1.ResourceList
		want  bool
	}{
		{"gpu 2", corev1.ResourceList{gpuResourceName: resource.MustParse("2")}, true},
		{"gpu 1", corev1.ResourceList{gpuResourceName: resource.MustParse("1")}, true},
		{"gpu 0", corev1.ResourceList{gpuResourceName: resource.MustParse("0")}, false},
		{"no gpu", corev1.ResourceList{"cpu": resource.MustParse("4")}, false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			node := &corev1.Node{}
			node.Status.Allocatable = c.alloc
			if got := gpuAllocatable(node); got != c.want {
				t.Errorf("gpuAllocatable = %v, want %v", got, c.want)
			}
		})
	}
}

func TestHasGPUTaint(t *testing.T) {
	match := corev1.Taint{Key: gpuTaintKey, Value: gpuTaintValue, Effect: gpuTaintEffect}
	cases := []struct {
		name   string
		taints []corev1.Taint
		want   bool
	}{
		{"exact match", []corev1.Taint{match}, true},
		{"wrong effect", []corev1.Taint{{Key: gpuTaintKey, Value: gpuTaintValue, Effect: corev1.TaintEffectNoExecute}}, false},
		{"wrong value", []corev1.Taint{{Key: gpuTaintKey, Value: "other", Effect: gpuTaintEffect}}, false},
		{"other taint", []corev1.Taint{{Key: "node.kubernetes.io/unschedulable", Effect: gpuTaintEffect}}, false},
		{"empty", nil, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			node := &corev1.Node{}
			node.Spec.Taints = c.taints
			if got := hasGPUTaint(node); got != c.want {
				t.Errorf("hasGPUTaint = %v, want %v", got, c.want)
			}
		})
	}
}

func TestBuildGPUTaint(t *testing.T) {
	tnt := buildGPUTaint()
	if tnt.Key != gpuTaintKey || tnt.Value != gpuTaintValue || tnt.Effect != gpuTaintEffect {
		t.Errorf("buildGPUTaint = %+v, want key=%s value=%s effect=%s",
			tnt, gpuTaintKey, gpuTaintValue, gpuTaintEffect)
	}
	// 빌드한 taint 는 hasGPUTaint 로 감지돼야 한다 (왕복 검증)
	node := &corev1.Node{}
	node.Spec.Taints = []corev1.Taint{tnt}
	if !hasGPUTaint(node) {
		t.Error("buildGPUTaint() 결과를 hasGPUTaint 가 감지하지 못함")
	}
}
