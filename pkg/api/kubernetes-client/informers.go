package kubernetesclient

import (
	informersappsv1 "k8s.io/client-go/informers/apps/v1"
	informersautoscalingv2 "k8s.io/client-go/informers/autoscaling/v2"
	informersautoscalingv2beta2 "k8s.io/client-go/informers/autoscaling/v2beta2"
	informersbatchv1 "k8s.io/client-go/informers/batch/v1"
	informersv1 "k8s.io/client-go/informers/core/v1"
	v1 "k8s.io/client-go/informers/networking/v1"
)

// Informers for the requested resources when BuildAndSyncNamedspacedCache is called
type Informers struct {
	Pod                            informersv1.PodInformer
	ReplicaController              informersv1.ReplicationControllerInformer
	ConfigMap                      informersv1.ConfigMapInformer
	Ingress                        v1.IngressInformer
	Deployment                     informersappsv1.DeploymentInformer
	StatefulSet                    informersappsv1.StatefulSetInformer
	DaemonSet                      informersappsv1.DaemonSetInformer
	Job                            informersbatchv1.JobInformer
	Service                        informersv1.ServiceInformer
	Secret                         informersv1.SecretInformer
	HorizontalPodAutoscalerV2beta2 informersautoscalingv2beta2.HorizontalPodAutoscalerInformer
	HorizontalPodAutoscalerV2      informersautoscalingv2.HorizontalPodAutoscalerInformer
	PersistentVolumeClaim          informersv1.PersistentVolumeClaimInformer
}
