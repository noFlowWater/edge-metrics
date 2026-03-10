package kubernetes

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

// InitClient initializes the Kubernetes client
// Tries in-cluster config first, then falls back to kubeconfig
func InitClient() error {
	config, err := rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, _ := os.UserHomeDir()
			kubeconfig = filepath.Join(home, ".kube", "config")
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return fmt.Errorf("failed to create Kubernetes config: %w", err)
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return nil
}

// GetClientset returns the initialized Kubernetes clientset
func GetClientset() *kubernetes.Clientset {
	return clientset
}

// IsInitialized checks if the Kubernetes client is initialized
func IsInitialized() bool {
	return clientset != nil
}

// HealthCheckResponse represents K8s health check result
type HealthCheckResponse struct {
	KubernetesAvailable bool              `json:"kubernetes_available"`
	ClientInitialized   bool              `json:"client_initialized"`
	NamespaceAccessible bool              `json:"namespace_accessible"`
	RBACPermissions     map[string]string `json:"rbac_permissions"`
}

// CheckHealth checks Kubernetes connectivity and permissions
func CheckHealth(namespace string) (*HealthCheckResponse, error) {
	response := &HealthCheckResponse{
		KubernetesAvailable: false,
		ClientInitialized:   IsInitialized(),
		NamespaceAccessible: false,
		RBACPermissions:     make(map[string]string),
	}

	if !IsInitialized() {
		return response, fmt.Errorf("kubernetes client not initialized")
	}

	response.KubernetesAvailable = true
	ctx := context.Background()

	// Check namespace accessibility
	_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
	if err != nil {
		response.RBACPermissions["namespace"] = fmt.Sprintf("error: %v", err)
	} else {
		response.NamespaceAccessible = true
		response.RBACPermissions["namespace"] = "ok"
	}

	// Check Services permissions
	_, err = clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		response.RBACPermissions["services"] = fmt.Sprintf("error: %v", err)
	} else {
		response.RBACPermissions["services"] = "ok"
	}

	// Check Endpoints permissions
	_, err = clientset.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{Limit: 1})
	if err != nil {
		response.RBACPermissions["endpoints"] = fmt.Sprintf("error: %v", err)
	} else {
		response.RBACPermissions["endpoints"] = "ok"
	}

	return response, nil
}
