package main

import (
	"context"
	"log"
	"pod_profiler/pkg/api/config"
	"time"

	v1Meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {

	config, err := config.NewConfig()
	if err != nil {
		log.Default().Fatalf("error initialising config: %s", err.Error())
	}

	// Print our configuration values
	log.Default().Println("")
	log.Default().Println("configuration")
	log.Default().Println("-------------------------------")
	log.Default().Println("")
	log.Default().Printf("namespace:  %s\n", config.Namespace)
	log.Default().Println("-------------------------------")

	podName := "sps-operator-78687b7c6b-mt2zj"

	for {
		data, err := config.K8sClient.Metrics.Pod().Get(context.Background(), podName, v1Meta.GetOptions{})
		if err != nil {
			log.Default().Printf("error: %s\n", err.Error())
		}

		for _, container := range data.Containers {
			log.Default().Printf("container: %s\n", container.Name)

			for resource, value := range container.Usage {
				log.Default().Printf("\tMilliValue %s: %v\n", resource, value.MilliValue())
			}

		}

		log.Default().Println("")
		time.Sleep(1 * time.Second)
	}

	//kubectl get --raw /apis/metrics.k8s.io/v1beta1/namespaces/sps-dave/pods/sps-operator-78687b7c6b-mt2zj

}
