package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/jessevdk/go-flags"
)

type MetricaStorageMode string

const (
	MetricaStorageModeSync  MetricaStorageMode = "sync"
	MetricaStorageModeAsync MetricaStorageMode = "async"
)

type ConfigServer struct {
	ServerHost             NetAddress `short:"a" long:"address" description:"Endpoint for server. Format host:port" default:"localhost:8080"`
	StoreInterval          int        `short:"i" long:"store-interval" description:"StoreInterval - time in sec after which data would be save in file" default:"300"`
	FileStorageName        string     `short:"f" long:"file-storage" description:"filename for storing data" default:"TempFileStorage"`
	RestoreFromFileStorage bool       `short:"r" long:"restore" description:"Load data from file on start"`
	DatabaseDSN            string     `short:"d" long:"database-dsn" description:"Connection string for postgresql"`
	SecretKeyForSign       string     `short:"k" long:"secret-key" description:"SecretKeyForSign - key for sign data in header HashSHA256"`
	AuditFile              string     `long:"audit-file" description:"AuditFile - Filename Subscriber on Update metrica event saving data to file" default:""`
	AuditURL               string     `long:"audit-url" description:"AuditURL - URL Subscriber on Update metrica event sending data to URL" default:""`
	StorageMode            MetricaStorageMode
}

func ParseParamsServer() (*ConfigServer, error) {

	var cfg ConfigServer

	// Создаем парсер (аналог FlagSet)
	parser := flags.NewParser(&cfg, flags.Default)

	// Парсим аргументы
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		return nil, err
	}

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

	if envKey := os.Getenv("KEY"); envKey != "" {
		cfg.SecretKeyForSign = envKey
	}

	if envAuditFile := os.Getenv("AUDIT_FILE"); envAuditFile != "" {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL := os.Getenv("AUDIT_URL"); envAuditURL != "" {
		cfg.AuditURL = envAuditURL
	}

	//Дополнительные проверки параметров
	if cfg.DatabaseDSN != "" {

		matched, _ := regexp.MatchString(`^postgres://.*/.*?sslmode=.*$`, cfg.DatabaseDSN)
		if !matched {
			return nil, fmt.Errorf("неверный формат строки подключения к БД PostgreSQL (%s)", cfg.DatabaseDSN)
		}

	}

	//Дополнительная трансляция параметров
	if cfg.StoreInterval == 0 {
		cfg.StorageMode = MetricaStorageModeSync
	} else {
		cfg.StorageMode = MetricaStorageModeAsync
	}

	return &cfg, nil
}
