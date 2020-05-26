package proto

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	coreV1 "k8s.io/api/core/v1"
)

const (
	// LabelPrefix is the shared prefix managed by orderrrr
	LabelPrefix = "orderrrr.kube-system.com"

	// LabelLastRollingRestart is an RFC3339 timestamp recording when
	// orderrrr last performed a rolling restart on a pod controller
	LabelLastRollingRestart = "last-rolling-restart"

	// LabelManagedResourcesHash is a SHA256 hash over all resources
	// managed by orderrrr on the target pod controller in order, in
	// case the controller is falling very far behind due to prior
	// errors, this helps us determine whether rolling restart is
	// required on a pod controller
	LabelManagedResourcesHash = "managed-resources-hash"
)

// ManagedResources represent a slice of resources which can be managed by orderrrr
// used by the target object.
type ManagedResources struct {
	Resources []*ManagedResource
}

// ManagedResource represent a resources which can be managed by orderrrr, the underlying
// supported types may be subject to future extension.
type ManagedResource struct {
	Secret    *coreV1.Secret
	ConfigMap *coreV1.ConfigMap
}

// GetHash returns a hash identifying the state of managed resources loaded by the target
// pod controller at the time of last restart.
func (rs ManagedResources) GetHash() (string, error) {
	var resources []string
	for _, r := range rs.Resources {
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

// ParseLastRollingRestartTimeBestEffort returns current time if we fail to parse,
// which is a failing-safe behaviour to ensure we don't repeatedly restart a pod
// controller while inserting invalid timestamps.
func ParseLastRollingRestartTimeBestEffort(labelValue string) time.Time {
	t, err := time.Parse(time.RFC3339, labelValue)
	if err != nil {
		return time.Now()
	}

	return t
}
