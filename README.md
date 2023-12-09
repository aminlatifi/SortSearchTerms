# SortSearchTerms
The application sorts up to 1 TB of input log files stored on a single hard drive.

## Design
![Design 001](https://github.com/aminlatifi/SortSearchTerms/assets/5684607/08a65e12-99d4-4986-bcf1-961589292f70)


## Run
**NOTE:** Built by go 1.13



### Build

```bash
go build
```

### Usage

Command useful flags:

```
  -i string
    	input directory path (default "inputserializer/testData/input")
  -k int
    	available memory (default 4)
  -l string
    	log file path
  -n int
    	limit number of open files (default 5000)
  -o string
    	result path (default "out.txt")
  -p int
    	number of processor to use (default 8)
  -t string
    	temporary storage path
  -v	verbose mode
```



### Example

```sh
./solution -i /tmp/words -k 10000000 -n 5000  -o /tmp/output/out.txt -t /tmp/tmpDir
```
***I sorted 1GB of text file by above command in less than 3 minutes on my own machine***
