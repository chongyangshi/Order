package cachers

import (
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
)

// statefulSetsCacheController holds an eventually consistent cache of statefulsets
// to allow orderrrr to determine what StatefulSet pods need to be rolling restarted
// quickly.
type statefulSetsCacheController struct {
	factory informers.SharedInformerFactory
	lister  appslisters.StatefulSetLister
	synced  cache.InformerSynced
}

// newStatefulSetsController initialises a StatefulSets controller
func newStatefulSetsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *statefulSetsCacheController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Apps().V1().StatefulSets()

	controller := &statefulSetsCacheController{
		factory: informerFactory,
	}

	controller.lister = informer.Lister()
	controller.synced = informer.Informer().HasSynced

	return controller
}

// run initialises and starts the controller
func (c *statefulSetsCacheController) run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting statefulset cache controller.")
	defer log.Println("Shutting down statefulset cache controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	<-stopChan
}
