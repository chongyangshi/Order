package cachers

import (
	"fmt"
	"time"

	"github.com/chongyangshi/Order/config"
	"github.com/chongyangshi/Order/logging"
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

// GetDaemonSets returns all DaemonSets currently in controller cache whose namespace
// we care about as set in config.
func GetDaemonSets() ([]*appsv1.DaemonSet, error) {
	if dsController == nil {
		return nil, fmt.Errorf("DaemonSet controller is not yet initialised")
	}

	dsControllers, err := dsController.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var results []*appsv1.DaemonSet
	for _, ds := range dsControllers {
		if ds == nil {
			// Should never happen
			continue
		}

		if inConfigNamespaces(ds.Namespace) {
			results = append(results, ds)
		}
	}

	return results, nil
}

// GetDeployments returns all Deployments currently in controller cache whose namespace
// we care about as set in config.
func GetDeployments() ([]*appsv1.Deployment, error) {
	if deployController == nil {
		return nil, fmt.Errorf("Deployment controller is not yet initialised")
	}

	deployControllers, err := deployController.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var results []*appsv1.Deployment
	for _, deploy := range deployControllers {
		if deploy == nil {
			// Should never happen
			continue
		}

		if inConfigNamespaces(deploy.Namespace) {
			results = append(results, deploy)
		}
	}

	return results, nil
}

// GetJobs returns all Jobs currently in controller cache whose namespace
// we care about as set in config.
func GetJobs() ([]*batchv1.Job, error) {
	if jobsController == nil {
		return nil, fmt.Errorf("Jobs controller is not yet initialised")
	}

	jobsControllers, err := jobsController.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var results []*batchv1.Job
	for _, job := range jobsControllers {
		if job == nil {
			// Should never happen
			continue
		}

		if inConfigNamespaces(job.Namespace) {
			results = append(results, job)
		}
	}

	return results, nil
}

// GetStatefulSets returns all StatefulSets currently in controller whose namespace
// we care about as set in config.
func GetStatefulSets() ([]*appsv1.StatefulSet, error) {
	if stsController == nil {
		return nil, fmt.Errorf("StatefulSets controller is not yet initialised")
	}

	stsControllers, err := stsController.lister.List(labels.Everything())
	if err != nil {
		return nil, err
	}

	var results []*appsv1.StatefulSet
	for _, sts := range stsControllers {
		if sts == nil {
			// Should never happen
			continue
		}

		if inConfigNamespaces(sts.Namespace) {
			results = append(results, sts)
		}
	}

	return results, nil
}

func inConfigNamespaces(namespace string) bool {
	if config.Config == nil {
		logging.Fatal("Config namespaces unexpectedly accessed before parsing when searching for %s", namespace)
	}

	// If no namespace specified in config, all namespaces are accepted except kube-system
	if len(config.Config.Namespaces) == 0 && namespace != "kube-system" {
		return true
	}

	// Otherwise, a namespace is accepted if it is whitelisted in config
	for _, ns := range config.Config.Namespaces {
		if namespace == ns {
			return true
		}
	}

	return false
}
