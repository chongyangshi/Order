package controllers

import (
	"log"
	"time"

	"github.com/icydoge/orderrrr/controllers/cachers"
	"github.com/icydoge/orderrrr/controllers/configmaps"
	"github.com/icydoge/orderrrr/controllers/secrets"
	"k8s.io/client-go/kubernetes"
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

	log.Println("Started all managed resource controllers")
}

// @TODO: Control loop
// - Pop from buffer
// - Check whether resource is managed
// - If resource is not managed, discard
// - If resource is managed, continue
// - Push back into buffer if minimum cooldown not yet met
// - Look up currently matching pod controllers from cachers
// - For each matching pod controller, if namespace does not qualify, skip
// - If namespace qualifies, check whether managed resource hash annotations match
// - If matches, skip as pod controller is already up to date
// - If does not match, check if minimum cooldown from last restart qualifies
// - If minimum cooldown does not qualifies, push back into buffer and skip
// - If minimum cooldown qualifies, perform rolling restart and continue.
