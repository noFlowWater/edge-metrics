package kubernetes

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// CreateOrUpdateService creates or updates a Kubernetes Service
func CreateOrUpdateService(namespace, deviceID, deviceType string, port int) error {
	if !IsInitialized() {
		return fmt.Errorf("kubernetes client not initialized")
	}

	serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":         "edge-exporter",
				"device_id":   deviceID,
				"device_type": deviceType,
				"managed_by":  "edge-metrics-server",
			},
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None", // Headless service
			Ports: []corev1.ServicePort{
				{
					Name:       "metrics",
					Port:       int32(port),
					TargetPort: intstr.FromInt(port),
					Protocol:   corev1.ProtocolTCP,
				},
			},
		},
	}

	ctx := context.Background()
	client := GetClientset().CoreV1().Services(namespace)

	// Try to get existing service
	_, err := client.Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new service
			_, err = client.Create(ctx, service, metav1.CreateOptions{})
			return err
		}
		return err
	}

	// Update existing service
	_, err = client.Update(ctx, service, metav1.UpdateOptions{})
	return err
}

// DeleteService deletes a Kubernetes Service
func DeleteService(namespace, deviceID string) error {
	if !IsInitialized() {
		return fmt.Errorf("kubernetes client not initialized")
	}

	serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))
	ctx := context.Background()
	client := GetClientset().CoreV1().Services(namespace)

	err := client.Delete(ctx, serviceName, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil // Already deleted
	}
	return err
}

// GetService gets a Kubernetes Service
func GetService(namespace, deviceID string) (*corev1.Service, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	serviceName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))
	ctx := context.Background()
	client := GetClientset().CoreV1().Services(namespace)

	service, err := client.Get(ctx, serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return service, nil
}

// ListEdgeServices lists all edge-device-* services in a namespace
func ListEdgeServices(namespace string) ([]string, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	ctx := context.Background()
	client := GetClientset().CoreV1().Services(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: "managed_by=edge-metrics-server",
	}

	services, err := client.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	var serviceNames []string
	for _, svc := range services.Items {
		serviceNames = append(serviceNames, svc.Name)
	}

	return serviceNames, nil
}
