package config

import (
	"flag"
	"fmt"
	"time"
)

type ConfigAgent struct {
	ServerHost       NetAddress
	PoolInterval     int
	ReportInterval   int
	ClientErrorCount int
}

func ParseParamsAgent() (*ConfigAgent, error) {
	var cfg ConfigAgent

	cfg.ClientErrorCount = 1000

	cfg.ServerHost = NetAddress{Host: "localhost", Port: 8080}
	flag.Var(&cfg.ServerHost, "a", "Server's endpoint. Format host:port")

	flag.IntVar(&cfg.PoolInterval, "p", 2, "Time in sec after which collecting mertic (>=1)")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "Time in sec after which sending to server (>=1)")

	flag.Parse()

	if cfg.PoolInterval < 1 {
		return nil, fmt.Errorf("PoolInterval can't be less 1, got %d", cfg.PoolInterval)
	}

	if cfg.ReportInterval < 1 {
		return nil, fmt.Errorf("ReportInterval can't be less 1, got %d", cfg.ReportInterval)
	}

	return &cfg, nil
}

func (ca *ConfigAgent) DurationTerminate() time.Duration {
	if ca.PoolInterval > ca.ReportInterval {
		return time.Duration(ca.PoolInterval) * time.Second
	} else {
		return time.Duration(ca.ReportInterval) * time.Second
	}
}
