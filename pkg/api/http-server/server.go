package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"pod_profiler/pkg/api/config"
)

type HttpServer struct {
	config  *config.Config
	running bool
}

func New(config *config.Config) (*HttpServer, error) {

	return &HttpServer{
		config:  config,
		running: true,
	}, nil
}

func (httpServer *HttpServer) Start() error {
	fs := http.FileServer(http.Dir(httpServer.config.ResultsPath))
	http.Handle("/", fs)

	log.Default().Printf("Starting to listen on port: %v\n", httpServer.config.HttpPort)
	err := http.ListenAndServe(fmt.Sprintf(":%v", httpServer.config.HttpPort), nil)
	if err != nil {
		return err
	}
	return nil
}
