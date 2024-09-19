package kubernetesclient

// CachedResource type, this is basically an enum that can be used to tell the client
// which resources we want to cache. The main reason we need to cache only the resources that are relevant is
// due to RBAC permissions of the consuming code. It also helps to be more efficient only using what we need
type CachedResource string

var CachedResource_Pod CachedResource = "Pod"
var CachedResource_ReplicaController CachedResource = "ReplicaController"
var CachedResource_ConfigMap CachedResource = "ConfigMap"
var CachedResource_Ingress CachedResource = "Ingress"
var CachedResource_Deployment CachedResource = "Deployment"
var CachedResource_StatefulSet CachedResource = "StatefulSet"
var CachedResource_DaemonSet CachedResource = "DaemonSet"
var CachedResource_Job CachedResource = "Job"
var CachedResource_Service CachedResource = "Service"
var CachedResource_Secret CachedResource = "Secret"
var CachedResource_HPA CachedResource = "HorizontalPodAutoscaler"
var CachedResource_PVC CachedResource = "PersistentVolumeClaim"

// returns true if the given cached resource is in an array of cached resources or false otherwise
func (cr CachedResource) In(cachedResources []CachedResource) bool {
	for _, v := range cachedResources {
		if cr == v {
			return true
		}
	}
	return false
}
