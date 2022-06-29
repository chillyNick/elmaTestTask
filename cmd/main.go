package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	substring  = "Go"
	maxWorkers = 5
)

type wordCounter struct {
	source string
	number int
	err    error
}

func main() {
	urls := []string{
		"https://golang.org",
		"https://golang.org",
	}

	workersNum := maxWorkers
	if workersNum > len(urls) {
		workersNum = len(urls)
	}

	counters := make(chan wordCounter, workersNum)

	showWg := new(sync.WaitGroup)
	showWg.Add(1)
	go showResults(showWg, counters)

	urlsCh := make(chan string, len(urls))
	for _, url := range urls {
		urlsCh <- url
	}
	close(urlsCh)

	wg := new(sync.WaitGroup)
	wg.Add(workersNum)
	for i := 0; i < workersNum; i++ {
		go countWord(wg, urlsCh, counters)
	}

	wg.Wait()
	close(counters)

	showWg.Wait()
}

func countWord(wg *sync.WaitGroup, urls <-chan string, urlCounters chan<- wordCounter) {
	defer wg.Done()

	for url := range urls {
		counter := wordCounter{source: url}

		resp, err := http.Get(url)
		if err != nil {
			log.Println(err)
			counter.err = errors.New("get request failed")
			urlCounters <- counter

			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Println(err)
			counter.err = errors.New("get body failed")
			urlCounters <- counter

			continue
		}

		counter.number = strings.Count(string(body), substring)
		urlCounters <- counter
	}
}

func showResults(wg *sync.WaitGroup, counters <-chan wordCounter) {
	total := 0
	for counter := range counters {
		if counter.err != nil {
			fmt.Printf("%s: %s\n", counter.source, counter.err)
			continue
		}

		total += counter.number
		fmt.Printf("%s: %d\n", counter.source, counter.number)
	}

	fmt.Printf("%s: %d\n", "Total", total)
	wg.Done()
}
