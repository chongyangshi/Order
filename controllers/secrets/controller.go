package secrets

import (
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/icydoge/Order/logging"
)

// SecretsController is a controller monitoring changes to secrets
type SecretsController struct {
	factory informers.SharedInformerFactory
	Lister  corelisters.SecretLister
	Synced  cache.InformerSynced
}

// NewSecretsController initialises a secrets controller
func NewSecretsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *SecretsController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Core().V1().Secrets()

	controller := &SecretsController{
		factory: informerFactory,
	}

	// We don't process informer events for the time being, and just rely on
	// fixed interval resynchronizations.
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) {},
		UpdateFunc: func(old interface{}, new interface{}) {},
		DeleteFunc: func(obj interface{}) {},
	})

	controller.Lister = informer.Lister()
	controller.Synced = informer.Informer().HasSynced

	return controller
}

// Run initialises and starts the controller
func (c *SecretsController) Run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	logging.Log("Starting secret controller.")
	defer logging.Log("Shutting down secret controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.Synced); !ok {
		logging.Fatal("Failed to wait for cache synchronization")
	}

	<-stopChan
}
