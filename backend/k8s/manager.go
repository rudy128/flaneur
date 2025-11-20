package k8s

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sManager manages Kubernetes resources for WhatsApp services
type K8sManager struct {
	clientset *kubernetes.Clientset
	namespace string
	image     string
}

// NewK8sManager creates a new Kubernetes manager
func NewK8sManager() (*K8sManager, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first (when running inside K8s)
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig (for local development)
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = os.Getenv("HOME") + "/.kube/config"
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	namespace := os.Getenv("K8S_NAMESPACE")
	if namespace == "" {
		namespace = "social-api"
	}

	image := os.Getenv("K8S_WHATSAPP_IMAGE")
	if image == "" {
		image = "whatsapp-service:latest"
	}

	return &K8sManager{
		clientset: clientset,
		namespace: namespace,
		image:     image,
	}, nil
}

// GetNamespace returns the namespace the manager is operating in
func (k *K8sManager) GetNamespace() string {
	return k.namespace
}

// CreateWhatsAppPod creates a new WhatsApp service pod with a unique session ID
func (k *K8sManager) CreateWhatsAppPod(sessionID string, userID string) (*corev1.Pod, error) {
	ctx := context.Background()

	podName := fmt.Sprintf("whatsapp-%s", sessionID)
	serviceName := fmt.Sprintf("whatsapp-svc-%s", sessionID)
	pvcName := fmt.Sprintf("whatsapp-pvc-%s", sessionID)

	// Create PVC for persistent storage (tokens, auth data)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcName,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":        "whatsapp-service",
				"session-id": sessionID,
				"user-id":    userID,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse("1Gi"),
				},
			},
		},
	}

	_, err := k.clientset.CoreV1().PersistentVolumeClaims(k.namespace).Create(ctx, pvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create PVC: %v", err)
	}

	// Create Pod
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      podName,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":        "whatsapp-service",
				"session-id": sessionID,
				"user-id":    userID,
			},
			Annotations: map[string]string{
				"created-at": time.Now().Format(time.RFC3339),
			},
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyAlways,
			Containers: []corev1.Container{
				{
					Name:            "whatsapp",
					Image:           k.image,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							ContainerPort: 8083,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Env: []corev1.EnvVar{
						{
							Name:  "SESSION_ID",
							Value: sessionID,
						},
						{
							Name:  "USER_ID",
							Value: userID,
						},
						{
							Name:  "PORT",
							Value: "8083",
						},
					},
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "whatsapp-data",
							MountPath: "/app/tokens",
						},
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("128Mi"),
							corev1.ResourceCPU:    resource.MustParse("100m"),
						},
						Limits: corev1.ResourceList{
							corev1.ResourceMemory: resource.MustParse("512Mi"),
							corev1.ResourceCPU:    resource.MustParse("500m"),
						},
					},
					LivenessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.FromInt(8083),
							},
						},
						InitialDelaySeconds: 30,
						PeriodSeconds:       10,
						TimeoutSeconds:      5,
						FailureThreshold:    3,
					},
					ReadinessProbe: &corev1.Probe{
						ProbeHandler: corev1.ProbeHandler{
							HTTPGet: &corev1.HTTPGetAction{
								Path: "/health",
								Port: intstr.FromInt(8083),
							},
						},
						InitialDelaySeconds: 10,
						PeriodSeconds:       5,
						TimeoutSeconds:      3,
						FailureThreshold:    3,
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "whatsapp-data",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: pvcName,
						},
					},
				},
			},
		},
	}

	createdPod, err := k.clientset.CoreV1().Pods(k.namespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		// Clean up PVC if pod creation fails
		_ = k.clientset.CoreV1().PersistentVolumeClaims(k.namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
		return nil, fmt.Errorf("failed to create pod: %v", err)
	}

	// Create Service for the pod
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      serviceName,
			Namespace: k.namespace,
			Labels: map[string]string{
				"app":        "whatsapp-service",
				"session-id": sessionID,
				"user-id":    userID,
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Port:       8083,
					TargetPort: intstr.FromInt(8083),
					Protocol:   corev1.ProtocolTCP,
				},
			},
			Selector: map[string]string{
				"session-id": sessionID,
			},
		},
	}

	_, err = k.clientset.CoreV1().Services(k.namespace).Create(ctx, service, metav1.CreateOptions{})
	if err != nil {
		// Clean up pod and PVC if service creation fails
		_ = k.clientset.CoreV1().Pods(k.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
		_ = k.clientset.CoreV1().PersistentVolumeClaims(k.namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
		return nil, fmt.Errorf("failed to create service: %v", err)
	}

	return createdPod, nil
}

// DeleteWhatsAppPod deletes a WhatsApp service pod and its resources
func (k *K8sManager) DeleteWhatsAppPod(sessionID string) error {
	ctx := context.Background()

	podName := fmt.Sprintf("whatsapp-%s", sessionID)
	serviceName := fmt.Sprintf("whatsapp-svc-%s", sessionID)
	pvcName := fmt.Sprintf("whatsapp-pvc-%s", sessionID)

	// Delete Service
	err := k.clientset.CoreV1().Services(k.namespace).Delete(ctx, serviceName, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Warning: failed to delete service %s: %v\n", serviceName, err)
	}

	// Delete Pod
	err = k.clientset.CoreV1().Pods(k.namespace).Delete(ctx, podName, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Warning: failed to delete pod %s: %v\n", podName, err)
	}

	// Delete PVC
	err = k.clientset.CoreV1().PersistentVolumeClaims(k.namespace).Delete(ctx, pvcName, metav1.DeleteOptions{})
	if err != nil {
		fmt.Printf("Warning: failed to delete PVC %s: %v\n", pvcName, err)
	}

	return nil
}

// GetPodStatus gets the status of a WhatsApp pod
func (k *K8sManager) GetPodStatus(sessionID string) (string, error) {
	ctx := context.Background()
	podName := fmt.Sprintf("whatsapp-%s", sessionID)

	pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod: %v", err)
	}

	return string(pod.Status.Phase), nil
}

