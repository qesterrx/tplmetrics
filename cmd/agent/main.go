package main

import "github.com/qesterrx/tplmetrics/internal/agent"

func main() {

	pollInterval := 2
	reportInterval := 10
	host := "http://localhost:8080"

	if err := agent.RunAgent(pollInterval, reportInterval, host); err != nil {
		panic(err)
	}
}
