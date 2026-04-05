package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (statusErr *StatusError) listNodesDetails(cs *ClientSet) {
	clientset := cs.clientset
	// Configure slog logger
	//statusErr := &StatusError{}
	//logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	nodeSelectorLabels := os.Getenv("KUBERNETES_NODE_SELECTOR_JSON")
	nodeTolerations := os.Getenv("KUBERNETES_TOLERATIONS_JSON")

	selectors := map[string]string{}
	tolerations := []corev1.Toleration{}

	if strings.TrimSpace(nodeSelectorLabels) != "" {
		parsedSelectors, err := parseNodeSelectors(nodeSelectorLabels)
		if err != nil {
			statusErr.NodeStatus = fmt.Errorf("invalid KUBERNETES_NODE_SELECTOR_LABEL: %w", err)
			fmt.Printf("\n[%s][error] %v", time.Now().Format("2006-01-02 15:04:05"), statusErr.NodeStatus)
			return
		}
		selectors = parsedSelectors
	}

	if strings.TrimSpace(nodeTolerations) != "" {
		parsedTolerations, err := parseNodeTolerations(nodeTolerations)
		if err != nil {
			statusErr.NodeStatus = fmt.Errorf("invalid KUBERNETES_NODE_TOLERATIONS: %w", err)
			fmt.Printf("\n[%s][error] %v", time.Now().Format("2006-01-02 15:04:05"), statusErr.NodeStatus)
			return
		}
		tolerations = parsedTolerations
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		statusErr.NodeStatus = err
		fmt.Printf("\n[%s][error] listing nodes: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		return
	}

	filteredNodes := make([]corev1.Node, 0, len(nodes.Items))
	filtersProvided := len(selectors) > 0 || len(tolerations) > 0
	for _, nd := range nodes.Items {
		if !nodeMatchesSelectors(nd, selectors) {
			continue
		}
		if !nodeMatchesTolerations(nd, tolerations) {
			continue
		}
		filteredNodes = append(filteredNodes, nd)
	}

	if len(filteredNodes) == 0 {
		if filtersProvided {
			fmt.Printf("\n[%s][error] no nodes matched KUBERNETES_NODE_SELECTOR_LABEL/KUBERNETES_NODE_TOLERATIONS filters", time.Now().Format("2006-01-02 15:04:05"))
			statusErr.NodeStatus = fmt.Errorf("no nodes matched selector/toleration filters")
		} else {
			fmt.Printf("\n[%s][error] no schedulable nodes were returned by the cluster", time.Now().Format("2006-01-02 15:04:05"))
			statusErr.NodeStatus = fmt.Errorf("no nodes found in cluster")
		}
		return
	}

	for i, nd := range filteredNodes {
		var errs []string

		cpu := nd.Status.Capacity.Cpu().MilliValue()
		mem := nd.Status.Capacity.Memory().Value()
		storage := nd.Status.Capacity.StorageEphemeral().Value()
		memMB := mem / (1024 * 1024)
		storageMB := storage / (1024 * 1024)
		if cpu < 2000 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{
				fmt.Sprintf("cpu node %d", i): fmt.Errorf("insufficient %d", cpu),
			})
			errs = append(errs, fmt.Sprintf("CPU: %d", cpu))
		}
		if memMB < 4096 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{
				fmt.Sprintf("memory node %d", i): fmt.Errorf("insufficient %d", memMB),
			})
			errs = append(errs, fmt.Sprintf("MEM: %d MB", memMB))
		}
		if storageMB < 64 {
			statusErr.NodeResourceStatus = append(statusErr.NodeResourceStatus, map[string]error{
				fmt.Sprintf("storage node %d", i): fmt.Errorf("insufficient %d", storageMB),
			})
			errs = append(errs, fmt.Sprintf("STORAGE: %d MB", storageMB))
		}

		if len(errs) > 0 {
			fmt.Printf("\n[%s][error] node %d insufficient resources: %s", time.Now().Format("2006-01-02 15:04:05"), i+1, errs)
		} else {
			fmt.Printf("\n[%s][INFO] Node: %d, CPU: %d, MEM: %d MB, Storage: %d MB", time.Now().Format("2006-01-02 15:04:05"), i+1, cpu, memMB, storageMB)
		}
	}
}

func parseNodeSelectors(raw string) (map[string]string, error) {
	selectors := make(map[string]string)
	if strings.TrimSpace(raw) == "" {
		return selectors, nil
	}

	if err := json.Unmarshal([]byte(raw), &selectors); err == nil {
		return selectors, nil
	}

	for _, pair := range strings.Split(raw, ",") {
		kv := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(kv) != 2 || strings.TrimSpace(kv[0]) == "" {
			return nil, fmt.Errorf("selector must be valid JSON object or k:v list")
		}
		selectors[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
	}

	return selectors, nil
}

func parseNodeTolerations(raw string) ([]corev1.Toleration, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	var list []corev1.Toleration
	if err := json.Unmarshal([]byte(raw), &list); err == nil {
		return list, nil
	}

	var single corev1.Toleration
	if err := json.Unmarshal([]byte(raw), &single); err == nil {
		return []corev1.Toleration{single}, nil
	}

	for _, token := range strings.Split(raw, ",") {
		parts := strings.Split(strings.TrimSpace(token), ":")
		if len(parts) != 3 {
			return nil, fmt.Errorf("tolerations must be valid JSON or key:value:effect list")
		}

		list = append(list, corev1.Toleration{
			Key:      strings.TrimSpace(parts[0]),
			Value:    strings.TrimSpace(parts[1]),
			Effect:   corev1.TaintEffect(strings.TrimSpace(parts[2])),
			Operator: corev1.TolerationOpEqual,
		})
	}

	return list, nil
}

func nodeMatchesSelectors(node corev1.Node, selectors map[string]string) bool {
	if len(selectors) == 0 {
		return true
	}

	for key, value := range selectors {
		if node.Labels[key] != value {
			return false
		}
	}

	return true
}

func nodeMatchesTolerations(node corev1.Node, tolerations []corev1.Toleration) bool {
	if len(tolerations) == 0 {
		return true
	}

	for _, taint := range node.Spec.Taints {
		if taint.Effect != corev1.TaintEffectNoSchedule && taint.Effect != corev1.TaintEffectNoExecute {
			continue
		}

		if !isTaintTolerated(taint, tolerations) {
			return false
		}
	}

	return true
}

func isTaintTolerated(taint corev1.Taint, tolerations []corev1.Toleration) bool {
	for _, tol := range tolerations {
		effectMatches := tol.Effect == "" || tol.Effect == taint.Effect
		if !effectMatches {
			continue
		}

		switch tol.Operator {
		case corev1.TolerationOpExists:
			if tol.Key == "" || tol.Key == taint.Key {
				return true
			}
		default:
			if tol.Key == taint.Key && tol.Value == taint.Value {
				return true
			}
		}
	}

	return false
}