// GetPodIP gets the IP address of a WhatsApp pod
func (k *K8sManager) GetPodIP(sessionID string) (string, error) {
	ctx := context.Background()
	podName := fmt.Sprintf("whatsapp-%s", sessionID)

	pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get pod: %v", err)
	}

	if pod.Status.PodIP == "" {
		return "", fmt.Errorf("pod IP not assigned yet")
	}

	return pod.Status.PodIP, nil
}

// GetServiceURL gets the service URL for a WhatsApp pod
func (k *K8sManager) GetServiceURL(sessionID string) string {
	serviceName := fmt.Sprintf("whatsapp-svc-%s", sessionID)
	return fmt.Sprintf("http://%s.%s.svc.cluster.local:8083", serviceName, k.namespace)
}

// WaitForPodReady waits for a pod to be ready
func (k *K8sManager) WaitForPodReady(sessionID string, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	podName := fmt.Sprintf("whatsapp-%s", sessionID)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for pod to be ready")
		default:
			pod, err := k.clientset.CoreV1().Pods(k.namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("failed to get pod: %v", err)
			}

			// Check if pod is ready
			for _, condition := range pod.Status.Conditions {
				if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
					return nil
				}
			}

			time.Sleep(2 * time.Second)
		}
	}
}

// ListWhatsAppPods lists all WhatsApp service pods
func (k *K8sManager) ListWhatsAppPods() ([]corev1.Pod, error) {
	ctx := context.Background()

	pods, err := k.clientset.CoreV1().Pods(k.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=whatsapp-service",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	return pods.Items, nil
}

// GetPodLogs gets logs from a WhatsApp pod
func (k *K8sManager) GetPodLogs(sessionID string, tailLines int64) (string, error) {
	ctx := context.Background()
	podName := fmt.Sprintf("whatsapp-%s", sessionID)

	req := k.clientset.CoreV1().Pods(k.namespace).GetLogs(podName, &corev1.PodLogOptions{
		TailLines: &tailLines,
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get logs: %v", err)
	}
	defer logs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return "", fmt.Errorf("failed to copy logs: %v", err)
	}

	return buf.String(), nil
}
