package processor

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// orderrrr.kube-system.com/managed-resources-hash

// managedResource represent a resources which can be managed by orderrrr, the underlying
// supported types may be subject to future extension. Currently Secret and ConfigMap.
type managedResource struct {
	secret    *corev1.Secret
	configMap *corev1.ConfigMap
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

type managedResources struct {
	resources []*managedResource
}

// getHash returns a hash identifying the state of managed resources loaded by the target
// pod controller at the time of last restart.
func (rs *managedResources) getHash() (string, error) {
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

// orderrrr.kube-system.com/last-rolling-restart

// parseLastRollingRestartTimeBestEffort returns current time if we fail to parse,
// which is a failing-safe behaviour to ensure we don't repeatedly restart a pod
// controller while inserting invalid timestamps.
func parseLastRollingRestartTimeBestEffort(labelValue string) time.Time {
	t, err := time.Parse(time.RFC3339, labelValue)
	if err != nil {
		return time.Now()
	}

	return t
}
