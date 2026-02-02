package config

import "flag"

type ConfigServer struct {
	ServerHost NetAddress
}

func ParseParamsServer() *ConfigServer {
	var cfg ConfigServer
	cfg.ServerHost = NetAddress{Host: "localhost", Port: 8080}
	flag.Var(&cfg.ServerHost, "a", "Endpoint for server. Format host:port")

	flag.Parse()

	return &cfg
}
