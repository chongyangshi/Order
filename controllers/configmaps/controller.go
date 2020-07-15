package configmaps

import (
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/chongyangshi/Order/logging"
)

// ConfigMapsController is a controller monitoring changes to ConfigMaps
type ConfigMapsController struct {
	factory informers.SharedInformerFactory
	Lister  corelisters.ConfigMapLister
	Synced  cache.InformerSynced
}

// NewConfigMapsController initialises a ConfigMaps controller
func NewConfigMapsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *ConfigMapsController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Core().V1().ConfigMaps()

	controller := &ConfigMapsController{
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
func (c *ConfigMapsController) Run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	logging.Log("Starting configmap controller.")
	defer logging.Log("Shutting down configmap controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.Synced); !ok {
		logging.Fatal("Failed to wait for cache synchronization")
	}

	<-stopChan
}
