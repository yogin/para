package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type ParaResult struct {
	Results []RunnerOutput
}

type RunnerOutput struct {
	Command       string
	Raw           string
	Json          map[string]interface{}
	ExecutionTime string
}

func main() {
	prettyPrintFlag := flag.Bool("pp", false, "Pretty print json output")
	commandFileFlag := flag.String("file", "", "Path to commands file")
	concurrentRunFlag := flag.Int("c", 10, "Maximum number of commands to run at the same time")
	flag.Parse()

	if *concurrentRunFlag < 1 {
		log.Fatalf("Maximum number of concurrent commands should be > 0")
	}

	commands := []string{}
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

func readFromFile(path string) []string {
	if len(path) == 0 {
		return []string{}
	}

	f, err := os.Open(path)
	defer f.Close()

	if err != nil {
		log.Fatalf("Failed reading from file %s", err)
	}

	return commandsFromBuffer(f)
}

func readFromStdin() []string {
	fi, err := os.Stdin.Stat()
	if err != nil {
		log.Fatalf("Failed getting stats from stdin: %s", err)
	}

	if fi.Mode()&os.ModeNamedPipe == 0 {
		// no piped data
		return []string{}
	}

	return commandsFromBuffer(os.Stdin)
}

func commandsFromBuffer(buffer io.Reader) []string {
	commands := []string{}

	stream := bufio.NewScanner(buffer)
	for stream.Scan() {
		cmd := strings.TrimSpace(stream.Text())

		if len(cmd) > 0 {
			commands = append(commands, cmd)
		}
	}

	return commands
}

func handler(commands []string, concurrent *int) ParaResult {
	n := len(commands)
	results := make(chan RunnerOutput, n)
	outputs := []RunnerOutput{}
	bucket := make(chan bool, *concurrent)

	var wg sync.WaitGroup
	wg.Add(n)

	for i := 0; i < n; i++ {
		bucket <- true

		go func(cmd string) {
			defer func() {
				wg.Done()
				<-bucket
			}()

			results <- runner(cmd)
		}(commands[i])
	}

	wg.Wait()

	for i := 0; i < n; i++ {
		outputs = append(outputs, <-results)
	}

	return ParaResult{Results: outputs}
}

func runner(cmd string) RunnerOutput {
	start := time.Now()
	out, _ := exec.Command("sh", "-c", cmd).CombinedOutput()
	elapsed := time.Since(start)

	var rawJson map[string]interface{}
	json.Unmarshal(out, &rawJson)

	return RunnerOutput{
		Command:       cmd,
		Raw:           string(out[:]),
		Json:          rawJson,
		ExecutionTime: fmt.Sprintf("%s", elapsed),
	}
}
