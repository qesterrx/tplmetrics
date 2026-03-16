package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

const ClientErrorCount = 1000

type ConfigAgent struct {
	ServerHost       NetAddress
	PoolInterval     int
	ReportInterval   int
	ClientErrorCount int
	SecretKeyForSign string
}

func ParseParamsAgent() (*ConfigAgent, error) {

	fs := flag.NewFlagSet("", flag.PanicOnError)

	var cfg ConfigAgent

	cfg.ClientErrorCount = ClientErrorCount

	cfg.ServerHost = NetAddress{Host: "localhost", Port: 8080}
	fs.Var(&cfg.ServerHost, "a", "Server's endpoint. Format host:port")

	fs.IntVar(&cfg.PoolInterval, "p", 2, "PoolInterval - time in sec after which collecting mertic (>=1)")
	fs.IntVar(&cfg.ReportInterval, "r", 10, "ReportInterval - time in sec after which sending to server (>=1)")
	fs.StringVar(&cfg.SecretKeyForSign, "k", "", "SecretKeyForSign - key for sign data in header HashSHA256")

	fs.Parse(os.Args[1:])

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

	if envPoolInterval := os.Getenv("POLL_INTERVAL"); envPoolInterval != "" {
		intPoolInterval, err := strconv.ParseInt(envPoolInterval, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("env $POLL_INTERVAL has wrong format: %v", err.Error())
		}
		cfg.PoolInterval = int(intPoolInterval)
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		intReportInterval, err := strconv.ParseInt(envReportInterval, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("env $REPORT_INTERVAL has wrong format: %v", err.Error())
		}
		cfg.ReportInterval = int(intReportInterval)
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.SecretKeyForSign = envKey
	}

	//Дополнительные проверки параметров
	if cfg.PoolInterval < 1 {
		return nil, fmt.Errorf("PoolInterval can't be less 1, got %d, check flag -p and ENV $POLL_INTERVAL", cfg.PoolInterval)
	}

	if cfg.ReportInterval < 1 {
		return nil, fmt.Errorf("ReportInterval can't be less 1, got %d, check flag -r and ENV $REPORT_INTERVAL", cfg.ReportInterval)
	}

	return &cfg, nil
}
