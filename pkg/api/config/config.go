package config

import (
	"fmt"
	"log"
	"os"
	kubernetesClient "pod_profiler/pkg/api/kubernetes-client"
)

// This holds the config for the entire application
type Config struct {

	// The namespace the application is running in
	Namespace string `yaml:"namespace"`

	// The client used to access kubernetes resources
	K8sClient *kubernetesClient.Client
}

func NewConfig() (*Config, error) {

	// Load the config
	config, err := load()
	if err != nil {
		return nil, err
	}

	log.Default().Printf("name: %s\n", config.Namespace)

	// create a new k8s client
	config.K8sClient, err = newK8sClient(config.Namespace)
	if err != nil {
		return nil, err
	}

	// Return the config and a nil error indicating a success
	return config, nil

}

func load() (*Config, error) {
	// Initialise an empty config
	config := &Config{}

	// Initialise a new viper
	namespace, exists := os.LookupEnv("NAMESPACE")
	if !exists {
		return nil, fmt.Errorf("NAMESPACE env var does not exists")
	}

	if namespace == "" {
		return nil, fmt.Errorf("NAMESPACE can not be blank")
	}

	config.Namespace = namespace

	// Return the config and nil error to indicate a success
	return config, nil
}

func newK8sClient(namespace string) (*kubernetesClient.Client, error) {

	log.Default().Println("Create kubernetes client")

	// Create the kubernetes client
	client, err := kubernetesClient.NewClient()
	if err != nil {
		return nil, fmt.Errorf("error creating kubernetes client: %s", err.Error())
	}

	// Create a list of resources to cache
	cacheResources := []kubernetesClient.CachedResource{
		kubernetesClient.CachedResource_Pod,
	}

	log.Default().Println("Starting to sync the cache")

	// Start to sync the cache
	client.BuildAndSyncNamedspacedCache(namespace, cacheResources...)

	log.Default().Println("Cache sync complete")

	// Return the new client and nil error to indicate a success
	return client, nil
}
