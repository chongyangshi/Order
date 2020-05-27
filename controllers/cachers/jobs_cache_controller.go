package cachers

import (
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	batchlisters "k8s.io/client-go/listers/batch/v1"
	"k8s.io/client-go/tools/cache"
)

// jobsCacheController holds an eventually consistent cache of batch jobs
// to allow Order to determine what Job pods need to be rolling restarted
// quickly.
type jobsCacheController struct {
	factory informers.SharedInformerFactory
	lister  batchlisters.JobLister
	synced  cache.InformerSynced
}

// jewJobsController initialises a Jobs controller
func newJobsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *jobsCacheController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Batch().V1().Jobs()

	controller := &jobsCacheController{
		factory: informerFactory,
	}

	controller.lister = informer.Lister()
	controller.synced = informer.Informer().HasSynced

	return controller
}

// Run initialises and starts the controller
func (c *jobsCacheController) run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting job cache controller.")
	defer log.Println("Shutting down job cache controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	<-stopChan
}
