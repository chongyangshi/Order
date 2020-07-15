package main

import (
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/chongyangshi/Order/config"
	"github.com/chongyangshi/Order/controllers"
	"github.com/chongyangshi/Order/logging"
)

type stop struct{}

const (
	ingressMultihomeNamespace = "multihome-ingress-system"
	defaultConfigMountPath    = "/etc/order/config.yaml"
	resyncInterval            = time.Second * 30
)

var defaultKubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")

func main() {
	// Load application config first
	configMountPath := defaultConfigMountPath
	if os.Getenv("ORDER_CONFIG_PATH") != "" {
		configMountPath = os.Getenv("ORDER_CONFIG_PATH")
	}
	err := config.LoadConfig(configMountPath)
	if err != nil {
		logging.Fatal("Error loading config from %s: %v", configMountPath, err)
	}

	// Now load Kubernetes config
	var config *rest.Config

	if os.Getenv("KUBECONFIG") != "" {
		config, err = clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
		if err != nil {
			logging.Fatal("Cannot read kubeconfig from environment variable: %v", err.Error())
		}
		logging.Log("Using kubeconfig from environment variable location KUBECONFIG=%s", os.Getenv("KUBE_CONFIG"))
	} else {
		logging.Log("No KUBECONFIG found in environment, assumi we are in cluster, using in-cluster client config.")
		config, err = rest.InClusterConfig()
		if err != nil {
			logging.Fatal("Could not load in-cluster config: %v", err)
		}
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logging.Fatal("Error initialising Kubernetes Node Client based on kubeconfig: %v", err)
	}

	stopChan := make(chan struct{})
	defer close(stopChan)

	// Start controllers
	controllers.Init(clientSet, stopChan, resyncInterval)
	logging.Log("Started all controllers")

	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, syscall.SIGINT, syscall.SIGTERM)

	// Run forever (until interrupted)
	select {
	case <-osSignals:
		stopChan <- stop{}
	default:
	}
}
