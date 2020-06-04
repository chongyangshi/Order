package processor

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"

	"github.com/icydoge/Order/config"
	"github.com/icydoge/Order/logging"
	"github.com/icydoge/Order/proto"
)

func isDaemonSetInScope(ds *appsv1.DaemonSet) bool {
	if ds == nil {
		return false
	}

	return isPodControllerInScope(ds.Name, ds.Namespace, proto.PodControllerTypeDaemonSets)
}

func isDeploymentInScope(deploy *appsv1.Deployment) bool {
	if deploy == nil {
		return false
	}

	return isPodControllerInScope(deploy.Name, deploy.Namespace, proto.PodControllerTypeDeployments)
}

func isJobInScope(job *batchv1.Job) bool {
	if job == nil {
		return false
	}

	return isPodControllerInScope(job.Name, job.Namespace, proto.PodControllerTypeJobs)
}

func isStatefulSetInScope(sts *appsv1.StatefulSet) bool {
	if sts == nil {
		return false
	}

	return isPodControllerInScope(sts.Name, sts.Namespace, proto.PodControllerTypeStatefulSets)
}

func isPodControllerInScope(name, namespace, controllerType string) bool {
	if config.Config == nil {
		logging.Fatal("Error: pod controllers in config unexpectedly requested before config is parsed")
	}

	// First check if pod controller's namespace is allowed for changes
	// By default this is applied to all namespaces except kube-system,
	// unless specified.
	namespaceMatched := false
	for _, ns := range config.Config.XXXParsedNamespaces {
		if ns == namespace {
			namespaceMatched = true
			break
		}
	}

	if len(config.Config.XXXParsedNamespaces) > 0 && !namespaceMatched {
		return false
	}

	var matchedResources []*proto.ManagedResource
	for _, resource := range config.Config.ManagedResources {
		if resource == nil {
			// Should never happen
			continue
		}

		// Check if controller is avoided by rule for that resource
		controllerAvoided := false
		for _, avoid := range resource.AvoidControllers {
			if avoid.Type == controllerType && avoid.Name == name && avoid.Namespace == namespace {
				controllerAvoided = true
				break
			}
		}

		if controllerAvoided {
			logging.Debug("Ignoring controller for %s (%s) of type %s due to avoid rule in %v", name, namespace, controllerType, resource)
			continue
		}

		podInWhitelist := false
		for _, item := range resource.AdditionalControllers {
			if item.Type == controllerType && item.Name == name && item.Namespace == namespace {
				podInWhitelist = true
			}
		}

		switch {
		case podInWhitelist:
			// If specified in additional controllers, and not in avoid controllers earlier,
			// this pod controller is in scope.
			matchedResources = append(matchedResources, resource)
		case resource.AvoidAllControllersUnlessWhitelisted && !podInWhitelist:
			logging.Debug("Ignoring controller for %s (%s) of type %s due to whitelist mode on for %v", name, namespace, controllerType, resource)
		default:
			// Whitelist mode not on for resource, this pod controller is in scope by default
			matchedResources = append(matchedResources, resource)
		}
	}

	// If controller is in a global target namespace and matches on at least one managed
	// resource, it is an in-scope pod controller. This doesn't necessarily means it needs
	// to be restarted, it depends on whether all its resource versions match the up-to-date
	// version.
	return len(matchedResources) > 0
}
