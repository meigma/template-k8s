package main

import (
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	ctrl "sigs.k8s.io/controller-runtime"
)

func main() {
	options := mustParseManagerOptions(os.Args[1:])
	ctrl.SetLogger(mustNewControllerLogger(options, os.Stderr))

	mgr := mustNewManager(options)
	mustRegisterControllers(mgr)
	mustRegisterHealthChecks(mgr)
	mustStartManager(mgr)
}
