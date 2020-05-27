package processor

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
)

// managedResource represent a resources which can be managed by orderrrr, the underlying
// supported types may be subject to future extension. Currently Secret and ConfigMap.
type managedResource struct {
	Secret    *corev1.Secret
	ConfigMap *corev1.ConfigMap
}

type managedResources struct {
	resources []*managedResource
}

// getHash returns a hash identifying the state of managed resources loaded by the target
// pod controller at the time of last restart.
func (rs managedResources) getHash() (string, error) {
	var resources []string
	for _, r := range rs.resources {
		switch {
		case r.Secret != nil:
			resources = append(resources, fmt.Sprintf("Secret:%v:%s", r.Secret.GetUID(), r.Secret.ResourceVersion))
		case r.ConfigMap != nil:
			resources = append(resources, fmt.Sprintf("ConfigMap:%v:%s", r.ConfigMap.GetUID(), r.ConfigMap.ResourceVersion))
		}
	}

	hasher := sha256.New()
	_, err := hasher.Write([]byte(strings.Join(resources, "-")))
	if err != nil {
		return "", err
	}

	return hex.Dump(hasher.Sum(nil)), nil
}

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
