package processor

import (
	"context"

	"github.com/icydoge/Order/controllers/cachers"
)

// @TODO: Control loop
// - Pop from buffer
// - Check whether resource is managed
// - If resource is not managed, discard
// - If resource is managed, continue
// - Push back into buffer if minimum cooldown not yet met
// - Look up currently matching pod controllers from cachers
// - For each matching pod controller, if namespace does not qualify, skip
// - If namespace qualifies, check whether managed resource hash annotations match
// - If matches, skip as pod controller is already up to date
// - If does not match, check if minimum cooldown from last restart qualifies
// - If minimum cooldown does not qualifies, push back into buffer and skip
// - If minimum cooldown qualifies, perform rolling restart and continue.

// In a control loop, we validate all pod controllers against the versions of managed
// resources they run. It is unnecessary to use locking and keep caches in a consistent
// state while we process them, as it will simply be covered in the next loop under
// an eventually consistent model.
func controlLoop(ctx context.Context) error {
	// Retrieve DaemonSets currently in cache matching target namespaces
	daemonSets, err := cachers.GetDaemonSets()
	if err != nil {
		return err
	}

	// Retrieve Deployments currently in cache matching target namespaces
	deployments, err := cachers.GetDeployments()
	if err != nil {
		return err
	}

	// Retrieve jobs currently in cache matching target namespaces
	jobs, err := cachers.GetJobs()
	if err != nil {
		return err
	}

	// Retrieve StatefulSets currently in cache matching target namespaces
	statefulSets, err := cachers.GetStatefulSets()
	if err != nil {
		return err
	}

	// Retrieve state of managed resources
	managedResources, err := getManagedResourcesInConfig()
	if err != nil {
		return err
	}

	// For each managed resource, compute what pod controllers currently in cache
	// they apply to.
	for _, resource := range managedResources {
	}

}
