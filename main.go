package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
)

type ParaResult struct {
	Results []*Runner
}

func main() {
	prettyPrintFlag := flag.Bool("pp", false, "Pretty print json output")
	commandFileFlag := flag.String("file", "", "Path to commands file")
	concurrentRunFlag := flag.Int("c", 10, "Maximum number of commands to run at the same time")
	flag.Parse()

	if *concurrentRunFlag < 1 {
		log.Fatalf("Maximum number of concurrent commands should be > 0")
	}

	commands := []*Runner{}
	commands = append(commands, readFromStdin()...)
	commands = append(commands, readFromFile(*commandFileFlag)...)

	results := handler(commands, concurrentRunFlag)
	render(results, *prettyPrintFlag)
}

func render(results ParaResult, pp bool) {
	var out []byte
	var err error

	if pp {
		out, err = json.MarshalIndent(results, "", "  ")
	} else {
		out, err = json.Marshal(results)
	}

	if err != nil {
		log.Fatalf("JSON Marshal failed: %s", err)
	}

	fmt.Printf("%s\n", string(out))
}

func readFromFile(path string) []*Runner {
	if len(path) == 0 {
		return []*Runner{}
	}

	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		log.Fatalf("Failed reading from file %s", err)
	}

	return NewRunnersFromBuffer(f)
}

func readFromStdin() []*Runner {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Failed getting stats from stdin: %s", err)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		// no piped data
		return []*Runner{}
	}

	return NewRunnersFromBuffer(os.Stdin)
}

func handler(runners []*Runner, concurrent *int) ParaResult {
	bucket := make(chan bool, *concurrent)

	var wg sync.WaitGroup
	wg.Add(len(runners))

	for i := range runners {
		bucket <- true

		go func(r *Runner) {
			defer func() {
				wg.Done()
				<-bucket
			}()

			r.Run()
		}(runners[i])
	}

	wg.Wait()

	return ParaResult{Results: runners}
}
