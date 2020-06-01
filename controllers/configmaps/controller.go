package configmaps

import (
	"log"
	"time"

	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

// ConfigMapsController is a controller monitoring changes to ConfigMaps
type ConfigMapsController struct {
	factory informers.SharedInformerFactory
	lister  corelisters.ConfigMapLister
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

	controller.lister = informer.Lister()
	controller.Synced = informer.Informer().HasSynced

	return controller
}

// Run initialises and starts the controller
func (c *ConfigMapsController) Run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting configmap controller.")
	defer log.Println("Shutting down configmap controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.Synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	<-stopChan
}
