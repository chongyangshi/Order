package configmaps

import (
	"log"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"

	"github.com/icydoge/orderrrr/controllers/buffer"
)

// ConfigMapsController is a controller monitoring changes to ConfigMaps
type ConfigMapsController struct {
	factory informers.SharedInformerFactory
	lister  corelisters.ConfigMapLister
	synced  cache.InformerSynced
}

// NewConfigMapsController initialises a ConfigMaps controller
func NewConfigMapsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *ConfigMapsController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Core().V1().ConfigMaps()

	controller := &ConfigMapsController{
		factory: informerFactory,
	}

	// We don't process delete, as unintentional deletion of a mounted configmap coupled
	// with an immediate restart will make things fall over, which is usually not
	// desirable.
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.add,
		UpdateFunc: controller.update,
		DeleteFunc: func(obj interface{}) {},
	})

	controller.lister = informer.Lister()
	controller.synced = informer.Informer().HasSynced

	return controller
}

// Run initialises and starts the controller
func (c *ConfigMapsController) Run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting configmap controller.")
	defer log.Println("Shutting down configmap controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	// Push all secrets into buffer to be processed if pod controllers
	// requiring them are not up-to-date.
	configMaps, err := c.lister.List(labels.Everything())
	if err != nil {
		log.Fatalln("Failed to load secrets initially")
	}

	for _, configMap := range configMaps {
		item := buffer.BufferItem{
			TypeMeta:               configMap.TypeMeta,
			Name:                   configMap.GetName(),
			Namespace:              configMap.GetNamespace(),
			PendingResourceVersion: configMap.GetResourceVersion(),
			LastProcessed:          time.Now(),
			Attempts:               0,
		}
		buffer.Push(&item)
	}

	<-stopChan
}

func (c *ConfigMapsController) add(obj interface{}) {
	configMapState, ok := obj.(*corev1.ConfigMap)
	if !ok {
		log.Printf("Could not process add: unexpected type for ConfigMap: %v", obj)
		return
	}

	if configMapState == nil {
		return
	}

	item := buffer.BufferItem{
		TypeMeta:               configMapState.TypeMeta,
		Name:                   configMapState.GetName(),
		Namespace:              configMapState.GetNamespace(),
		PendingResourceVersion: configMapState.GetResourceVersion(),
		LastProcessed:          time.Now(),
		Attempts:               0,
	}
	buffer.Push(&item)
}

func (c *ConfigMapsController) update(old, new interface{}) {
	oldState, ok := old.(*corev1.ConfigMap)
	if !ok {
		log.Printf("Could not process update: unexpected old state type for ConfigMap: %v", new)
		return
	}
	newState, ok := new.(*corev1.ConfigMap)
	if !ok {
		log.Printf("Could not process update: unexpected new state type for ConfigMap: %v", new)
		return
	}

	if newState.GetResourceVersion() == oldState.GetResourceVersion() {
		// No change in configmap
		return
	}

	item := buffer.BufferItem{
		TypeMeta:               newState.TypeMeta,
		Name:                   newState.GetName(),
		Namespace:              newState.GetNamespace(),
		PendingResourceVersion: newState.GetResourceVersion(),
		LastProcessed:          time.Now(),
		Attempts:               0,
	}
	buffer.Push(&item)
}
