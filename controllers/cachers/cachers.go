package cachers

import (
	"time"

	"k8s.io/client-go/kubernetes"
)

// Init launches a series of pod controllers which may run pods mounting resources managed by
// orderrrr. They provide an eventually consistent cache we use to determine whether a rolling
// restart is required in response to changes to a managed resource.
func Init(clientSet kubernetes.Interface, stopChan chan struct{}, resyncInterval time.Duration) {
	dsController := newDaemonSetsController(clientSet, resyncInterval)
	go dsController.run(stopChan)

	deployController := newDeploymentsController(clientSet, resyncInterval)
	go deployController.run(stopChan)

	jobsController := newJobsController(clientSet, resyncInterval)
	go jobsController.run(stopChan)

	stsController := newStatefulSetsController(clientSet, resyncInterval)
	go stsController.run(stopChan)
}
