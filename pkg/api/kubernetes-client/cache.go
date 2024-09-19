package kubernetesclient

import (
	v1apps "k8s.io/client-go/listers/apps/v1"
	v2autoscaling "k8s.io/client-go/listers/autoscaling/v2"
	v2beta2autoscaling "k8s.io/client-go/listers/autoscaling/v2beta2"
	v1batch "k8s.io/client-go/listers/batch/v1"
	v1core "k8s.io/client-go/listers/core/v1"
	v1networking "k8s.io/client-go/listers/networking/v1"
)

// Cache stores listers and informers for the requested resources when BuildAndSyncNamedspacedCache is called
type Cache struct {
	Informers *Informers
	Listers   *Listers
	Namespace string
}

// Convenience function to return the lister for deployment on the cached namespace
func (cache *Cache) Deployment() v1apps.DeploymentNamespaceLister {
	return cache.Listers.Deployment.Deployments(cache.Namespace)
}

// Convenience function to return the lister for config maps on the cached namespace
func (cache *Cache) ConfigMap() v1core.ConfigMapNamespaceLister {
	return cache.Listers.ConfigMap.ConfigMaps(cache.Namespace)
}

// Convenience function to return the lister for v2 HPA on the cached namespace
func (cache *Cache) HPAv2() v2autoscaling.HorizontalPodAutoscalerNamespaceLister {
	return cache.Listers.HorizontalPodAutoscalerV2.HorizontalPodAutoscalers(cache.Namespace)
}

// Convenience function to return the lister for v2beta2 HPA on the cached namespace
func (cache *Cache) HPAv2beta2() v2beta2autoscaling.HorizontalPodAutoscalerNamespaceLister {
	return cache.Listers.HorizontalPodAutoscalerV2beta2.HorizontalPodAutoscalers(cache.Namespace)
}

// Convenience function to return the lister for ingress on the cached namespace
func (cache *Cache) Ingress() v1networking.IngressNamespaceLister {
	return cache.Listers.Ingress.Ingresses(cache.Namespace)
}

// Convenience function to return the lister for jobs on the cached namespace
func (cache *Cache) Job() v1batch.JobNamespaceLister {
	return cache.Listers.Job.Jobs(cache.Namespace)
}

// Convenience function to return the lister for pods on the cached namespace
func (cache *Cache) Pod() v1core.PodNamespaceLister {
	return cache.Listers.Pod.Pods(cache.Namespace)
}

// Convenience function to return the lister for secrets on the cached namespace
func (cache *Cache) Secret() v1core.SecretNamespaceLister {
	return cache.Listers.Secret.Secrets(cache.Namespace)
}

// Convenience function to return the lister for pvc on the cached namespace
func (cache *Cache) PersistentVolumeClaim() v1core.PersistentVolumeClaimNamespaceLister {
	return cache.Listers.PersistentVolumeClaim.PersistentVolumeClaims(cache.Namespace)
}
