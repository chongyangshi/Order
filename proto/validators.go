package proto

import (
	"fmt"
	"time"
)

const (
	controllerResyncSafetyLowerBound     = time.Second * 15
	restartColldownSafetyLowerBound      = time.Second * 30
	podControllerStaggerSafetyLowerBound = time.Second * 5
)

func validateManagedResourceType(t string) bool {
	switch t {
	case ManagedResourceTypeConfigMaps, ManagedResourceTypeSecrets:
		return true
	}

	return false
}

func validatePodControllerType(t string) bool {
	switch t {
	case PodControllerTypeDaemonSets,
		PodControllerTypeDeployments,
		PodControllerTypeJobs,
		PodControllerTypeStatefulSets:
		return true
	}

	return false
}

func getControllerResyncPeriod(d string) (*time.Duration, error) {
	if d == "" {
		defaultPeriod := controllerResyncSafetyLowerBound
		return &defaultPeriod, nil
	}

	t, err := time.ParseDuration(d)
	if err != nil {
		return nil, fmt.Errorf("Error parsing controller resync period %s", d)
	}

	if t < controllerResyncSafetyLowerBound {
		return nil, fmt.Errorf("Specified controller resync period %s below safety threshold %v", d, controllerResyncSafetyLowerBound)
	}

	return &t, nil
}

func getRestartCooldownPeriod(d string) (*time.Duration, error) {
	if d == "" {
		defaultPeriod := restartColldownSafetyLowerBound
		return &defaultPeriod, nil
	}

	t, err := time.ParseDuration(d)
	if err != nil {
		return nil, fmt.Errorf("Error parsing restart cooldown resync period %s", d)
	}

	if t < restartColldownSafetyLowerBound {
		return nil, fmt.Errorf("Specified restart cooldown period %s below safety threshold %v", d, restartColldownSafetyLowerBound)
	}

	return &t, nil
}

func getPodControllerStaggerPeriod(d string) (*time.Duration, error) {
	if d == "" {
		defaultPeriod := podControllerStaggerSafetyLowerBound
		return &defaultPeriod, nil
	}

	t, err := time.ParseDuration(d)
	if err != nil {
		return nil, fmt.Errorf("Error parsing pod controller stagger period %s", d)
	}

	if t < podControllerStaggerSafetyLowerBound {
		return nil, fmt.Errorf("Specified pod controller stagger period %s below safety threshold %v", d, podControllerStaggerSafetyLowerBound)
	}

	return &t, nil
}
