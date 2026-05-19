package main

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	examplev1alpha1 "github.com/meigma/template-k8s/api/v1alpha1"
	"github.com/meigma/template-k8s/internal/controller"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(examplev1alpha1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func mustRegisterControllers(mgr manager.Manager) {
	err := (&controller.NginxDeploymentReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	exitOnError(err, "Failed to create controller", "controller", "nginxdeployment")
	// +kubebuilder:scaffold:builder
}

func mustRegisterHealthChecks(mgr manager.Manager) {
	exitOnError(mgr.AddHealthzCheck("healthz", healthz.Ping), "Failed to set up health check")
	exitOnError(mgr.AddReadyzCheck("readyz", healthz.Ping), "Failed to set up ready check")
}

func mustStartManager(mgr manager.Manager) {
	setupLog.Info("Starting manager")
	exitOnError(mgr.Start(ctrl.SetupSignalHandler()), "Failed to run manager")
}

func exitOnError(err error, msg string, keysAndValues ...any) {
	if err == nil {
		return
	}

	setupLog.Error(err, msg, keysAndValues...)
	os.Exit(1)
}
