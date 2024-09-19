package kubernetesclient

import (
	listersappsv1 "k8s.io/client-go/listers/apps/v1"
	listersautoscalingv2 "k8s.io/client-go/listers/autoscaling/v2"
	listersautoscalingv2beta2 "k8s.io/client-go/listers/autoscaling/v2beta2"
	listersbatchv1 "k8s.io/client-go/listers/batch/v1"
	listersv1 "k8s.io/client-go/listers/core/v1"
	listersnetworkingv1 "k8s.io/client-go/listers/networking/v1"
)

// Listers for the requested resources when BuildAndSyncNamedspacedCache is called
type Listers struct {
	Pod                            listersv1.PodLister
	ReplicaController              listersv1.ReplicationControllerLister
	ConfigMap                      listersv1.ConfigMapLister
	Ingress                        listersnetworkingv1.IngressLister
	Deployment                     listersappsv1.DeploymentLister
	StatefulSet                    listersappsv1.StatefulSetLister
	DaemonSet                      listersappsv1.DaemonSetLister
	Job                            listersbatchv1.JobLister
	Service                        listersv1.ServiceLister
	Secret                         listersv1.SecretLister
	HorizontalPodAutoscalerV2beta2 listersautoscalingv2beta2.HorizontalPodAutoscalerLister
	HorizontalPodAutoscalerV2      listersautoscalingv2.HorizontalPodAutoscalerLister
	PersistentVolumeClaim          listersv1.PersistentVolumeClaimLister
}
