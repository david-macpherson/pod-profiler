package kubernetesclient

import (
	"context"
	"errors"
	"fmt"
	"os"
	"pod_profiler/pkg/api/defaults/cloud"
	"reflect"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	v1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	v1batch "k8s.io/client-go/kubernetes/typed/batch/v1"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// The path to look for the config if we're not on cluster
var ConfigPath string = "/home/dan/.kube/config"

// flag to say whether or not we wish to override using on-cluster config
var UseInClusterConfig bool = true

// struct that extends client.Client interface
type Client struct {

	// the kubernetes client interface
	client.Client

	// Clientset (clientgo)
	Clientset kubernetes.Interface

	// Cache stores listers and informers for the requested resources when BuildAndSyncNamedspacedCache is called
	Cache *Cache

	// Our informer factories used to create our informers.
	SharedInformerFactory informers.SharedInformerFactory

	config *rest.Config

	// The kubernetes metrics
	Metrics *Metrics
}

// Create a new client wrapper around client.Client
func NewClient(...informers.SharedInformerOption) (*Client, error) {

	// register our application and version crds
	SchemeBuilder := &scheme.Builder{}

	s := runtime.NewScheme()
	clientgoscheme.AddToScheme(s)
	SchemeBuilder.AddToScheme(s)

	var config *rest.Config
	var err error

	// Check if we're currently running inside a k8s cluster
	_, inK8s := os.LookupEnv("KUBERNETES_SERVICE_HOST")

	// If not inside a k8s cluster, or we've explicitliy set to not use cluster config, look for our k8s config externally
	if !inK8s || !UseInClusterConfig {
		config, err = clientcmd.BuildConfigFromFlags("", ConfigPath)
		if err != nil {
			return nil, err
		}
	} else {
		// Use the k8s in cluster config that k8s creates for us
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	}

	client, err := client.New(config, client.Options{Scheme: s})
	if err != nil {
		return nil, err
	}

	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	spsclient := &Client{
		Client:    client,
		Clientset: clientset,
		config:    config,
	}

	return spsclient, nil
}

// Creates and syncs a new cache for the namespace provideded. It can take optional CachedResources that you can choose
// only build and sync for this instance of the client. If you choose only specific cached resources, it's important that you do not attempt to access
// any informers or listers for resources that have not been built and synced otherwise it will panic
func (c *Client) BuildAndSyncNamedspacedCache(namespace string, cachedResources ...CachedResource) error {

	// en sure we try and build something, we can't just have an empty cache
	if len(cachedResources) <= 0 {
		return errors.New("did not build and sync cache, no cached resources specified")
	}

	// create a new shared informer factory and build the informers we need for the REST API
	c.SharedInformerFactory = c.NewSharedInformerFactoryWithOptions(namespace)

	// create our cache object
	c.Cache = &Cache{
		Informers: &Informers{},
		Listers:   &Listers{},
		Namespace: namespace,
	}

	// this will store the informers that we wish to keep sync with
	toSync := make([]cache.InformerSynced, 0)

	// iterate over our cached resources and create our informers and listers accordingly.
	for _, cachedResource := range cachedResources {
		switch cachedResource {

		case CachedResource_Pod:
			c.Cache.Informers.Pod = c.SharedInformerFactory.Core().V1().Pods()
			c.Cache.Listers.Pod = c.Cache.Informers.Pod.Lister()
			podInformer := c.Cache.Informers.Pod.Informer()
			toSync = append(toSync, podInformer.HasSynced)

		case CachedResource_ReplicaController:
			c.Cache.Informers.ReplicaController = c.SharedInformerFactory.Core().V1().ReplicationControllers()
			c.Cache.Listers.ReplicaController = c.Cache.Informers.ReplicaController.Lister()
			replicaInformer := c.Cache.Informers.ReplicaController.Informer()
			toSync = append(toSync, replicaInformer.HasSynced)

		case CachedResource_ConfigMap:
			c.Cache.Informers.ConfigMap = c.SharedInformerFactory.Core().V1().ConfigMaps()
			c.Cache.Listers.ConfigMap = c.Cache.Informers.ConfigMap.Lister()
			configMapInformer := c.Cache.Informers.ConfigMap.Informer()
			toSync = append(toSync, configMapInformer.HasSynced)

		case CachedResource_Ingress:
			c.Cache.Informers.Ingress = c.SharedInformerFactory.Networking().V1().Ingresses()
			c.Cache.Listers.Ingress = c.Cache.Informers.Ingress.Lister()
			ingressInformer := c.Cache.Informers.Ingress.Informer()
			toSync = append(toSync, ingressInformer.HasSynced)

		case CachedResource_Deployment:
			c.Cache.Informers.Deployment = c.SharedInformerFactory.Apps().V1().Deployments()
			c.Cache.Listers.Deployment = c.Cache.Informers.Deployment.Lister()
			deploymentInformer := c.Cache.Informers.Deployment.Informer()
			toSync = append(toSync, deploymentInformer.HasSynced)

		case CachedResource_StatefulSet:
			c.Cache.Informers.StatefulSet = c.SharedInformerFactory.Apps().V1().StatefulSets()
			c.Cache.Listers.StatefulSet = c.Cache.Informers.StatefulSet.Lister()
			statefulsetInformer := c.Cache.Informers.StatefulSet.Informer()
			toSync = append(toSync, statefulsetInformer.HasSynced)

		case CachedResource_DaemonSet:
			c.Cache.Informers.DaemonSet = c.SharedInformerFactory.Apps().V1().DaemonSets()
			c.Cache.Listers.DaemonSet = c.Cache.Informers.DaemonSet.Lister()
			daemonSetInformer := c.Cache.Informers.DaemonSet.Informer()
			toSync = append(toSync, daemonSetInformer.HasSynced)

		case CachedResource_Job:
			c.Cache.Informers.Job = c.SharedInformerFactory.Batch().V1().Jobs()
			c.Cache.Listers.Job = c.Cache.Informers.Job.Lister()
			jobInformer := c.Cache.Informers.Job.Informer()
			toSync = append(toSync, jobInformer.HasSynced)

		case CachedResource_Service:
			c.Cache.Informers.Service = c.SharedInformerFactory.Core().V1().Services()
			c.Cache.Listers.Service = c.Cache.Informers.Service.Lister()
			serviceInformer := c.Cache.Informers.Service.Informer()
			toSync = append(toSync, serviceInformer.HasSynced)

		case CachedResource_Secret:
			c.Cache.Informers.Secret = c.SharedInformerFactory.Core().V1().Secrets()
			c.Cache.Listers.Secret = c.Cache.Informers.Secret.Lister()
			secretInformer := c.Cache.Informers.Secret.Informer()
			toSync = append(toSync, secretInformer.HasSynced)

		case CachedResource_HPA:

			// if we're deploying on CoreWeave, we need to use v2beta2 of the HPA resource
			// as they are using an old version of kubernetes (1.20 at the time of writing this)
			if cloud.PLATFORM == cloud.CloudPlatformType_CoreWeave {
				c.Cache.Informers.HorizontalPodAutoscalerV2beta2 = c.SharedInformerFactory.Autoscaling().V2beta2().HorizontalPodAutoscalers()
				c.Cache.Listers.HorizontalPodAutoscalerV2beta2 = c.Cache.Informers.HorizontalPodAutoscalerV2beta2.Lister()
				hpaInformer := c.Cache.Informers.HorizontalPodAutoscalerV2beta2.Informer()
				toSync = append(toSync, hpaInformer.HasSynced)
			} else {
				c.Cache.Informers.HorizontalPodAutoscalerV2 = c.SharedInformerFactory.Autoscaling().V2().HorizontalPodAutoscalers()
				c.Cache.Listers.HorizontalPodAutoscalerV2 = c.Cache.Informers.HorizontalPodAutoscalerV2.Lister()
				hpaInformer := c.Cache.Informers.HorizontalPodAutoscalerV2.Informer()
				toSync = append(toSync, hpaInformer.HasSynced)
			}

		case CachedResource_PVC:
			c.Cache.Informers.PersistentVolumeClaim = c.SharedInformerFactory.Core().V1().PersistentVolumeClaims()
			c.Cache.Listers.PersistentVolumeClaim = c.Cache.Informers.PersistentVolumeClaim.Lister()
			pvcInformer := c.Cache.Informers.PersistentVolumeClaim.Informer()
			toSync = append(toSync, pvcInformer.HasSynced)
		}
	}

	informerStopper := make(chan struct{})
	defer utilRuntime.HandleCrash()

	// we want the informer to run for the duration of the restapi
	// the minute we send a stop signal to this informer, it will no longer inform us of changes to the watched resources
	go c.SharedInformerFactory.Start(informerStopper)

	// wait for our informer cache to sync
	if !cache.WaitForCacheSync(informerStopper, toSync...) {
		utilRuntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	// initialise metrics
	err := c.initialiseMetrics()
	if err != nil {
		return err
	}

	return nil
}

// Creates and returns a new NewSharedInformerFactory with the appropriate namespace
// This is used to create informers throughout our framework so we can utilise the cache where possible
// and save hits against the kubeapi server
func (c *Client) NewSharedInformerFactoryWithOptions(namespace string) informers.SharedInformerFactory {

	if c.SharedInformerFactory == nil {
		c.SharedInformerFactory = informers.NewSharedInformerFactoryWithOptions(c.Clientset, 0, informers.WithNamespace(namespace))
	}
	return c.SharedInformerFactory
}

// UpdateRetry If a conflict occurs during an update, we re-get the object and apply the update again.
// Default is to retry 10 times with 100ms intervals
func (c *Client) UpdateRetry(ctx context.Context, obj client.Object, updateFunc func() error, options ...client.UpdateOption) error {

	// get the object type so we can use the specific clientset API if defined
	objectTypeAsString := reflect.ValueOf(obj).Type().String()
	var err error
	for i := 0; i < 10; i++ {

		// call the update func, which will actually apply the changes to our object
		err = updateFunc()
		if err != nil {
			return err
		}

		// Use the clientset for a specific resource if we have it defined below, otherwise default to using the client
		// NOTE: This is a step moving towards using the clientset for everything, as it is what is used for our unit tests and cache
		switch objectTypeAsString {
		case "*v1.ConfigMap":
			_, err = c.Clientset.CoreV1().ConfigMaps(obj.GetNamespace()).Update(ctx, obj.(*corev1.ConfigMap), v1meta.UpdateOptions{})
		case "*v1.Deployment":
			_, err = c.Clientset.AppsV1().Deployments(obj.GetNamespace()).Update(ctx, obj.(*appsv1.Deployment), v1meta.UpdateOptions{})
		default:
			err = c.Client.Update(ctx, obj, options...)
		}

		// check if we have a conflict error
		if apierrors.IsConflict(err) {

			// if we have a conflict, re-get the resource from the cluster and attempt to update it again on the next iteration
			err = c.Client.Get(ctx, types.NamespacedName{Name: obj.GetName(), Namespace: obj.GetNamespace()}, obj)
			if err != nil {
				return err
			}

			time.Sleep(time.Millisecond * 100)
			continue
		}

		return err
	}

	// if we get here, we've failed after a number of attempts
	return errors.New("failed to update the object")
}

// Convenience function to return the clientset service interface
func (c *Client) Service() v1core.ServiceInterface {
	return c.Clientset.CoreV1().Services(c.Cache.Namespace)
}

// Convenience function to return the clientset job interface
func (c *Client) Job() v1batch.JobInterface {
	return c.Clientset.BatchV1().Jobs(c.Cache.Namespace)
}

// Convenience function to return the clientset pvc interface
func (c *Client) PVC() v1core.PersistentVolumeClaimInterface {
	return c.Clientset.CoreV1().PersistentVolumeClaims(c.Cache.Namespace)
}

// Convenience function to return the clientset secret interface
func (c *Client) Secret() v1core.SecretInterface {
	return c.Clientset.CoreV1().Secrets(c.Cache.Namespace)
}

// Convenience function to return the clientset pod interface
func (c *Client) Pod() v1core.PodInterface {
	return c.Clientset.CoreV1().Pods(c.Cache.Namespace)
}

// Convenience function to return the clientset config map interface
func (c *Client) ConfigMap() v1core.ConfigMapInterface {
	return c.Clientset.CoreV1().ConfigMaps(c.Cache.Namespace)
}

// Convenience function to return the clientset deployment interface
func (c *Client) Deployments() v1.DeploymentInterface {
	return c.Clientset.AppsV1().Deployments(c.Cache.Namespace)
}

// Initialise the metrics
func (c *Client) initialiseMetrics() error {

	// Create a new metrics client set
	clientSet, err := metrics.NewForConfig(c.config)
	if err != nil {
		return err
	}

	// Store the metrics as part of the client
	c.Metrics = &Metrics{
		clientSet: clientSet,
		namespace: c.Cache.Namespace,
	}

	// Return nil to indicate a success
	return nil
}
