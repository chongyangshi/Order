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

	PodControllerTypeDaemonSets   = "DaemonSets"
	PodControllerTypeDeployments  = "Deployments"
	PodControllerTypeJobs         = "Jobs"
	PodControllerTypeStatefulSets = "StatefulSets"
)

// OrderConfig is a structue of system-wide and managed resource-specific
// configurations for Order.
type OrderConfig struct {
	// Version represents the config version of Order
	Version float64 `yaml:"version"`

	// Namespaces represents rules under which Order should eithr action on
	// or ignore a pod controller, depending on what namespace it lives in.
	Namespaces struct {
		// If Whitelist is set, Order will only perform rolling restarts on pod
		// controllers in the specific namespaces listed. This includes if you
		// want to allow Order to perform rolling restarts in kube-system
		// namespace, which is not recommended.
		Whitelist []string `yaml:"whitelist"`

		// If Blacklist is set, Order will only perform rolling restarts on pod
		// controllers in the specific namespaces listed. kube-system will be
		// blacklisted by default. If you want to allow Order to perform rolling
		// restarts in kube-system namespace, you need to declare this namespace
		// in Whitelist, which is not recommended.
		Blacklist []string `yaml:"blacklist"`
	} `yaml:"namespace"`
	XXXParsedNamespaces []string

	// ControllerResyncDuration is a Go duration which defines how frequently controllers
	// should do a full refresh to ensure that their state is up to date with what's in
	// cluster. This is a sanity check for eventual consistency in Kubernetes Controllers.
	// For small clusters, 1m+ is recommended. For large clusters, consider 5m+. This value
	// cannot be below 15s.
	ControllerResyncDuration    string `yaml:"controller_resync_duration"`
	XXXControllerResyncDuration time.Duration

	// DefaultRestartCooldown is a Go duration which defines at most how frequently can
	// Order restart a pod controller, in order to prevent thrashing. Only positive
	// durations are acceptable. For small clusters, 2m+ is recommended. For large clusters,
	// consider 5m+. This can be overridden by restart_cooldown fields in individual
	// managed resource configurations. This value cannot be below 30s.
	DefaultRestartCooldown          string `yaml:"default_restart_cooldown"`
	XXXParsedDefaultRestartCooldown time.Duration

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

		// RestartCooldown is a Go duration which defines how frequently a pod controller
		// can be restarted for this resource. This overrides global and resource defaults.
		RestartCooldown          string `yaml:"restart_cooldown"`
		XXXParsedRestartCooldown time.Duration
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
	// Order avoid performing actions on all controllers unless it is explicitly whitelisted
	// as AdditionalControllers.
	AvoidAllControllersUnlessWhitelisted bool `yaml:"avoid_all_controllers_unless_whitelisted"`

	// RestartCooldown is a Go duration which defines how frequently a pod controller can
	// be restarted for this resource. This overrides global defaults.
	RestartCooldown          string `yaml:"restart_cooldown"`
	XXXParsedRestartCooldown time.Duration
}

// Parse populates parsed fields of the config which are derived from YAML values.
func (c *OrderConfig) Parse() error {
	// Nil config, it parses to nil
	if c == nil {
		return nil
	}

	// Parse namespaces based on whitelist and blacklist
	if len(c.Namespaces.Whitelist) > 0 {
		// kube-system can be specified here if desired
		c.XXXParsedNamespaces = c.Namespaces.Whitelist
	} else {
		kubeSystemFound := false
		for _, ns := range c.Namespaces.Blacklist {
			if ns == "kube-system" {
				kubeSystemFound = true
				break
			}
		}

		// kube-system is blacklisted by default unless in whitelist mode
		if !kubeSystemFound {
			c.XXXParsedNamespaces = append(c.Namespaces.Blacklist, "kube-system")
		}
	}

	// Parse controller resync duration
	controllerResyncDuration, err := getControllerResyncPeriod(c.ControllerResyncDuration)
	if err != nil {
		return err
	}
	c.XXXControllerResyncDuration = *controllerResyncDuration

	// Parse default cooldown duration
	defaultRestartCooldown, err := getRestartCooldownPeriod(c.DefaultRestartCooldown)
	if err != nil {
		return err
	}
	c.XXXParsedDefaultRestartCooldown = *defaultRestartCooldown

	// Parse pod controller stagger duration
	podControllerStagger, err := getPodControllerStaggerPeriod(c.PodControllerStagger)
	if err != nil {
		return err
	}
	c.XXXParsedPodControllerStagger = *podControllerStagger

	// Parse additional controllers specified for managed resources
	for _, resource := range c.ManagedResources {
		if resource == nil {
			continue
		}

		// Parse restart cooldowns with global default
		for _, controller := range resource.AdditionalControllers {
			if controller.RestartCooldown == "" {
				controller.XXXParsedRestartCooldown = *defaultRestartCooldown
			}

			controllerRestartCooldown, err := getRestartCooldownPeriod(controller.RestartCooldown)
			if err != nil {
				return err
			}

			controller.XXXParsedRestartCooldown = *controllerRestartCooldown
		}
	}

	return nil
}
