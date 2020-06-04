package processor

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"

	"github.com/icydoge/Order/config"
	"github.com/icydoge/Order/controllers"
	"github.com/icydoge/Order/logging"
	"github.com/icydoge/Order/proto"
)

// managedResource represent a resource which can be monitored for change by Order.
// The underlying supported types may be subject to future extension. Currently Secret
// and ConfigMap are supported.
type managedResource struct {
	secret    *corev1.Secret
	configMap *corev1.ConfigMap
	config    *proto.ManagedResource
}

func (r managedResource) getUID() string {
	switch {
	case r.secret != nil:
		return string(r.secret.GetUID())
	case r.configMap != nil:
		return string(r.configMap.GetUID())
	}

	return ""
}

func (r managedResource) exists() bool {
	switch {
	case r.secret != nil,
		r.configMap != nil:
		return true
	}

	return false
}

// A batch query to map managed resources specified in config to resources which actually
// exist in the cluster at any given point in time. We can't just run this once and keep
// a cache of results, as managed resources can be added or deleted between runs.
func getManagedResourcesByReference() ([]managedResource, error) {
	if config.Config == nil {
		logging.Fatal("Error: managed resources in config unexpectedly requested before config is parsed")
	}

	secrets, err := controllers.GetSecrets()
	if err != nil {
		return nil, err
	}

	configMaps, err := controllers.GetConfigMaps()
	if err != nil {
		return nil, err
	}

	var resources []managedResource
	for _, resource := range config.Config.ManagedResources {
		if resource == nil {
			// Should never happen
			continue
		}

		r := managedResource{}

		switch resource.Type {
		case proto.ManagedResourceTypeSecrets:
			secret := findSecretByReference(secrets, resource.Name, resource.Namespace)
			if secret == nil {
				logging.Debug("Managed Secret %s of namespace %s not found in controller cache", resource.Name, resource.Namespace)
				break
			}
			r.secret = secret

		case proto.ManagedResourceTypeConfigMaps:
			configMap := findConfigMapByReference(configMaps, resource.Name, resource.Namespace)
			if configMap == nil {
				logging.Debug("Managed ConfigMap %s of namespace %s not found in controller cache", resource.Name, resource.Namespace)
				break
			}
			r.configMap = configMap

		default:
			logging.Debug("Unsupported managed resource type %s, this should not have passed validation.", resource.Type)
		}

		if r.exists() {
			r.config = resource
			resources = append(resources, r)
		}
	}

	return resources, nil
}

func findSecretByReference(secrets []*corev1.Secret, name, namespace string) *corev1.Secret {
	for _, secret := range secrets {
		if secret.Name == name && secret.Namespace == namespace {
			return secret
		}
	}

	return nil
}

func findConfigMapByReference(configMaps []*corev1.ConfigMap, name, namespace string) *corev1.ConfigMap {
	for _, configMap := range configMaps {
		if configMap.Name == name && configMap.Namespace == namespace {
			return configMap
		}
	}

	return nil
}

// This is a representation for managed resources matched to a pod controller based on its
// references in its pod template.
type managedResourcesForPodController struct {
	resources []*managedResource
}

// getHash returns a hash identifying the state of managed resources loaded by the target
// pod controller at the time of last restart. To be used to identify managed resource
// versions from the last rolling restart in order.kube-system.com/managed-resources-hash.
func (rs *managedResourcesForPodController) getHash() (string, error) {
	// Maintain a stable order of resources based on their UIDs
	resources := rs.resources
	sort.SliceStable(resources, func(i, j int) bool { return resources[i].getUID() < resources[j].getUID() })

	var resourceVersions []string
	for _, r := range resources {
		switch {
		case r.secret != nil:
			resourceVersions = append(resourceVersions, fmt.Sprintf("Secret:%v:%s", r.secret.GetUID(), r.secret.ResourceVersion))
		case r.configMap != nil:
			resourceVersions = append(resourceVersions, fmt.Sprintf("ConfigMap:%v:%s", r.configMap.GetUID(), r.configMap.ResourceVersion))
		}
	}

	hasher := sha256.New()
	_, err := hasher.Write([]byte(strings.Join(resourceVersions, "-")))
	if err != nil {
		return "", err
	}

	return hex.Dump(hasher.Sum(nil)), nil
}
