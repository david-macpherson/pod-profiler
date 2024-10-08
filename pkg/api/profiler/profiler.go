package profiler

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"
	"pod_profiler/pkg/api/capture"
	"pod_profiler/pkg/api/config"
	kubernetesClient "pod_profiler/pkg/api/kubernetes-client"

	"github.com/fsnotify/fsnotify"
)

type Profiler struct {
	Config *config.Config

	// The client used to access kubernetes resources
	K8sClient *kubernetesClient.Client

	Errors chan error

	running chan bool

	restart chan bool

	captures []*capture.Capture
}

func New() (*Profiler, error) {

	config, err := config.Load(true)
	if err != nil {
		return nil, err
	}

	config.VarDump()

	// create a new k8s client
	K8sClient, err := newK8sClient(config.Namespace)
	if err != nil {
		return nil, err
	}

	return &Profiler{
		Config:    config,
		K8sClient: K8sClient,
		Errors:    make(chan error),
		running:   make(chan bool),
	}, nil

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
		kubernetesClient.CachedResource_StatefulSet,
	}

	log.Default().Println("Starting to sync the cache")

	// Start to sync the cache
	client.BuildAndSyncNamedspacedCache(namespace, cacheResources...)

	log.Default().Println("Cache sync complete")

	// Return the new client and nil error to indicate a success
	return client, nil
}

func (profiler *Profiler) Start() error {

	if _, err := os.Stat(profiler.Config.ResultsPath); os.IsNotExist(err) {
		err := os.Mkdir(profiler.Config.ResultsPath, os.ModePerm)
		if err != nil {
			return err
		}
	}

	go profiler.Config.OnConfigChange(profiler.OnConfigChange)

	go profiler.process()

	profiler.running <- true

	return nil
}

func (profiler *Profiler) process() {

	for {
		select {
		case running := <-profiler.running:

			if running {
				log.Default().Println("Starting captures")

				err := profiler.initialiseCaptures()
				if err != nil {
					profiler.Errors <- err
				}

				err = profiler.CreateCaptureList()
				if err != nil {
					profiler.Errors <- err
				}

				for _, capture := range profiler.captures {
					go capture.StartCapture()
				}

			} else {
				log.Default().Println("Stopping capture")
				for _, capture := range profiler.captures {
					capture.StopCapture()
				}

			}
		case <-profiler.Errors:
			return
		}

	}

}

func (profiler *Profiler) OnConfigChange(event fsnotify.Event) {

	newConfig, err := config.Load(false)
	if err != nil {
		log.Default().Fatalf("unable to load config: %s\n", err.Error())
	}

	*profiler.Config = *newConfig
	log.Default().Println("config file updated")

	profiler.running <- false

	profiler.Config.VarDump()

	profiler.running <- true

}

func (profiler *Profiler) initialiseCaptures() error {
	captures := []*capture.Capture{}

	for _, deployment := range profiler.Config.PodLabels {
		capture, err := capture.New(profiler.K8sClient, profiler.Config.ResultsPath, deployment)
		if err != nil {
			return err
		}

		captures = append(captures, capture)
	}

	profiler.captures = captures

	return nil
}

func (profiler *Profiler) CreateCaptureList() error {

	listOfFiles := []string{}

	files, err := os.ReadDir(profiler.Config.ResultsPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Name() != "index.json" {
			listOfFiles = append(listOfFiles, file.Name())
		}
	}

	bytes, err := json.Marshal(listOfFiles)
	if err != nil {
		return err
	}

	indexFile, err := os.OpenFile(path.Join(profiler.Config.ResultsPath, "index.json"), os.O_WRONLY|os.O_CREATE, 0777)
	if err != nil {
		return err
	}

	_, err = indexFile.Write(bytes)
	if err != nil {
		return err
	}

	err = indexFile.Close()
	if err != nil {
		return err
	}
	return nil

}
