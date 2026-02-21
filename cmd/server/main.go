package main

import (
	"net/http"

	"github.com/qesterrx/tplmetrics/internal/config"
	"github.com/qesterrx/tplmetrics/internal/handler"
	"github.com/qesterrx/tplmetrics/internal/logger"
	"github.com/qesterrx/tplmetrics/internal/repository"
)

func main() {

	if err := run(); err != nil {
		panic(err)
	}

}

func run() error {

	logger.InitLogger()

	config, err := config.ParseParamsServer()
	if err != nil {
		return err
	}

	storage := repository.NewMemStorage()

	logger.Log.Debug().Msg("Running server")

	return http.ListenAndServe(config.ServerHost.String(), handler.GetRouter(storage))
}
