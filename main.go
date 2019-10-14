package main

import (
	"context"
	"fmt"
	"log"

	"AID/solution/inputserializer"
)

var inputPath string = "inputserializer/testData/input"

func main() {
	var inputSerializer inputserializer.InputSerializer

	// Stop whole subprocesses in case of exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use File Serializer to read directory files' content
	inputSerializer = inputserializer.NewFileSerializer(inputPath)
	reacCh, err := inputSerializer.GetSeralizerCh(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Simple output
	for v := range reacCh {
		fmt.Println(v)
	}
}
