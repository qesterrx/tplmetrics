// Пакет config содержит в себе описание объектов и функция для конфигурации приложения (клиента или сервера)
package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"os"
	"strconv"

	"github.com/jessevdk/go-flags"
)

// ConfigAgent структура для хранения конфигурации клиента, содержит следующие поля
// ServerHost Адрес клиента, задается параметром "a" или переменной окружения ADDRESS
// PoolInterval Интервал опроса метрик, задается параметром "p" или переменной окружения POLL_INTERVAL
// ReportInterval Интервал отправки метрик на сервер, задается параметром "r" или переменной окружения REPORT_INTERVAL
// RateLimit Количество горутин, отправляющий данные на сервер, задается параметром "l" или переменной окружения RATE_LIMIT
// CryptoKey Задает путь к публичному ключу для шифрования сообщений от агента
// SecretKeyForSign Ключ для HMAC шифрования сообщения перед отправкой, задается параметром "k" или переменной окружения KEY
// Config Имя JSON файла с параметрами приложения
// PublicKeyRSA публичный ключ шифрования
type ConfigAgent struct {
	ServerHost       NetAddress `short:"a" long:"address" description:"Endpoint for server. Format host:port" default:"localhost:8080"`
	PoolInterval     int        `short:"p" long:"pool" description:"PoolInterval - time in sec after which collecting mertic (>=1)" default:"2"`
	ReportInterval   int        `short:"r" long:"report" description:"ReportInterval - time in sec after which sending to server (>=1)" default:"10"`
	RateLimit        int        `short:"l" long:"rate" description:"RateLimit - count of clients for send metric (>=1)" default:"1"`
	SecretKeyForSign string     `short:"k" long:"secret" description:"SecretKeyForSign - key for sign data in header HashSHA256" default:""`
	CryptoKey        string     `long:"crypto-key" description:"Filename public key for RSA" default:""`
	Config           string     `short:"c" long:"config" description:"Filename with json config" default:""`
	PublicKeyRSA     *x509.Certificate
}

// структура для загрузки параметров из JSON
type configAgentJSON struct {
	ServerHost       *string `json:"address,omitempty"`
	PoolInterval     *int    `json:"poll_interval,omitempty"`
	ReportInterval   *int    `json:"report_interval,omitempty"`
	RateLimit        *int    `json:"rate_limit,omitempty"`
	SecretKeyForSign *string `json:"secret,omitempty"`
	CryptoKey        *string `json:"crypto_key,omitempty"`
}

// ParseParamsAgent - Процедура создания структуры ConfigAgent на основе параметров командной строки и переменных окружения
func ParseParamsAgent() (*ConfigAgent, error) {

	var cfg ConfigAgent

	// Создаем парсер (аналог FlagSet)
	parser := flags.NewParser(&cfg, flags.Default)

	// Парсим аргументы
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		return nil, err
	}

	//Проверяем переменную для загрузки конфигурации из JSON
	if envConfig, exists := os.LookupEnv("CONFIG"); exists && envConfig != "" {
		cfg.Config = envConfig
	}

	//Доопределим конфигурацию из JSON файла
	err = cfg.redefineFromJSON(parser)
	if err != nil {
		return nil, fmt.Errorf("JSON Config: %v", err.Error())
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

	if envCryptoKey, exists := os.LookupEnv("CRYPTO_KEY"); exists && envCryptoKey != "" {
		cfg.CryptoKey = envCryptoKey
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

// LoadPublicKey - функиця загрузки публичного ключа для RSA шифрования обращений к серверу
func (cfg *ConfigAgent) LoadPublicKey() error {
	if cfg.CryptoKey == "" {
		return nil
	}

	certFile, err := os.ReadFile(cfg.CryptoKey)
	if err != nil {
		return fmt.Errorf("Ошибка чтения файла %s: %v", cfg.CryptoKey, err)
	}

	certBlock, _ := pem.Decode(certFile)
	if certBlock == nil {
		return fmt.Errorf("Сертификат не найден в %s: %v", cfg.CryptoKey, err)
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return err
	}

	//Разборки со слишком длинным сообщением
	key := cert.PublicKey.(*rsa.PublicKey)
	bits := key.N.BitLen()
	bytes := key.N.BitLen() / 8

	fmt.Printf("Key size: %d bits (%d bytes)\n", bits, bytes)
	fmt.Printf("Max message size: %d bytes\n", bytes-2*32-2)

	cfg.PublicKeyRSA = cert
	return nil
}

// RedefineFromJSON - доопределение параметров из JSON файла, типичный способ как сделать простое и понятное сложным и непонятным. За это программисты и получают свои 100500К/наносекунду
func (cfg *ConfigAgent) redefineFromJSON(parser *flags.Parser) error {

	if cfg.Config != "" {

		data, err := os.ReadFile(cfg.Config)
		if err != nil {
			return err
		}

		if len(data) > 0 {

			cfgJSON := configAgentJSON{}

			err = json.Unmarshal(data, &cfgJSON)
			if err != nil {
				return err
			}

			// Для наименьшего приоритета надо переопределить только те параметры которые были не заданы аргументами

			opt := parser.FindOptionByLongName("address")
			if (opt == nil || opt.IsSetDefault()) && cfgJSON.ServerHost != nil {
				newServerHost := NetAddress{}
				err := newServerHost.Set(*cfgJSON.ServerHost)
				if err != nil {
					return fmt.Errorf("JSON CONFIG address has wrong format: %v", err.Error())
				} else {
					cfg.ServerHost = newServerHost
				}
			}

			opt = parser.FindOptionByLongName("pool")
			if (opt == nil || opt.IsSetDefault()) && cfgJSON.PoolInterval != nil {
				cfg.PoolInterval = *cfgJSON.PoolInterval
			}

			opt = parser.FindOptionByLongName("report")
			if (opt == nil || opt.IsSetDefault()) && cfgJSON.ReportInterval != nil {
				cfg.ReportInterval = *cfgJSON.ReportInterval
			}

			opt = parser.FindOptionByLongName("rate")
			if (opt == nil || opt.IsSetDefault()) && cfgJSON.RateLimit != nil {
				cfg.RateLimit = *cfgJSON.RateLimit
			}

			opt = parser.FindOptionByLongName("secret")
			if (opt == nil || opt.IsSetDefault()) && cfgJSON.SecretKeyForSign != nil {
				cfg.SecretKeyForSign = *cfgJSON.SecretKeyForSign
			}

			opt = parser.FindOptionByLongName("crypto-key")
			if (opt == nil || opt.IsSetDefault()) && cfgJSON.CryptoKey != nil {
				cfg.CryptoKey = *cfgJSON.CryptoKey
			}

		}

	}
	return nil
}
