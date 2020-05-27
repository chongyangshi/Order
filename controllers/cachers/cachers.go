package cachers

import (
	"log"
	"time"

	"k8s.io/client-go/kubernetes"
)

// Init launches a series of pod controllers which may run pods mounting resources managed by
// Order. They provide an eventually consistent cache we use to determine whether a rolling
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

	// Block until all controllers have synced
	for {
		allSynced := true
		switch {
		case !dsController.synced():
			log.Println("DaemonSets controller not yet synced")
			allSynced = false
		case !deployController.synced():
			log.Println("Deployments controller not yet synced")
			allSynced = false
		case !jobsController.synced():
			log.Println("Jobs controller not yet synced")
			allSynced = false
		case !stsController.synced():
			log.Println("StatefulSets controller not yet synced")
			allSynced = false
		}

		if allSynced {
			log.Println("All cache controllers synced and ready")
			break
		}

		log.Println("Not all cache controllers synced, waiting for a short while before re-checking")
		time.Sleep(time.Millisecond * 2000)
	}
}
