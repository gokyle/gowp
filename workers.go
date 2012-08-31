package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var num_workers = 3

func worker(urls <-chan string, out chan<- error, id int) {
	fmt.Println("[*] starting worker: ", id)
	for url := range urls {
		fmt.Printf("[+] %d fetching %s\n", id, url)
		res, err := http.Get(url)
		if err != nil {
			fmt.Printf("[!] %d http connection failed\n", id)
			out <- err
			continue
		}
		defer res.Body.Close()

		r, err := ioutil.ReadAll(res.Body)
		if err != nil {
		} else if len(r) > 0 {
			fmt.Printf("[+] %d connection succeeded\n", id)
			err = nil
		} else {
			fmt.Printf("[!] %d zero read length\n", id)
			err = fmt.Errorf("zero read length")
		}

		out <- err
	}
}

func pool(urls []string) int {
	workers_in := make(chan string, 2*num_workers)
	workers_out := make(chan error, num_workers)

        // launch workers
	for i := 0; i < num_workers; i++ {
		go worker(workers_in, workers_out, i)
	}

        // load urls and check for errors
	go load_urls(workers_in, urls)
	return scan_for_errors(workers_out, len(urls))
}

func load_urls(work_in chan<- string, urls []string) {
	for _, url := range urls {
		work_in <- url
	}
}

func scan_for_errors(errs <-chan error, tasks int) int {
	results := 0
	nerrs := 0
	for err := range errs {
		if err != nil {
			fmt.Printf("[!] failed to retrieve a url: %s!\n", err)
			nerrs++
		} else {
			fmt.Println("[+] worker returned successfully")
		}
		results++

		if results == tasks {
			break
		}
	}

	return nerrs
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("please pass a list of urls")
		os.Exit(1)
	}

	fmt.Println("[+] kicking off worker pool")
	errors := pool(os.Args[1:])
	fmt.Printf("[+] worker pool complete (%d errors)\n", errors)
}
