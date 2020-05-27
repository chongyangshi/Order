package controllers

import (
	"log"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/icydoge/orderrrr/controllers/cachers"
	"github.com/icydoge/orderrrr/controllers/configmaps"
	"github.com/icydoge/orderrrr/controllers/secrets"
)

// Init launches a processor which is responsible for periodically inspecting managed
// resources on the buffer, and check whether all their depending pod controllers are
// up-to-date in terms of restarts. If any isn't and they can be restarted based on the
// cooldown configured, then the procesor will apply an annotation to ask Kubernetes to
// restart the said pod controller.
func Init(clientSet kubernetes.Interface, stopChan chan struct{}, resyncInterval time.Duration) {
	// Start cachers first to build a list of pod controllers
	cachers.Init(clientSet, stopChan, resyncInterval)
	log.Println("Started all cache controllers")

	// Now start controllers for managed resources.
	secretsController := secrets.NewSecretsController(clientSet, resyncInterval)
	go secretsController.Run(stopChan)

	configMapsController := configmaps.NewConfigMapsController(clientSet, resyncInterval)
	go configMapsController.Run(stopChan)

	log.Println("Started all managed resource controllers, waiting for them to sync")

	// Block until all controllers have synced
	for {
		allSynced := true
		switch {
		case !secretsController.Synced():
			log.Println("Secrets controller not yet synced")
			allSynced = false
		case !configMapsController.Synced():
			log.Println("ConfigMaps controller not yet synced")
			allSynced = false
		}

		if allSynced {
			log.Println("All managed resources controllers synced and ready")
			break
		}

		log.Println("Not all managed resources controllers synced, waiting for a short while before re-checking")
		time.Sleep(time.Millisecond * 2000)
	}
}
