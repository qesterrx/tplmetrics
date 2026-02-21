package config

import (
	"flag"
	"fmt"
	"os"
)

type ConfigServer struct {
	ServerHost NetAddress
}

func ParseParamsServer() (*ConfigServer, error) {
	var cfg ConfigServer
	cfg.ServerHost = NetAddress{Host: "localhost", Port: 8080}
	flag.Var(&cfg.ServerHost, "a", "Endpoint for server. Format host:port")

	flag.Parse()

	//Переопределим параметрами из ENV
	if envServerHost := os.Getenv("ADDRESS"); envServerHost != "" {
		newServerHost := NetAddress{}
		err := newServerHost.Set(envServerHost)
		if err != nil {
			return nil, fmt.Errorf("env $ADDRESS has wrong format: %v", err.Error())
		} else {
			cfg.ServerHost = newServerHost
		}
	}

	return &cfg, nil
}
