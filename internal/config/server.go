package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	"github.com/jessevdk/go-flags"
)

// MetricaStorageMode тип описывающий варианты работы сохранения сервером данных
// Возможные варианты
// MetricaStorageModeSync = "sync" - Синхронный режим, запись происходит сразу при изменении данных
// MetricaStorageModeAsync = "async" - Асинхронный режим, запись происходит через заданный интервал времени
type MetricaStorageMode string

const (
	MetricaStorageModeSync  MetricaStorageMode = "sync"
	MetricaStorageModeAsync MetricaStorageMode = "async"
)

// ConfigServer структура для хранения конфигурации сервера, содержит следующие поля
// ServerHost Адрес сервера, задается параметром "a" или переменной окружения ADDRESS
// StoreInterval Интервал сохранения метрик, задается параметром "i" или переменной окружения STORE_INTERVAL
// FileStorageName Имя файла для сохранения данных в случае если хранение необходимо организовать в файле, задается параметром "f" или переменной окружения FILE_STORAGE_PATH
// RestoreFromFileStorage - Признак необходимости загрузки сохраненных в FileStorageName метрик перед началом работы, задается параметром "r" или переменной окружения RESTORE
// DatabaseDSN Адрес Postgresql сервера в случае если хранение необходимо организовать в БД, задается параметром "d" или переменной окружения DATABASE_DSN
// SecretKeyForSign Ключ для HMAC расшифровки сообщения после получения от клиента, задается параметром "k" или переменной окружения KEY
// AuditFile Задает имя файла в который пишется дополнительный аудит по обновлению метрик, задается параметром "audit-file" или переменной окружения AUDIT_FILE
// AuditURL Задает url в который отправляется POST запрос с дополнительным аудитом по обновлению метрик, задается параметром "audit-url" или переменной окружения AUDIT_URL
// StorageMode Выбранный режим сохранения данных, рассчитывается на основе StoreInterval, если StoreInterval передан 0 то синхронный режим, если больше 0 то асинхронный
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

// ParseParamsServer - Процедура создания структуры ConfigServer на основе параметров командной строки и переменных окружения
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
	if envServerHost, exists := os.LookupEnv("ADDRESS"); exists && envServerHost != "" {
		newServerHost := NetAddress{}
		err := newServerHost.Set(envServerHost)
		if err != nil {
			return nil, fmt.Errorf("env $ADDRESS has wrong format: %v", err.Error())
		} else {
			cfg.ServerHost = newServerHost
		}
	}

	if envStoreInterval, exists := os.LookupEnv("STORE_INTERVAL"); exists && envStoreInterval != "" {
		intStoreInterval, err := strconv.ParseInt(envStoreInterval, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("env $STORE_INTERVAL has wrong format: %v", err.Error())
		}

		cfg.StoreInterval = int(intStoreInterval)
	}

	if envFileStorageName, exists := os.LookupEnv("FILE_STORAGE_PATH"); exists && envFileStorageName != "" {
		cfg.FileStorageName = envFileStorageName
	}

	if envRestoreFromFileStorage, exists := os.LookupEnv("RESTORE"); exists && (envRestoreFromFileStorage == "true" || envRestoreFromFileStorage == "false") {
		if envRestoreFromFileStorage == "true" {
			cfg.RestoreFromFileStorage = true
		} else {
			cfg.RestoreFromFileStorage = false
		}
	}

	if envDatabaseDSN, exists := os.LookupEnv("DATABASE_DSN"); exists && envDatabaseDSN != "" {
		cfg.DatabaseDSN = envDatabaseDSN
	}

	if envKey, exists := os.LookupEnv("KEY"); exists && envKey != "" {
		cfg.SecretKeyForSign = envKey
	}

	if envAuditFile, exists := os.LookupEnv("AUDIT_FILE"); exists && envAuditFile != "" {
		cfg.AuditFile = envAuditFile
	}

	if envAuditURL, exists := os.LookupEnv("AUDIT_URL"); exists && envAuditURL != "" {
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
