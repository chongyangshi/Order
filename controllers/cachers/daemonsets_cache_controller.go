package cachers

import (
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/chongyangshi/Order/logging"
)

// daemonSetsCacheController holds an eventually consistent cache of daemonsets
// to allow Order to determine what DaemonSet pods need to be rolling
// restarted quickly.
type daemonSetsCacheController struct {
	factory informers.SharedInformerFactory
	lister  appslisters.DaemonSetLister
	synced  cache.InformerSynced
}

// newDaemonSetsController initialises a DaemonSets controller
func newDaemonSetsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *daemonSetsCacheController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Apps().V1().DaemonSets()

	controller := &daemonSetsCacheController{
		factory: informerFactory,
	}

	controller.lister = informer.Lister()
	controller.synced = informer.Informer().HasSynced

	return controller
}

// run initialises and starts the controller
func (c *daemonSetsCacheController) run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	logging.Log("Starting daemonset cache controller.")
	defer logging.Log("Shutting down daemonset cache controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		logging.Fatal("Failed to wait for cache synchronization")
	}

	<-stopChan
}
