package capture

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"pod_profiler/pkg/api/defaults"
	kubernetesClient "pod_profiler/pkg/api/kubernetes-client"
	"strconv"
	"time"

	v1Core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1Meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type Capture struct {
	client      *kubernetesClient.Client
	csvFile     *os.File
	resultsPath string
	Deployment  string `json:"deployment"`
	OnRecord    chan Record
	Errors      chan error
	running     chan bool
}

type Record struct {
	DateStamp time.Time `json:"datestamp"`
	Pod       Pod       `json:"pod"`
}

type Pod struct {
	Name       string      `json:"name"`
	Containers []Container `json:"containers"`
}

type Container struct {
	Name   string `csv:"name"`
	Cpu    int64  `csv:"cpu"`
	Memory int64  `csv:"memory"`
}

func New(client *kubernetesClient.Client, resultsPath, deploymentName string) (*Capture, error) {

	if deploymentName == "" {
		return nil, fmt.Errorf("deployment name can not be blank")

	}
	if client == nil {
		return nil, fmt.Errorf("kubernetes client can not be nil")
	}

	capture := &Capture{
		client:      client,
		Deployment:  deploymentName,
		resultsPath: resultsPath,
		OnRecord:    make(chan Record),
		Errors:      make(chan error),
		running:     make(chan bool),
	}

	pods, err := capture.GetPods()
	if err != nil {
		capture.Errors <- err
	}

	for _, pod := range pods {
		err := capture.createFile(pod.GetName())
		if err != nil {
			return nil, err
		}
	}

	return capture, nil

}

func (capture *Capture) GetPods() ([]*v1Core.Pod, error) {
	label, err := labels.Parse(defaults.KUBERNETES_NAME_LABEL + "=" + capture.Deployment)
	if err != nil {
		return nil, err
	}

	pods, err := capture.client.Cache.Pod().List(label)
	if err != nil {
		return nil, err
	}

	return pods, nil

}

func (capture *Capture) StartCapture() {

	log.Default().Printf("Starting capture of %s\n", capture.Deployment)

	go capture.process()

	capture.running <- true
}

func (capture *Capture) process() {

	for {
		select {

		case running := <-capture.running:

			if running {
				pods, err := capture.GetPods()
				if err != nil {
					capture.Errors <- err
				}

				for _, pod := range pods {
					go capture.startContainerCapture(pod)
				}
			} else {
				err := capture.csvFile.Close()
				if err != nil {
					log.Default().Printf("error: %s\n", err.Error())
					capture.Errors <- err
				}
				return
			}

		case record := <-capture.OnRecord:
			log.Default().Printf("on record %s\n", record.Pod.Name)
			err := capture.saveRecord(record)
			if err != nil {
				log.Default().Printf("error: %s\n", err.Error())
				capture.Errors <- err
			}

		case err := <-capture.Errors:
			log.Default().Printf("error: %s", err.Error())
			capture.Errors <- err
		}
	}
}

func (capture *Capture) StopCapture() {
	capture.running <- false
}

func (capture *Capture) startContainerCapture(pod *v1Core.Pod) error {

	var lastCapture v1Meta.Time

	for {

		data, err := capture.client.Metrics.Pod().Get(context.Background(), pod.GetName(), v1Meta.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				time.Sleep(10 * time.Second)
				continue
			}
			capture.Errors <- err
		}

		if data.Timestamp.Time == lastCapture.Time {
			time.Sleep(10 * time.Second)
			continue
		}

		lastCapture = data.Timestamp

		record := Record{
			DateStamp: data.Timestamp.Time,
			Pod: Pod{
				Name: pod.GetName(),
			},
		}

		for _, container := range data.Containers {
			containerResult := Container{
				Name:   container.Name,
				Cpu:    container.Usage.Cpu().MilliValue(),
				Memory: container.Usage.Memory().Value(),
			}

			record.Pod.Containers = append(record.Pod.Containers, containerResult)
		}
		capture.OnRecord <- record
	}
}

func (capture *Capture) createFile(podName string) error {

	filename := fmt.Sprintf("%s/%s.csv", capture.resultsPath, podName)
	flags := os.O_APPEND | os.O_WRONLY

	if _, err := os.Stat(filename); os.IsNotExist(err) {
		flags = os.O_APPEND | os.O_WRONLY | os.O_CREATE
	}

	var err error
	capture.csvFile, err = os.OpenFile(filename, flags, 0777)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(capture.csvFile)
	defer writer.Flush()

	if flags == os.O_APPEND|os.O_WRONLY|os.O_CREATE {
		headerRow := []string{"time", "name", "cpu", "memory"}
		err = writer.Write(headerRow)
		if err != nil {
			return err
		}
	}

	return nil

}

func (capture *Capture) saveRecord(record Record) error {

	writer := csv.NewWriter(capture.csvFile)
	defer writer.Flush()

	for _, container := range record.Pod.Containers {
		row := []string{
			time.Now().Format(time.TimeOnly),
			container.Name,
			strconv.FormatInt(container.Cpu, 10),
			strconv.FormatInt(container.Memory, 10),
		}

		err := writer.Write(row)
		if err != nil {
			return err
		}
	}

	return nil

}
