package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	examplev1alpha1 "github.com/meigma/template-k8s/api/v1alpha1"
)

const testNamespace = "default"

var _ = Describe("NginxDeployment Controller", func() {
	ctx := context.Background()
	var controllerReconciler *NginxDeploymentReconciler

	BeforeEach(func() {
		controllerReconciler = &NginxDeploymentReconciler{
			Client: k8sClient,
			Scheme: k8sClient.Scheme(),
		}
	})

	It("creates owned nginx resources and reports initial availability", func() {
		spec := examplev1alpha1.NginxDeploymentSpec{
			Replicas: 2,
			Image:    "nginx:1.27",
			Port:     8080,
			Config:   "events {}\nhttp { server { listen 8080; } }\n",
		}
		resource := createNginxDeployment(ctx, "creates-children", spec)

		reconcileResource(ctx, controllerReconciler, resource)

		expectConfigMap(resource, spec.Config)
		expectDeployment(resource, spec)
		expectService(resource, spec.Port)

		current := fetchNginxDeployment(ctx, objectKeyFor(resource))
		Expect(current.Status.ReadyReplicas).To(Equal(int32(0)))
		expectAvailableCondition(current, metav1.ConditionFalse, reasonDeploymentProgressing)
	})

	It("updates owned resources when the spec changes", func() {
		initialSpec := examplev1alpha1.NginxDeploymentSpec{
			Replicas: 1,
			Image:    "nginx:stable",
			Port:     80,
			Config:   "events {}\nhttp { server { listen 80; } }\n",
		}
		resource := createNginxDeployment(ctx, "updates-children", initialSpec)
		reconcileResource(ctx, controllerReconciler, resource)
		initialDeployment := fetchDeployment(ctx, objectKeyFor(resource))
		initialHash := initialDeployment.Spec.Template.Annotations[configHashAnnotation]

		updated := fetchNginxDeployment(ctx, objectKeyFor(resource))
		updated.Spec = examplev1alpha1.NginxDeploymentSpec{
			Replicas: 3,
			Image:    "nginx:1.28",
			Port:     8081,
			Config:   "events {}\nhttp { server { listen 8081; } }\n",
		}
		Expect(k8sClient.Update(ctx, updated)).To(Succeed())

		reconcileResource(ctx, controllerReconciler, updated)

		expectConfigMap(updated, updated.Spec.Config)
		deployment := expectDeployment(updated, updated.Spec)
		Expect(deployment.Spec.Template.Annotations[configHashAnnotation]).NotTo(Equal(initialHash))
		expectService(updated, updated.Spec.Port)
	})

	It("uses a default nginx config when the spec omits config", func() {
		spec := examplev1alpha1.NginxDeploymentSpec{
			Replicas: 1,
			Image:    "nginx:stable",
			Port:     8082,
		}
		resource := createNginxDeployment(ctx, "defaults-config", spec)

		reconcileResource(ctx, controllerReconciler, resource)

		expectedConfig := nginxConfig(resource)
		expectConfigMap(resource, expectedConfig)
		deployment := expectDeployment(resource, spec)
		Expect(deployment.Spec.Template.Annotations).To(HaveKeyWithValue(configHashAnnotation, configHash(expectedConfig)))
		expectService(resource, spec.Port)
	})

	It("marks the resource available when the owned deployment has enough ready replicas", func() {
		spec := examplev1alpha1.NginxDeploymentSpec{
			Replicas: 2,
			Image:    "nginx:stable",
			Port:     80,
			Config:   "events {}\nhttp { server { listen 80; } }\n",
		}
		resource := createNginxDeployment(ctx, "reports-readiness", spec)

		reconcileResource(ctx, controllerReconciler, resource)
		current := fetchNginxDeployment(ctx, objectKeyFor(resource))
		Expect(current.Status.ReadyReplicas).To(Equal(int32(0)))
		expectAvailableCondition(current, metav1.ConditionFalse, reasonDeploymentProgressing)

		deployment := fetchDeployment(ctx, objectKeyFor(resource))
		deployment.Status.Replicas = spec.Replicas
		deployment.Status.ReadyReplicas = spec.Replicas
		deployment.Status.AvailableReplicas = spec.Replicas
		Expect(k8sClient.Status().Update(ctx, deployment)).To(Succeed())

		reconcileResource(ctx, controllerReconciler, resource)

		current = fetchNginxDeployment(ctx, objectKeyFor(resource))
		Expect(current.Status.ReadyReplicas).To(Equal(spec.Replicas))
		expectAvailableCondition(current, metav1.ConditionTrue, reasonDeploymentReady)
	})
})

func createNginxDeployment(
	ctx context.Context,
	name string,
	spec examplev1alpha1.NginxDeploymentSpec,
) *examplev1alpha1.NginxDeployment {
	resource := &examplev1alpha1.NginxDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: testNamespace,
		},
		Spec: spec,
	}
	Expect(k8sClient.Create(ctx, resource)).To(Succeed())
	DeferCleanup(cleanupNginxDeployment, ctx, types.NamespacedName{Name: name, Namespace: testNamespace})
	return resource
}

func cleanupNginxDeployment(ctx context.Context, key types.NamespacedName) {
	objects := []client.Object{
		&examplev1alpha1.NginxDeployment{ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: key.Name, Namespace: key.Namespace}},
	}
	for _, object := range objects {
		err := k8sClient.Delete(ctx, object)
		Expect(client.IgnoreNotFound(err)).To(Succeed())
	}
}

