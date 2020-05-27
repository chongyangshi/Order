package secrets

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

	"github.com/icydoge/Order/controllers/buffer"
)

// SecretsController is a controller monitoring changes to secrets
type SecretsController struct {
	factory informers.SharedInformerFactory
	lister  corelisters.SecretLister
	Synced  cache.InformerSynced
}

// NewSecretsController initialises a secrets controller
func NewSecretsController(clientSet kubernetes.Interface, resyncInterval time.Duration) *SecretsController {
	informerFactory := informers.NewSharedInformerFactoryWithOptions(clientSet, resyncInterval)
	informer := informerFactory.Core().V1().Secrets()

	controller := &SecretsController{
		factory: informerFactory,
	}

	// We don't process delete, as unintentional deletion of a mounted secret coupled
	// with an immediate restart will make things fall over, which is usually not
	// desirable.
	informer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    controller.add,
		UpdateFunc: controller.update,
		DeleteFunc: func(obj interface{}) {},
	})

	controller.lister = informer.Lister()
	controller.Synced = informer.Informer().HasSynced

	return controller
}

// Run initialises and starts the controller
func (c *SecretsController) Run(stopChan chan struct{}) {
	defer runtime.HandleCrash()

	log.Println("Starting secret controller.")
	defer log.Println("Shutting down secret controller.")

	c.factory.Start(stopChan)

	if ok := cache.WaitForCacheSync(stopChan, c.Synced); !ok {
		log.Fatalln("Failed to wait for cache synchronization")
	}

	// Push all secrets into buffer to be processed if pod controllers
	// requiring them are not up-to-date.
	secrets, err := c.lister.List(labels.Everything())
	if err != nil {
		log.Fatalln("Failed to load secrets initially")
	}

	for _, secret := range secrets {
		item := buffer.BufferItem{
			TypeMeta:               secret.TypeMeta,
			Name:                   secret.GetName(),
			Namespace:              secret.GetNamespace(),
			PendingResourceVersion: secret.GetResourceVersion(),
			LastProcessed:          time.Now(),
			Attempts:               0,
		}
		buffer.Push(&item)
	}

	<-stopChan
}

func (c *SecretsController) add(obj interface{}) {
	secretState, ok := obj.(*corev1.Secret)
	if !ok {
		log.Printf("Could not process add: unexpected type for Secret: %v", obj)
		return
	}

	item := buffer.BufferItem{
		TypeMeta:               secretState.TypeMeta,
		Name:                   secretState.GetName(),
		Namespace:              secretState.GetNamespace(),
		PendingResourceVersion: secretState.GetResourceVersion(),
		LastProcessed:          time.Now(),
		Attempts:               0,
	}
	buffer.Push(&item)
}

func (c *SecretsController) update(old, new interface{}) {
	oldState, ok := old.(*corev1.Secret)
	if !ok {
		log.Printf("Could not process update: unexpected old state type for Secret: %v", new)
		return
	}
	newState, ok := new.(*corev1.Secret)
	if !ok {
		log.Printf("Could not process update: unexpected new state type for Secret: %v", new)
		return
	}

	if newState.GetResourceVersion() == oldState.GetResourceVersion() {
		// No change in secret
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
