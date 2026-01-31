package main

import (
	"net/http"

	"github.com/qesterrx/tplmetrics/internal/handler"
	"github.com/qesterrx/tplmetrics/internal/repository"
)

func main() {

	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {
	storage := repository.NewMemStorage()

	mux := http.NewServeMux()

	mux.HandleFunc("/update/{kind}/{name}/{value}", handler.UpdateMetric(storage))

	return http.ListenAndServe("localhost:8080", mux)
}
