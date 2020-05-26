package secrets

import (
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
)

// DaemonSetHasReference returns a boolean value on whether it references the Secret
// we are looking for.
func DaemonSetHasReference(ds *appsv1.DaemonSet, secret *corev1.Secret) bool {
	if ds == nil || secret == nil {
		return false
	}

	// If DaemonSet and Secret are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if ds.Namespace != secret.Namespace {
		return false
	}

	return podSpecHasReference(ds.Spec.Template.Spec, secret.Name)
}

// DeploymentHasReference returns a boolean value on whether it references the Secret
// we are looking for.
func DeploymentHasReference(deploy *appsv1.Deployment, secret *corev1.Secret) bool {
	if deploy == nil || secret == nil {
		return false
	}

	// If Deployment and Secret are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if deploy.Namespace != secret.Namespace {
		return false
	}

	return podSpecHasReference(deploy.Spec.Template.Spec, secret.Name)
}

// JobHasReference returns a boolean value on whether it references the Secret
// we are looking for.
func JobHasReference(job *batchv1.Job, secret *corev1.Secret) bool {
	if job == nil || secret == nil {
		return false
	}

	// If Job and Secret are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if job.Namespace != secret.Namespace {
		return false
	}

	return podSpecHasReference(job.Spec.Template.Spec, secret.Name)
}

// StatefulSetHasReference returns a boolean value on whether it references the Secret
// we are looking for.
func StatefulSetHasReference(sts *appsv1.StatefulSet, secret *corev1.Secret) bool {
	if sts == nil || secret == nil {
		return false
	}

	// If StatefulSet and Secret are in different namespaces, the former won't be
	// able to reference the latter, so it won't be what we are looking for.
	if sts.Namespace != secret.Namespace {
		return false
	}

	return podSpecHasReference(sts.Spec.Template.Spec, secret.Name)
}

func podSpecHasReference(podSpec corev1.PodSpec, name string) bool {
	// Check whether our Secret is mounted as a volume in the pod controller's pod template.
	for _, volume := range podSpec.Volumes {
		if volume.Secret == nil {
			continue
		}

		if volume.Secret.SecretName == name {
			return true
		}
	}

	// Check whether our Secret is referenced in the pod controller's container environment
	// variables.
	for _, container := range podSpec.Containers {
		for _, env := range container.Env {
			if env.ValueFrom == nil {
				continue
			}

			if env.ValueFrom.SecretKeyRef == nil {
				continue
			}

			if env.ValueFrom.SecretKeyRef.Name == name {
				return true
			}
		}
	}

	return false
}
