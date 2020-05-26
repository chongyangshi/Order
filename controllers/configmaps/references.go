package configmaps

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// DaemonSetHasReference returns a boolean value on whether it references the ConfigMap
// we are looking for.
func DaemonSetHasReference(ds *appsv1.DaemonSet, cm *corev1.ConfigMap) bool {
	if ds == nil || cm == nil {
		return false
	}

	// If DaemonSet and ConfigMap are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if ds.Namespace != cm.Namespace {
		return false
	}

	return podSpecHasReference(ds.Spec.Template.Spec, cm.Name)
}

// DeploymentHasReference returns a boolean value on whether it references the ConfigMap
// we are looking for.
func DeploymentHasReference(deploy *appsv1.Deployment, cm *corev1.ConfigMap) bool {
	if deploy == nil || cm == nil {
		return false
	}

	// If Deployment and ConfigMap are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if deploy.Namespace != cm.Namespace {
		return false
	}

	return podSpecHasReference(deploy.Spec.Template.Spec, cm.Name)
}

// JobHasReference returns a boolean value on whether it references the ConfigMap
// we are looking for.
func JobHasReference(job *batchv1.Job, cm *corev1.ConfigMap) bool {
	if job == nil || cm == nil {
		return false
	}

	// If Job and ConfigMap are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if job.Namespace != cm.Namespace {
		return false
	}

	return podSpecHasReference(job.Spec.Template.Spec, cm.Name)
}

// StatefulSetHasReference returns a boolean value on whether it references the ConfigMap
// we are looking for.
func StatefulSetHasReference(sts *appsv1.StatefulSet, cm *corev1.ConfigMap) bool {
	if sts == nil || cm == nil {
		return false
	}

	// If StatefulSet and ConfigMap are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if sts.Namespace != cm.Namespace {
		return false
	}

	return podSpecHasReference(sts.Spec.Template.Spec, cm.Name)
}

func podSpecHasReference(podSpec corev1.PodSpec, name string) bool {
	// Check whether our ConfigMap is mounted as a volume in the pod controller's pod template.
	for _, volume := range podSpec.Volumes {
		if volume.ConfigMap == nil {
			continue
		}

		if volume.ConfigMap.Name == name {
			return true
		}
	}

	// Check whether our ConfigMap is referenced in the pod controller's container environment
	// variables.
	for _, container := range podSpec.Containers {
		for _, env := range container.Env {
			if env.ValueFrom == nil {
				continue
			}

			if env.ValueFrom.ConfigMapKeyRef == nil {
				continue
			}

			if env.ValueFrom.ConfigMapKeyRef.Name == name {
				return true
			}
		}
	}

	return false
}
