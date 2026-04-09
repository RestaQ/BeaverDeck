package kube

import (
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func summarizeRBACSubjects(items []rbacv1.Subject) string {
	if len(items) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(items))
	for _, subject := range items {
		if subject.Namespace != "" {
			parts = append(parts, fmt.Sprintf("%s:%s/%s", subject.Kind, subject.Namespace, subject.Name))
			continue
		}
		parts = append(parts, fmt.Sprintf("%s:%s", subject.Kind, subject.Name))
	}
	return strings.Join(parts, ", ")
}

func age(t time.Time) string {
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

func podResourceTotals(pod *corev1.Pod) (int64, int64, int64, int64) {
	var (
		sumReqCPU     int64
		sumLimCPU     int64
		sumReqMem     int64
		sumLimMem     int64
		maxInitReqCPU int64
		maxInitLimCPU int64
		maxInitReqMem int64
		maxInitLimMem int64
	)

	for _, container := range pod.Spec.Containers {
		sumReqCPU += container.Resources.Requests.Cpu().MilliValue()
		sumLimCPU += container.Resources.Limits.Cpu().MilliValue()
		sumReqMem += container.Resources.Requests.Memory().Value()
		sumLimMem += container.Resources.Limits.Memory().Value()
	}
	for _, container := range pod.Spec.InitContainers {
		maxInitReqCPU = max64(maxInitReqCPU, container.Resources.Requests.Cpu().MilliValue())
		maxInitLimCPU = max64(maxInitLimCPU, container.Resources.Limits.Cpu().MilliValue())
		maxInitReqMem = max64(maxInitReqMem, container.Resources.Requests.Memory().Value())
		maxInitLimMem = max64(maxInitLimMem, container.Resources.Limits.Memory().Value())
	}

	return max64(sumReqCPU, maxInitReqCPU), max64(sumLimCPU, maxInitLimCPU), max64(sumReqMem, maxInitReqMem), max64(sumLimMem, maxInitLimMem)
}

func podGPURequestCount(pod *corev1.Pod) int64 {
	var (
		sumContainers int64
		maxInit       int64
	)
	for _, container := range pod.Spec.Containers {
		sumContainers += containerGPUCount(container)
	}
	for _, container := range pod.Spec.InitContainers {
		maxInit = max64(maxInit, containerGPUCount(container))
	}
	return max64(sumContainers, maxInit)
}

func containerGPUCount(container corev1.Container) int64 {
	limitQuantity := container.Resources.Limits[corev1.ResourceName("nvidia.com/gpu")]
	limit := limitQuantity.Value()
	if limit > 0 {
		return limit
	}
	requestQuantity := container.Resources.Requests[corev1.ResourceName("nvidia.com/gpu")]
	return requestQuantity.Value()
}

func podRootContexts(pod *corev1.Pod) []string {
	out := make([]string, 0)
	if pod.Spec.SecurityContext != nil && pod.Spec.SecurityContext.RunAsUser != nil && *pod.Spec.SecurityContext.RunAsUser == 0 {
		out = append(out, "pod securityContext.runAsUser=0")
	}
	allContainers := make([]corev1.Container, 0, len(pod.Spec.InitContainers)+len(pod.Spec.Containers))
	allContainers = append(allContainers, pod.Spec.InitContainers...)
	allContainers = append(allContainers, pod.Spec.Containers...)
	for _, container := range allContainers {
		if container.SecurityContext != nil && container.SecurityContext.RunAsUser != nil && *container.SecurityContext.RunAsUser == 0 {
			out = append(out, fmt.Sprintf("container %s securityContext.runAsUser=0", container.Name))
		}
	}
	sort.Strings(out)
	return out
}

func workloadHasService(workloadLabels map[string]string, services []corev1.Service) bool {
	if len(workloadLabels) == 0 {
		return false
	}
	for _, service := range services {
		if len(service.Spec.Selector) == 0 {
			continue
		}
		if selectorMatchesLabels(service.Spec.Selector, workloadLabels) {
			return true
		}
	}
	return false
}

func selectorMatchesLabels(selector, labels map[string]string) bool {
	for key, value := range selector {
		if labels[key] != value {
			return false
		}
	}
	return true
}

func formatSelector(selector map[string]string) string {
	if len(selector) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(selector))
	for key, value := range selector {
		parts = append(parts, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

func ingressBackendServiceNames(ingress *networkingv1.Ingress) []string {
	if ingress == nil {
		return nil
	}
	out := make([]string, 0)
	seen := make(map[string]struct{})
	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, ok := seen[name]; ok {
			return
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	if ingress.Spec.DefaultBackend != nil && ingress.Spec.DefaultBackend.Service != nil {
		add(ingress.Spec.DefaultBackend.Service.Name)
	}
	for _, rule := range ingress.Spec.Rules {
		if rule.HTTP == nil {
			continue
		}
		for _, path := range rule.HTTP.Paths {
			if path.Backend.Service != nil {
				add(path.Backend.Service.Name)
			}
		}
	}
	sort.Strings(out)
	return out
}

func uniqueStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, item := range in {
		v := strings.TrimSpace(item)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}

func formatBytesIEC(v int64) string {
	const unit = int64(1024)
	if v < unit {
		return fmt.Sprintf("%dB", v)
	}
	div, exp := unit, 0
	for n := v / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(v)/float64(div), "KMGTPE"[exp])
}

func formatMilliUsage(used, total int64) string {
	if total > 0 {
		return fmt.Sprintf("%dm / %dm", used, total)
	}
	if used > 0 {
		return fmt.Sprintf("%dm / -", used)
	}
	return "-"
}

func formatByteUsage(used, total int64) string {
	if total > 0 {
		return fmt.Sprintf("%s / %s", formatBytesIEC(used), formatBytesIEC(total))
	}
	if used > 0 {
		return fmt.Sprintf("%s / -", formatBytesIEC(used))
	}
	return "-"
}

func formatMilliUsageUnknown(hasUsage bool, used, total int64) string {
	if total > 0 {
		if hasUsage {
			return fmt.Sprintf("%dm / %dm", used, total)
		}
		return fmt.Sprintf("- / %dm", total)
	}
	if hasUsage && used > 0 {
		return fmt.Sprintf("%dm / -", used)
	}
	return "-"
}

func formatByteUsageUnknown(hasUsage bool, used, total int64) string {
	if total > 0 {
		if hasUsage {
			return fmt.Sprintf("%s / %s", formatBytesIEC(used), formatBytesIEC(total))
		}
		return fmt.Sprintf("- / %s", formatBytesIEC(total))
	}
	if hasUsage && used > 0 {
		return fmt.Sprintf("%s / -", formatBytesIEC(used))
	}
	return "-"
}

func formatGPUUsage(usage gpuUsageValues, count int64) string {
	parts := make([]string, 0, 3)
	if usage.hasUtil {
		parts = append(parts, fmt.Sprintf("%d%%", usage.utilPercent))
	}
	if usage.hasMemory {
		parts = append(parts, formatBytesIEC(usage.memoryUsedBytes))
	}
	if count > 0 {
		if count == 1 {
			parts = append(parts, "1 GPU")
		} else {
			parts = append(parts, fmt.Sprintf("%d GPUs", count))
		}
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, " / ")
}

func max64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func boolPtr(v bool) *bool { return &v }
