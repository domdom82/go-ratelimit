package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

var limit int
var numRequests int
var muNumRequests sync.Mutex

func addRequest() {
	muNumRequests.Lock()
	defer muNumRequests.Unlock()
	numRequests++
}

func limitReached() bool {
	if numRequests > limit {
		return true
	}
	return false
}

func resetStats() {
	nextCheck := time.Tick(1 * time.Second)

	for {
		<-nextCheck
		muNumRequests.Lock()
		numRequests = 0
		muNumRequests.Unlock()
	}
}

func main() {
	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}

	limitStr, ok := os.LookupEnv("RATE_LIMIT")
	if !ok {
		limit = 100
	} else {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil {
			panic(fmt.Sprintf("Expected integer in RATE_LIMIT but got: %s", err))
		}
	}

	go resetStats()

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(writer http.ResponseWriter, request *http.Request) {
		addRequest()

		if limitReached() {
			msg := fmt.Sprintf("LIMIT REACHED. Current req/s: %d  Limit req/s: %d\n", numRequests, limit)
			fmt.Printf(msg)
			http.Error(writer, msg, http.StatusTooManyRequests)
			return
		}

		time.Sleep(time.Second)
		writer.Write([]byte(fmt.Sprintf("OK. Current req/s: %d  Limit req/s: %d\n", numRequests, limit)))
	})

	fmt.Printf("Rate limit is %d requests per sec\n", limit)
	fmt.Printf("Listening on :%s\n", port)
	http.ListenAndServe(fmt.Sprintf(":%s", port), mux)
}
