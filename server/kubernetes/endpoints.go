package kubernetes

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateOrUpdateEndpoints creates or updates a Kubernetes Endpoints
func CreateOrUpdateEndpoints(namespace, deviceID, ipAddress string, port int) error {
	if !IsInitialized() {
		return fmt.Errorf("kubernetes client not initialized")
	}

	endpointsName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))

	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      endpointsName,
			Namespace: namespace,
			Labels: map[string]string{
				"app":        "edge-exporter",
				"device_id":  deviceID,
				"managed_by": "edge-metrics-server",
			},
		},
		Subsets: []corev1.EndpointSubset{
			{
				Addresses: []corev1.EndpointAddress{
					{
						IP: ipAddress,
					},
				},
				Ports: []corev1.EndpointPort{
					{
						Name:     "metrics",
						Port:     int32(port),
						Protocol: corev1.ProtocolTCP,
					},
				},
			},
		},
	}

	ctx := context.Background()
	client := GetClientset().CoreV1().Endpoints(namespace)

	// Try to get existing endpoints
	_, err := client.Get(ctx, endpointsName, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			// Create new endpoints
			_, err = client.Create(ctx, endpoints, metav1.CreateOptions{})
			return err
		}
		return err
	}

	// Update existing endpoints
	_, err = client.Update(ctx, endpoints, metav1.UpdateOptions{})
	return err
}

// DeleteEndpoints deletes a Kubernetes Endpoints
func DeleteEndpoints(namespace, deviceID string) error {
	if !IsInitialized() {
		return fmt.Errorf("kubernetes client not initialized")
	}

	endpointsName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))
	ctx := context.Background()
	client := GetClientset().CoreV1().Endpoints(namespace)

	err := client.Delete(ctx, endpointsName, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil // Already deleted
	}
	return err
}

// GetEndpoints gets a Kubernetes Endpoints
func GetEndpoints(namespace, deviceID string) (*corev1.Endpoints, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	endpointsName := fmt.Sprintf("edge-device-%s", strings.ToLower(deviceID))
	ctx := context.Background()
	client := GetClientset().CoreV1().Endpoints(namespace)

	endpoints, err := client.Get(ctx, endpointsName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return endpoints, nil
}

// ListEdgeEndpoints lists all edge-device-* endpoints in a namespace
func ListEdgeEndpoints(namespace string) ([]string, error) {
	if !IsInitialized() {
		return nil, fmt.Errorf("kubernetes client not initialized")
	}

	ctx := context.Background()
	client := GetClientset().CoreV1().Endpoints(namespace)

	listOptions := metav1.ListOptions{
		LabelSelector: "managed_by=edge-metrics-server",
	}

	endpointsList, err := client.List(ctx, listOptions)
	if err != nil {
		return nil, err
	}

	var endpointsNames []string
	for _, ep := range endpointsList.Items {
		endpointsNames = append(endpointsNames, ep.Name)
	}

	return endpointsNames, nil
}
