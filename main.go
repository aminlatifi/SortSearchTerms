package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"runtime"

	"AID/solution/inputserializer"
	log "github.com/sirupsen/logrus"
)

var (
	inputPath       = flag.String("i", "inputserializer/testData/input", "input directory path")
	logPath         = flag.String("l", "", "log file path")
	isLogVerbose    = flag.Bool("v", false, "verbose mode")
	processorNumber = flag.Int("p", runtime.NumCPU(), "number of processor to use")
)

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	if *isLogVerbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	if *logPath != "" {
		lf, err := os.OpenFile(*logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			log.Errorf("Unable to open log file for writing: %s", err)
		} else {
			log.SetOutput(io.MultiWriter(lf, os.Stdout))
		}
	}

	runtime.GOMAXPROCS(*processorNumber)
}

func main() {
	var inputSerializer inputserializer.InputSerializer

	// Stop whole sub processes in case of exit
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log.Infof("Read input from directory: %s", *inputPath)
	// Use File Serializer to read directory files' content
	inputSerializer = inputserializer.NewDirSerializer(*inputPath)
	readCh, err := inputSerializer.GetSerializerCh(ctx)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Simple output
	for v := range readCh {
		fmt.Println(v)
	}
}
