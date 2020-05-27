package proto

import (
	"time"

	"github.com/prometheus/common/log"
)

const (
	controllerResyncSafetyLowerBound = time.Second * 15
	restartColldownSafetyLowerBound  = time.Second * 30
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

	ManagedResourceTypeSecrets    = "Secrets"
	ManagedResourceTypeConfigMaps = "ConfigMaps"

	PodControllerTypeDaemonSets   = "DaemonSets"
	PodControllerTypeDeployments  = "Deployments"
	PodControllerTypeJobs         = "Jobs"
	PodControllerTypeStatefulSets = "StatefulSets"
)

// OrderrrrConfig is a structue of system-wide and managed resource-specific
// configurations for orderrrr.
type OrderrrrConfig struct {
	// Version represents the config version of orderrrr
	Version float64 `yaml:"version"`

	// Namespaces represents rules under which orderrrr should eithr action on
	// or ignore a pod controller, depending on what namespace it lives in.
	Namespaces struct {
		// If Whitelist is set, orderrrr will only perform rolling restarts on pod
		// controllers in the specific namespaces listed. This includes if you
		// want to allow orderrrr to perform rolling restarts in kube-system
		// namespace, which is not recommended.
		Whitelist []string `yaml:"whitelist"`

		// If Blacklist is set, orderrrr will only perform rolling restarts on pod
		// controllers in the specific namespaces listed. kube-system will be
		// blacklisted by default. If you want to allow orderrrr to perform rolling
		// restarts in kube-system namespace, which is not recommended.
		Blacklist []string `yaml:"blacklist"`
	} `yaml:"namespace"`

	// ControllerResyncDuration is a Go duration which defines how frequently controllers
	// should do a full refresh to ensure that their state is up to date with what's in
	// cluster. This is a sanity check for eventual consistency in Kubernetes Controllers.
	// For small clusters, 1m+ is recommended. For large clusters, consider 5m+.
	ControllerResyncDuration string `yaml:"controller_resync_duration"`

	// DefaultRestartCooldown is a Go duration which defines at most how frequently can
	// orderrrr restart a pod controller, in order to prevent thrashing. Only positive
	// durations are acceptable. For small clusters, 2m+ is recommended. For large clusters,
	// consider 5m+. This can be overridden by restart_cooldown fields in individual
	// managed resource configurations.
	DefaultRestartCooldown string `yaml:"default_restart_cooldown"`

	// ManagedResources are Secrets and ConfigMaps, which when updated we want orderrrr to
	// perform automatic rolling restarts, subject to namespace and restart cooldown
	// validation.
	ManagedResources []*ManagedResource `yaml:"managed_resources"`
}

// ManagedResource represents a mountable or referenceable resource whose changes are
// monitored by orderrrr.
type ManagedResource struct {
	// Type is a Kubernetes resource type for the managed resource, currently either Secrets
	// or ConfigMaps.
	Type string `yaml:"type"`

	// Name of the managed resource
	Name string `yaml:"name"`

	// Namespace of the managed resource
	Namespace string `yaml:"namespace"`

	// AdditionalControllers specify pod controllers which should be restarted if this managed
	// resource updates, even if it is not formally linked to the managed resource via its
	// pod template.
	AdditionalControllers []struct {
		// Name of the nominated pod controller
		Name string `yaml:"name"`

		// Namespace of the nominated pod controller
		Namespace string `yaml:"namespace"`

		// Type of the nominated pod controller, one of DaemonSets, Deployments, Jobs, or
		// StatefulSets.
		Type string `yaml:"type"`
	} `yaml:"additional_controllers"`

	// AvoidControllers specify pod controllers which should not be restarted when
	// this managed resource updates, even if it is formally linked to the managed resource
	// via its pod template.
	AvoidControllers []struct {
		// Name of the nominated pod controller
		Name string `yaml:"name"`

		// Namespace of the nominated pod controller
		Namespace string `yaml:"namespace"`

		// Type of the nominated pod controller, one of DaemonSets, Deployments, Jobs, or
		// StatefulSets.
		Type string `yaml:"type"`
	} `yaml:"avoid_controllers"`

	// AvoidAllControllersUnlessWhitelisted can be set instead of AvoidControllers to make
	// orderrrr avoid performing actions on all controllers unless it is explicitly whitelisted
	// as AdditionalControllers.
	AvoidAllControllersUnlessWhitelisted bool `yaml:"avoid_all_controllers_unless_whitelisted"`

	// RestartCooldown is a Go duration which defines how frequently a pod controller can
	// be restarted for this resource. This overrides global defaults.
	RestartCooldown string `yaml:"restart_cooldown"`
}

func ValidateManagedResourceType(t string) bool {
	switch t {
	case ManagedResourceTypeConfigMaps, ManagedResourceTypeSecrets:
		return true
	}

	return false
}

func ValidatePodControllerType(t string) bool {
	switch t {
	case PodControllerTypeDaemonSets,
		PodControllerTypeDeployments,
		PodControllerTypeJobs,
		PodControllerTypeStatefulSets:
		return true
	}

	return false
}

func GetControllerResyncPeriod(d string) time.Duration {
	t, err := time.ParseDuration(d)
	if err != nil {
		log.Errorf("Error parsing controller resync period %s, using default %v", d, controllerResyncSafetyLowerBound)
		return controllerResyncSafetyLowerBound
	}

	if t < controllerResyncSafetyLowerBound {
		log.Errorf("Specified controller resync period %s below safety threshold, using default %v", d, controllerResyncSafetyLowerBound)
		return controllerResyncSafetyLowerBound
	}

	return t
}

func GetRestartCooldownPeriod(d string) time.Duration {
	t, err := time.ParseDuration(d)
	if err != nil {
		log.Errorf("Error parsing restart cooldown period %s, using default %v", d, restartColldownSafetyLowerBound)
		return restartColldownSafetyLowerBound
	}

	if t < controllerResyncSafetyLowerBound {
		log.Errorf("Specified restart cooldown period %s below safety threshold, using default %v", d, restartColldownSafetyLowerBound)
		return restartColldownSafetyLowerBound
	}

	return t
}
