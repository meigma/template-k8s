package controller

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	examplev1alpha1 "github.com/meigma/template-k8s/api/v1alpha1"
)

const (
	nginxContainerName    = "nginx"
	nginxConfigKey        = "nginx.conf"
	nginxConfigVolumeName = "nginx-config"
	configHashAnnotation  = "example.meigma.io/config-hash"

	defaultNginxImage = "nginx:stable"
	defaultNginxPort  = int32(80)

	conditionAvailable          = "Available"
	reasonDeploymentReady       = "DeploymentReady"
	reasonDeploymentProgressing = "DeploymentProgressing"
	reasonDeploymentStale       = "DeploymentStatusStale"
)

// NginxDeploymentReconciler reconciles a NginxDeployment object
type NginxDeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.meigma.io,resources=nginxdeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.meigma.io,resources=nginxdeployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.meigma.io,resources=nginxdeployments/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=configmaps;services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.23.3/pkg/reconcile
func (r *NginxDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	instance := &examplev1alpha1.NginxDeployment{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if apierrors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	config := nginxConfig(instance)
	if err := r.reconcileConfigMap(ctx, instance, config); err != nil {
		return ctrl.Result{}, err
	}

	deployment, err := r.reconcileDeployment(ctx, instance, config)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, instance); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.reconcileStatus(ctx, instance, deployment)
}

// SetupWithManager sets up the controller with the Manager.
func (r *NginxDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1alpha1.NginxDeployment{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Named("nginxdeployment").
		Complete(r)
}

func (r *NginxDeploymentReconciler) reconcileConfigMap(
	ctx context.Context,
	instance *examplev1alpha1.NginxDeployment,
	config string,
) error {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, configMap, func() error {
		configMap.Labels = labelsFor(instance)
		configMap.Data = map[string]string{
			nginxConfigKey: config,
		}
		return ctrl.SetControllerReference(instance, configMap, r.Scheme)
	})
	return err
}

func (r *NginxDeploymentReconciler) reconcileDeployment(
	ctx context.Context,
	instance *examplev1alpha1.NginxDeployment,
	config string,
) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, deployment, func() error {
		replicas := nginxReplicas(instance)
		deployment.Labels = labelsFor(instance)
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Selector = &metav1.LabelSelector{
			MatchLabels: selectorLabelsFor(instance),
		}
		deployment.Spec.Template.Labels = labelsFor(instance)
		deployment.Spec.Template.Annotations = map[string]string{
			configHashAnnotation: configHash(config),
		}
		deployment.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:  nginxContainerName,
				Image: nginxImage(instance),
				Ports: []corev1.ContainerPort{
					{
						Name:          "http",
						ContainerPort: nginxPort(instance),
						Protocol:      corev1.ProtocolTCP,
					},
				},
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      nginxConfigVolumeName,
						MountPath: "/etc/nginx/nginx.conf",
						SubPath:   nginxConfigKey,
						ReadOnly:  true,
					},
				},
			},
		}
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: nginxConfigVolumeName,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: instance.Name,
						},
						Items: []corev1.KeyToPath{
							{
								Key:  nginxConfigKey,
								Path: nginxConfigKey,
							},
						},
					},
				},
			},
		}
		return ctrl.SetControllerReference(instance, deployment, r.Scheme)
	})
	if err != nil {
		return nil, err
	}
	return deployment, r.Get(ctx, client.ObjectKeyFromObject(deployment), deployment)
}

func (r *NginxDeploymentReconciler) reconcileService(
	ctx context.Context,
	instance *examplev1alpha1.NginxDeployment,
) error {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	_, err := controllerutil.CreateOrPatch(ctx, r.Client, service, func() error {
		service.Labels = labelsFor(instance)
		service.Spec.Type = corev1.ServiceTypeClusterIP
		service.Spec.Selector = selectorLabelsFor(instance)
		service.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "http",
				Port:       nginxPort(instance),
				TargetPort: intstr.FromString("http"),
				Protocol:   corev1.ProtocolTCP,
			},
		}
		return ctrl.SetControllerReference(instance, service, r.Scheme)
	})
	return err
}

func (r *NginxDeploymentReconciler) reconcileStatus(
	ctx context.Context,
	instance *examplev1alpha1.NginxDeployment,
	deployment *appsv1.Deployment,
) error {
	original := instance.DeepCopy()
	instance.Status.ReadyReplicas = deployment.Status.ReadyReplicas
	meta.SetStatusCondition(&instance.Status.Conditions, availableCondition(instance, deployment))
	if equality.Semantic.DeepEqual(original.Status, instance.Status) {
		return nil
	}
	return r.Status().Patch(ctx, instance, client.MergeFrom(original))
}

func availableCondition(
	instance *examplev1alpha1.NginxDeployment,
	deployment *appsv1.Deployment,
) metav1.Condition {
	desired := nginxReplicas(instance)
	ready := deployment.Status.ReadyReplicas
	if deployment.Status.ObservedGeneration < deployment.Generation {
		return metav1.Condition{
			Type:               conditionAvailable,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: instance.Generation,
			Reason:             reasonDeploymentStale,
			Message: fmt.Sprintf(
				"Deployment status has observed generation %d, waiting for generation %d",
				deployment.Status.ObservedGeneration,
				deployment.Generation,
			),
		}
	}
	if desired == 0 || (ready >= desired && deploymentAvailable(deployment)) {
		return metav1.Condition{
			Type:               conditionAvailable,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: instance.Generation,
			Reason:             reasonDeploymentReady,
			Message:            fmt.Sprintf("Deployment has %d/%d ready replicas", ready, desired),
		}
	}

	return metav1.Condition{
		Type:               conditionAvailable,
		Status:             metav1.ConditionFalse,
		ObservedGeneration: instance.Generation,
		Reason:             reasonDeploymentProgressing,
		Message:            fmt.Sprintf("Deployment has %d/%d ready replicas", ready, desired),
	}
}

func deploymentAvailable(deployment *appsv1.Deployment) bool {
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func labelsFor(instance *examplev1alpha1.NginxDeployment) map[string]string {
	return selectorLabelsFor(instance)
}

func selectorLabelsFor(instance *examplev1alpha1.NginxDeployment) map[string]string {
	return map[string]string{
		"app.kubernetes.io/name":       "nginx",
		"app.kubernetes.io/instance":   instance.Name,
		"app.kubernetes.io/managed-by": "template-k8s",
	}
}

func nginxConfig(instance *examplev1alpha1.NginxDeployment) string {
	if instance.Spec.Config != "" {
		return instance.Spec.Config
	}
	return fmt.Sprintf(`events {}
http {
  server {
    listen %d;
    location / {
      return 200 "hello from template-k8s\n";
    }
  }
}
`, nginxPort(instance))
}

func nginxImage(instance *examplev1alpha1.NginxDeployment) string {
	if instance.Spec.Image != "" {
		return instance.Spec.Image
	}
	return defaultNginxImage
}

func nginxReplicas(instance *examplev1alpha1.NginxDeployment) int32 {
	if instance.Spec.Replicas != nil {
		return *instance.Spec.Replicas
	}
	return 1
}

func nginxPort(instance *examplev1alpha1.NginxDeployment) int32 {
	if instance.Spec.Port > 0 {
		return instance.Spec.Port
	}
	return defaultNginxPort
}

func configHash(config string) string {
	sum := sha256.Sum256([]byte(config))
	return hex.EncodeToString(sum[:])
}
