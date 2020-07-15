package controllers

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	"github.com/chongyangshi/Order/controllers/cachers"
	"github.com/chongyangshi/Order/controllers/configmaps"
	"github.com/chongyangshi/Order/controllers/secrets"
	"github.com/chongyangshi/Order/logging"
)

var (
	configMapsController *configmaps.ConfigMapsController
	secretsController    *secrets.SecretsController
)

// Init launches a processor which is responsible for periodically inspecting managed
// resources on the buffer, and check whether all their depending pod controllers are
// up-to-date in terms of restarts. If any isn't and they can be restarted based on the
// cooldown configured, then the procesor will apply an annotation to ask Kubernetes to
// restart the said pod controller.
func Init(clientSet kubernetes.Interface, stopChan chan struct{}, resyncInterval time.Duration) {
	// Start cachers first to build a list of pod controllers
	cachers.Init(clientSet, stopChan, resyncInterval)
	logging.Log("Started all cache controllers")

	// Now start controllers for managed resources.
	secretsController = secrets.NewSecretsController(clientSet, resyncInterval)
	go secretsController.Run(stopChan)

	configMapsController = configmaps.NewConfigMapsController(clientSet, resyncInterval)
	go configMapsController.Run(stopChan)

	logging.Log("Started all managed resource controllers, waiting for them to sync")

	// Block until all controllers have synced
	for {
		allSynced := true
		switch {
		case !secretsController.Synced():
			logging.Debug("Secrets controller not yet synced")
			allSynced = false
		case !configMapsController.Synced():
			logging.Debug("ConfigMaps controller not yet synced")
			allSynced = false
		}

		if allSynced {
			logging.Log("All managed resources controllers synced and ready")
			break
		}

		logging.Log("Not all managed resources controllers synced, waiting for a short while before re-checking")
		time.Sleep(time.Millisecond * 2000)
	}
}

// GetSecrets returns all Secrets currently in controller cache, whether managed or not
func GetSecrets() ([]*corev1.Secret, error) {
	if secretsController == nil {
		return nil, fmt.Errorf("Secret controller is not yet initialised")
	}
	return secretsController.Lister.List(labels.Everything())
}

// GetConfigMaps returns all ConfigMaps currently in controller cache
func GetConfigMaps() ([]*corev1.ConfigMap, error) {
	if configMapsController == nil {
		return nil, fmt.Errorf("ConfigMap controller is not yet initialised")
	}
	return configMapsController.Lister.List(labels.Everything())
}
