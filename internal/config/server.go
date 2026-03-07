package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

type ConfigServer struct {
	ServerHost             NetAddress
	StoreInterval          int
	FileStorageName        string
	RestoreFromFileStorage bool
	DatabaseDSN            string
	DatabaseURL            string
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

	//Перевод
	if cfg.DatabaseDSN != "" {
		var host string
		var user string
		var password string
		var dbname string
		var sslmode string

		_, err := fmt.Sscanf(cfg.DatabaseDSN, "host=%s user=%s password=%s dbname=%s sslmode=%s", &host, &user, &password, &dbname, &sslmode)
		if err != nil {
			cfg.DatabaseURL = ""
		} else {
			cfg.DatabaseURL = fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=%s", user, password, host, dbname, sslmode)
		}

		fmt.Println(cfg.DatabaseURL)

	}

	return &cfg, nil
}
