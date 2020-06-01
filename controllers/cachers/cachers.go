package cachers

import (
	"fmt"
	"time"

	"github.com/icydoge/Order/logging"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

var (
	dsController     *daemonSetsCacheController
	deployController *deploymentsCacheController
	jobsController   *jobsCacheController
	stsController    *statefulSetsCacheController
)

// Init launches a series of pod controllers which may run pods mounting resources managed by
// Order. They provide an eventually consistent cache we use to determine whether a rolling
// restart is required in response to changes to a managed resource.
func Init(clientSet kubernetes.Interface, stopChan chan struct{}, resyncInterval time.Duration) {
	dsController = newDaemonSetsController(clientSet, resyncInterval)
	go dsController.run(stopChan)

	deployController = newDeploymentsController(clientSet, resyncInterval)
	go deployController.run(stopChan)

	jobsController = newJobsController(clientSet, resyncInterval)
	go jobsController.run(stopChan)

	stsController = newStatefulSetsController(clientSet, resyncInterval)
	go stsController.run(stopChan)

	// Block until all controllers have synced
	for {
		allSynced := true
		switch {
		case !dsController.synced():
			logging.Log("DaemonSets controller not yet synced")
			allSynced = false
		case !deployController.synced():
			logging.Log("Deployments controller not yet synced")
			allSynced = false
		case !jobsController.synced():
			logging.Log("Jobs controller not yet synced")
			allSynced = false
		case !stsController.synced():
			logging.Log("StatefulSets controller not yet synced")
			allSynced = false
		}

		if allSynced {
			logging.Log("All cache controllers synced and ready")
			break
		}

		logging.Log("Not all cache controllers synced, waiting for a short while before re-checking")
		time.Sleep(time.Millisecond * 2000)
	}
}

// GetDaemonSets returns all DaemonSets currently in controller cache
func GetDaemonSets() ([]*appsv1.DaemonSet, error) {
	if dsController == nil {
		return nil, fmt.Errorf("DaemonSet controller is not yet initialised")
	}
	return dsController.lister.List(labels.Everything())
}

// GetDeployments returns all Deployments currently in controller cache
func GetDeployments() ([]*appsv1.Deployment, error) {
	if dsController == nil {
		return nil, fmt.Errorf("Deployment controller is not yet initialised")
	}
	return deployController.lister.List(labels.Everything())
}

// GetJobs returns all Jobs currently in controller cache
func GetJobs() ([]*batchv1.Job, error) {
	if dsController == nil {
		return nil, fmt.Errorf("Jobs controller is not yet initialised")
	}
	return jobsController.lister.List(labels.Everything())
}

// GetStatefulSets returns all StatefulSets currently in controller cache
func GetStatefulSets() ([]*appsv1.StatefulSet, error) {
	if dsController == nil {
		return nil, fmt.Errorf("StatefulSets controller is not yet initialised")
	}
	return stsController.lister.List(labels.Everything())
}
