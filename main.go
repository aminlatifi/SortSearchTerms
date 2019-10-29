package main

import (
	"AID/solution/bundler"
	"AID/solution/tempstorage"
	"context"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"sync"

	"AID/solution/inputserializer"
	log "github.com/sirupsen/logrus"
)

var (
	inputPath       = flag.String("i", "inputserializer/testData/input", "input directory path")
	tempPath        = flag.String("t", "", "temporary storage path")
	outputPath      = flag.String("o", "out.txt", "result path")
	logPath         = flag.String("l", "", "log file path")
	isLogVerbose    = flag.Bool("v", false, "verbose mode")
	processorNumber = flag.Int("p", runtime.NumCPU(), "number of processor to use")
	k               = flag.Int("k", 4, "available memory")
)

func init() {
	flag.Parse()

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
	if *k < 2 {
		log.Fatalf("k cannot be less than 2")
		return
	}

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

	if *tempPath == "" {
		*tempPath, err = ioutil.TempDir("", "dir")
		if err != nil {
			log.Fatalf("error in creating temporary directory: %v", err)
			return
		}
	}

	b := bundler.GetNewBundler(*k)
	b.AddTransformFunc(bundler.SortTransform)

	bundlerCh := b.GetBundlerCh(ctx, readCh)

	ts, err := tempstorage.NewTempStorage(*tempPath)
	if err != nil {
		log.Fatalf("error in creating temporary storage: %v", err)
		return
	}

	var wg sync.WaitGroup
	var ch chan<- string
	for bundle := range bundlerCh {
		ch, err = ts.GetNextStoreCh(ctx, &wg)
		if err != nil {
			log.Fatal(err)
			return
		}

		wg.Add(1)

		go func(ch chan<- string, bundle []string) {
			for _, v := range bundle {
				ch <- v
			}
			close(ch)
		}(ch, bundle)
	}

	wg.Wait()

}
