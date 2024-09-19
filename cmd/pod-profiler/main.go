package main

import (
	"log"
	"pod_profiler/pkg/api/profiler"
)

func main() {

	profiler, err := profiler.New()
	if err != nil {
		log.Default().Fatalf("error initialising config: %s", err.Error())
	}

	profiler.Config.VarDump()

	go profiler.Start()
	for err := range profiler.Errors {
		log.Printf("%s\n", err.Error())
	}

	log.Default().Println("Stopping capture")

}
