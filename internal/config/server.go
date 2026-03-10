package config

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
)

type ConfigServer struct {
	ServerHost             NetAddress
	StoreInterval          int
	FileStorageName        string
	RestoreFromFileStorage bool
	DatabaseDSN            string
}

func ParseParamsServer() (*ConfigServer, error) {

	fs := flag.NewFlagSet("", flag.PanicOnError)

	var cfg ConfigServer
	cfg.ServerHost = NetAddress{Host: "localhost", Port: 8080}
	fs.Var(&cfg.ServerHost, "a", "Endpoint for server. Format host:port")
	fs.IntVar(&cfg.StoreInterval, "i", 300, "StoreInterval - time in sec after which data would be save in file")
	fs.StringVar(&cfg.FileStorageName, "f", "TempFileStorage", "filename for soraging data")
	fs.BoolVar(&cfg.RestoreFromFileStorage, "r", false, "Load data from file on start")
	fs.StringVar(&cfg.DatabaseDSN, "d", "", "Connection string for postgresql")

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

	if envStoreInterval := os.Getenv("STORE_INTERVAL"); envStoreInterval != "" {
		intStoreInterval, err := strconv.ParseInt(envStoreInterval, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("env $STORE_INTERVAL has wrong format: %v", err.Error())
		}

		cfg.StoreInterval = int(intStoreInterval)
	}

	if envFileStorageName := os.Getenv("FILE_STORAGE_PATH"); envFileStorageName != "" {
		cfg.FileStorageName = envFileStorageName
	}

	if envRestoreFromFileStorage := os.Getenv("RESTORE"); envRestoreFromFileStorage == "true" || envRestoreFromFileStorage == "false" {
		if envRestoreFromFileStorage == "true" {
			cfg.RestoreFromFileStorage = true
		} else {
			cfg.RestoreFromFileStorage = false
		}
	}

	if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	//Дополнительные проверки параметров
	if cfg.DatabaseDSN != "" {

		matched, _ := regexp.MatchString(`^postgres://.*/.*?sslmode=.*$`, cfg.DatabaseDSN)
		if !matched {
			return nil, fmt.Errorf("неверный формат строки подключения к БД PostgreSQL (%s)", cfg.DatabaseDSN)
		}

	}

	return &cfg, nil
}
