package merger

import (
	"AID/solution/tempstorage"
	"context"
	log "github.com/sirupsen/logrus"
	"os"
	"sync"
)

// StartMerge run merge process
func StartMerge(ctx context.Context, ts *tempstorage.TempStorage, outputPath string, k int) error {

	for {
		if hasSingle, resultPath := ts.HasSingleStoredFile(); hasSingle {

			err := os.Rename(resultPath, outputPath)
			if err != nil {
				log.Fatalf("error in moving result to output path: %v", err)
			}
			break
		}

		err := ts.SetupNextLevel()
		if err != nil {
			log.Errorf("error on setting up next level of TempStorage: %v", err)
			return err
		}

		var wg sync.WaitGroup

		for {
			// read channels
			var rChs []<-chan string
			rChs, err = ts.GetNextReadChs(ctx, k)
			if err != nil {
				log.Errorf("error on getting next read channels of TempStorage: %v", err)
				return err
			}

			// all read files have been processed, should go to next level
			if len(rChs) == 0 {
				break
			}

			// store channel
			var sCh chan<- string
			sCh, err = ts.GetNextStoreCh(ctx, &wg)
			if err != nil {
				log.Errorf("error on getting next store channel of TempStorage: %v", err)
				return err
			}
			wg.Add(1)

			sh := newSourceHeap(rChs)

			for sh.Len() > 0 {
				select {
				case <-ctx.Done():
					log.Warning("Merger stopped before finishing its job")
					return nil
				default:
					var head string
					head, err = sh.getHead()
					if err != nil {
						log.Errorf("error on getting smallest string from min heap: %v", err)
						return err
					}

					// Write to store channel
					sCh <- head

					// Update source item and heap
					err = sh.updateHead()
					if err != nil {
						log.Errorf("error on updating head of min heap: %v", err)
						return err
					}
				}
			}

			close(sCh)
		}

		// Wait till all store processes become complete
		wg.Wait()
	}
	return nil
}