func reconcileResource(
	ctx context.Context,
	controllerReconciler *NginxDeploymentReconciler,
	resource *examplev1alpha1.NginxDeployment,
) {
	_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
		NamespacedName: objectKeyFor(resource),
	})
	Expect(err).NotTo(HaveOccurred())
}

func expectConfigMap(resource *examplev1alpha1.NginxDeployment, config string) {
	configMap := &corev1.ConfigMap{}
	Expect(k8sClient.Get(context.Background(), objectKeyFor(resource), configMap)).To(Succeed())
	expectManagedObject(configMap, resource)
	Expect(configMap.Data).To(HaveKeyWithValue(nginxConfigKey, config))
}

func expectDeployment(
	resource *examplev1alpha1.NginxDeployment,
	spec examplev1alpha1.NginxDeploymentSpec,
) *appsv1.Deployment {
	deployment := fetchDeployment(context.Background(), objectKeyFor(resource))
	expectManagedObject(deployment, resource)
	Expect(deployment.Spec.Replicas).NotTo(BeNil())
	Expect(*deployment.Spec.Replicas).To(Equal(spec.Replicas))
	Expect(deployment.Spec.Selector.MatchLabels).To(Equal(selectorLabelsFor(resource)))
	Expect(deployment.Spec.Template.Labels).To(Equal(labelsFor(resource)))
	Expect(deployment.Spec.Template.Annotations).To(HaveKeyWithValue(configHashAnnotation, configHash(nginxConfig(resource))))

	Expect(deployment.Spec.Template.Spec.Containers).To(HaveLen(1))
	container := deployment.Spec.Template.Spec.Containers[0]
	Expect(container.Name).To(Equal(nginxContainerName))
	Expect(container.Image).To(Equal(nginxImage(resource)))
	Expect(container.Ports).To(HaveLen(1))
	Expect(container.Ports[0].Name).To(Equal("http"))
	Expect(container.Ports[0].ContainerPort).To(Equal(nginxPort(resource)))
	Expect(container.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
	Expect(container.VolumeMounts).To(HaveLen(1))
	Expect(container.VolumeMounts[0].Name).To(Equal(nginxConfigVolumeName))
	Expect(container.VolumeMounts[0].MountPath).To(Equal("/etc/nginx/nginx.conf"))
	Expect(container.VolumeMounts[0].SubPath).To(Equal(nginxConfigKey))
	Expect(container.VolumeMounts[0].ReadOnly).To(BeTrue())

	Expect(deployment.Spec.Template.Spec.Volumes).To(HaveLen(1))
	volume := deployment.Spec.Template.Spec.Volumes[0]
	Expect(volume.Name).To(Equal(nginxConfigVolumeName))
	Expect(volume.ConfigMap).NotTo(BeNil())
	Expect(volume.ConfigMap.Name).To(Equal(resource.Name))
	Expect(volume.ConfigMap.Items).To(ConsistOf(corev1.KeyToPath{
		Key:  nginxConfigKey,
		Path: nginxConfigKey,
	}))
	return deployment
}

func expectService(resource *examplev1alpha1.NginxDeployment, port int32) {
	service := &corev1.Service{}
	Expect(k8sClient.Get(context.Background(), objectKeyFor(resource), service)).To(Succeed())
	expectManagedObject(service, resource)
	Expect(service.Spec.Type).To(Equal(corev1.ServiceTypeClusterIP))
	Expect(service.Spec.Selector).To(Equal(selectorLabelsFor(resource)))
	Expect(service.Spec.Ports).To(HaveLen(1))
	Expect(service.Spec.Ports[0].Name).To(Equal("http"))
	Expect(service.Spec.Ports[0].Port).To(Equal(port))
	Expect(service.Spec.Ports[0].TargetPort).To(Equal(intstr.FromString("http")))
	Expect(service.Spec.Ports[0].Protocol).To(Equal(corev1.ProtocolTCP))
}

func expectManagedObject(object metav1.Object, owner *examplev1alpha1.NginxDeployment) {
	Expect(object.GetLabels()).To(Equal(labelsFor(owner)))
	for _, reference := range object.GetOwnerReferences() {
		if reference.APIVersion == examplev1alpha1.GroupVersion.String() &&
			reference.Kind == "NginxDeployment" &&
			reference.Name == owner.Name &&
			reference.UID == owner.UID &&
			reference.Controller != nil &&
			*reference.Controller {
			return
		}
	}
	Fail(fmt.Sprintf("expected %s to be owned by %s", object.GetName(), owner.Name))
}

func expectAvailableCondition(
	resource *examplev1alpha1.NginxDeployment,
	status metav1.ConditionStatus,
	reason string,
) {
	condition := meta.FindStatusCondition(resource.Status.Conditions, conditionAvailable)
	Expect(condition).NotTo(BeNil())
	Expect(condition.Status).To(Equal(status))
	Expect(condition.Reason).To(Equal(reason))
	Expect(condition.ObservedGeneration).To(Equal(resource.Generation))
}

func fetchNginxDeployment(
	ctx context.Context,
	key types.NamespacedName,
) *examplev1alpha1.NginxDeployment {
	resource := &examplev1alpha1.NginxDeployment{}
	Expect(k8sClient.Get(ctx, key, resource)).To(Succeed())
	return resource
}

func fetchDeployment(ctx context.Context, key types.NamespacedName) *appsv1.Deployment {
	deployment := &appsv1.Deployment{}
	Expect(k8sClient.Get(ctx, key, deployment)).To(Succeed())
	return deployment
}

func objectKeyFor(instance *examplev1alpha1.NginxDeployment) types.NamespacedName {
	return types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}
}
