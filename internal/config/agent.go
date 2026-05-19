// Пакет config содержит в себе описание объектов и функция для конфигурации приложения (клиента или сервера)
package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// ConfigAgent структура для хранения конфигурации клиента, содержит следующие поля
// ServerHost Адрес клиента, задается параметром "a" или переменной окружения ADDRESS
// PoolInterval Интервал опроса метрик, задается параметром "p" или переменной окружения POLL_INTERVAL
// ReportInterval Интервал отправки метрик на сервер, задается параметром "r" или переменной окружения REPORT_INTERVAL
// RateLimit - Количество горутин, отправляющий данные на сервер, задается параметром "l" или переменной окружения RATE_LIMIT
// SecretKeyForSign Ключ для HMAC шифрования сообщения перед отправкой, задается параметром "k" или переменной окружения KEY
type ConfigAgent struct {
	ServerHost       NetAddress
	PoolInterval     int
	ReportInterval   int
	RateLimit        int
	SecretKeyForSign string
}

// ParseParamsAgent - Процедура создания структуры ConfigAgent на основе параметров командной строки и переменных окружения
func ParseParamsAgent() (*ConfigAgent, error) {

	fs := flag.NewFlagSet("", flag.PanicOnError)

	var cfg ConfigAgent

	cfg.ServerHost = NetAddress{Host: "localhost", Port: 8080}
	fs.Var(&cfg.ServerHost, "a", "Server's endpoint. Format host:port")

	fs.IntVar(&cfg.PoolInterval, "p", 2, "PoolInterval - time in sec after which collecting mertic (>=1)")
	fs.IntVar(&cfg.RateLimit, "l", 1, "RateLimit - count of clients for send metric (>=1)")
	fs.IntVar(&cfg.ReportInterval, "r", 10, "ReportInterval - time in sec after which sending to server (>=1)")
	fs.StringVar(&cfg.SecretKeyForSign, "k", "", "SecretKeyForSign - key for sign data in header HashSHA256")

	fs.Parse(os.Args[1:])

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

	if envPoolInterval, exists := os.LookupEnv("POLL_INTERVAL"); exists && envPoolInterval != "" {
		intPoolInterval, err := strconv.ParseInt(envPoolInterval, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("env $POLL_INTERVAL has wrong format: %v", err.Error())
		}
		cfg.PoolInterval = int(intPoolInterval)
	}

	if envReportInterval, exists := os.LookupEnv("REPORT_INTERVAL"); exists && envReportInterval != "" {
		intReportInterval, err := strconv.ParseInt(envReportInterval, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("env $REPORT_INTERVAL has wrong format: %v", err.Error())
		}
		cfg.ReportInterval = int(intReportInterval)
	}

	if envRateLimit, exists := os.LookupEnv("RATE_LIMIT"); exists && envRateLimit != "" {
		intRateLimit, err := strconv.ParseInt(envRateLimit, 10, 0)
		if err != nil {
			return nil, fmt.Errorf("env $RATE_LIMIT has wrong format: %v", err.Error())
		}
		cfg.RateLimit = int(intRateLimit)
	}

	if envKey, exists := os.LookupEnv("KEY"); exists && envKey != "" {
		cfg.SecretKeyForSign = envKey
	}

	//Дополнительные проверки параметров
	if cfg.PoolInterval < 1 {
		return nil, fmt.Errorf("PoolInterval can't be less 1, got %d, check flag -p and ENV $POLL_INTERVAL", cfg.PoolInterval)
	}

	if cfg.ReportInterval < 1 {
		return nil, fmt.Errorf("ReportInterval can't be less 1, got %d, check flag -r and ENV $REPORT_INTERVAL", cfg.ReportInterval)
	}

	if cfg.RateLimit < 1 {
		return nil, fmt.Errorf("RateLimit can't be less 1, got %d, check flag -p and ENV $RATE_LIMIT", cfg.RateLimit)
	}

	return &cfg, nil
}
