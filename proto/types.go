package proto

import (
	"time"
)

const (
	// LabelPrefix is the shared prefix managed by Order
	LabelPrefix = "order.kube-system.com"

	// LabelLastRollingRestart is an RFC3339 timestamp recording when
	// Order last performed a rolling restart on a pod controller
	LabelLastRollingRestart = "last-rolling-restart"

	// LabelManagedResourcesHash is a SHA256 hash over all resources
	// managed by Order on the target pod controller in order, in
	// case the controller is falling very far behind due to prior
	// errors, this helps us determine whether rolling restart is
	// required on a pod controller
	LabelManagedResourcesHash = "managed-resources-hash"

	ManagedResourceTypeSecrets    = "Secrets"
	ManagedResourceTypeConfigMaps = "ConfigMaps"

	PodControllerTypeDaemonSets   = "DaemonSet"
	PodControllerTypeDeployments  = "Deployment"
	PodControllerTypeJobs         = "Job"
	PodControllerTypeStatefulSets = "StatefulSet"

	AllNamespaces = "*"
)

// OrderConfig is a structue of system-wide and managed resource-specific
// configurations for Order.
type OrderConfig struct {
	// Version represents the config version of Order
	Version float64 `yaml:"version"`

	// Namespaces represents rules under which Order should either action on
	// or ignore a pod controller, depending on what namespace it lives in.
	// If set, this can preclude pod controllers even if they are secified as
	// whitelisted_controllers for a managed resource. If not set or empty,
	// Order will be applicable to all namespaces except kube-system.
	Namespaces []string `yaml:"namespaces"`

	// ControllerResyncDuration is a Go duration which defines how frequently controllers
	// should do a full refresh to ensure that their state is up to date with what's in
	// cluster. This is a sanity check for eventual consistency in Kubernetes Controllers.
	// For small clusters, 1m+ is recommended. For large clusters, consider 5m+. This value
	// cannot be below 15s.
	ControllerResyncDuration    string `yaml:"controller_resync_duration"`
	XXXControllerResyncDuration time.Duration

	// RestartCooldown is a Go duration which defines at most how frequently can Order
	// restart a pod controller, in order to prevent thrashing. Only positive durations
	// are acceptable. For small clusters, 2m+ is recommended. For large clusters,
	// consider 5m+. This value cannot be below 30s.
	RestartCooldown          string `yaml:"default_restart_cooldown"`
	XXXParsedRestartCooldown time.Duration

	// PodControllerStagger is a Go duration which prevents Order from initiating subsequent
	// restarts too quickly and hence causing problems in the cluster, by setting a minimum
	// stagger interval. This value cannot be below 5s.
	PodControllerStagger          string `yaml:"pod_controller_stagger"`
	XXXParsedPodControllerStagger time.Duration

	// ManagedResources are Secrets and ConfigMaps, which when updated we want Order to
	// perform automatic rolling restarts, subject to namespace and restart cooldown
	// validation.
	ManagedResources []*ManagedResource `yaml:"managed_resources"`

	// DebugOutput controls whether we print debug messages to stdout at debug level
	DebugOutput bool `yaml:"debug_output"`
}

// ManagedResource represents a mountable or referenceable resource whose changes are
// monitored by Order.
type ManagedResource struct {
	// Type is a Kubernetes resource type for the managed resource, currently either Secrets
	// or ConfigMaps.
	Type string `yaml:"type"`

	// Name of the managed resource
	Name string `yaml:"name"`

	// Namespace of the managed resource
	Namespace string `yaml:"namespace"`

	// WhitelistedControllers if not empty will restrict pod controllers to be restarted
	// to those matching this list only. It takes precedence over blacklisted_controllers
	// below
	WhitelistedControllers []struct {
		// Name of the nominated pod controller
		Name string `yaml:"name"`

		// Namespace of the nominated pod controller
		Namespace string `yaml:"namespace"`

		// Type of the nominated pod controller, one of DaemonSets, Deployments, Jobs, or
		// StatefulSets.
		Type string `yaml:"type"`
	} `yaml:"whitelisted_controllers"`

	// BlacklistedControllers prevent pod controllers which would otherwise be restarted
	// due to change in a managed resource they mount from being restarted by Order. It
	// is ineffective if whitelisted_controllers is set above.
	BlacklistedControllers []struct {
		// Name of the nominated pod controller
		Name string `yaml:"name"`

		// Namespace of the nominated pod controller
		Namespace string `yaml:"namespace"`

		// Type of the nominated pod controller, one of DaemonSets, Deployments, Jobs, or
		// StatefulSets.
		Type string `yaml:"type"`
	} `yaml:"blacklisted_controllers"`
}

// Parse populates parsed fields of the config which are derived from YAML values.
func (c *OrderConfig) Parse() error {
	// Nil config, it parses to nil
	if c == nil {
		return nil
	}

	// Parse controller resync duration
	controllerResyncDuration, err := getControllerResyncPeriod(c.ControllerResyncDuration)
	if err != nil {
		return err
	}
	c.XXXControllerResyncDuration = *controllerResyncDuration

	// Parse default cooldown duration
	restartCooldown, err := getRestartCooldownPeriod(c.RestartCooldown)
	if err != nil {
		return err
	}
	c.XXXParsedRestartCooldown = *restartCooldown

	// Parse pod controller stagger duration
	podControllerStagger, err := getPodControllerStaggerPeriod(c.PodControllerStagger)
	if err != nil {
		return err
	}
	c.XXXParsedPodControllerStagger = *podControllerStagger

	return nil
}
