package config

import (
	"encoding/json"
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseParamsServer(t *testing.T) {
	app := os.Args[0]

	//Создаем временный файл с json настройками
	jsonConfig := configServerJSON{
		ServerHost:             func() *string { s := "localhost:8091"; return &s }(),
		StoreInterval:          func() *int { i := 11; return &i }(),
		FileStorageName:        func() *string { s := "json_1"; return &s }(),
		RestoreFromFileStorage: func() *bool { b := true; return &b }(),
		DatabaseDSN:            func() *string { s := "postgres://pgs/pgs?sslmode=pgs"; return &s }(),
		SecretKeyForSign:       func() *string { s := "json_3"; return &s }(),
		AuditFile:              func() *string { s := "json_4"; return &s }(),
		AuditURL:               func() *string { s := "json_5"; return &s }(),
		CryptoKey:              func() *string { s := "json_6"; return &s }(),
		TrustedSubnet:          func() *string { s := "192.168.0.1/24"; return &s }(),
	}

	data, err := json.Marshal(&jsonConfig)
	assert.NoError(t, err)
	file, err := os.CreateTemp(".", "tmp_config.json")
	assert.NoError(t, err)
	_, err = file.Write(data)
	assert.NoError(t, err)
	file.Close()

	// удалить файл после использования
	defer os.Remove(file.Name())

	tests := []struct {
		name     string
		flagset  []string
		envset   map[string]string
		expected ConfigServer
		err      bool //Сокращенный вариант, не стал добавлять сюда все переменные, мне главное проверить ошибки
	}{
		{
			name:    "default flag",
			flagset: nil,
			envset:  nil,
			expected: ConfigServer{
				ServerHost:             NetAddress{Host: "localhost", Port: 8080},
				StoreInterval:          300,
				FileStorageName:        "TempFileStorage",
				RestoreFromFileStorage: false,
				StorageMode:            MetricaStorageModeAsync,
				AuditFile:              "",
				AuditURL:               "",
				CryptoKey:              "",
				SecretKeyForSign:       "",
				TrustedSubnet:          "",
			},
			err: false,
		},
		{
			name:    "json flag",
			flagset: []string{"-c=" + file.Name()},
			envset:  nil,
			expected: ConfigServer{
				ServerHost:             NetAddress{Host: "localhost", Port: 8091},
				StoreInterval:          11,
				FileStorageName:        "json_1",
				RestoreFromFileStorage: true,
				StorageMode:            MetricaStorageModeAsync,
				DatabaseDSN:            "postgres://pgs/pgs?sslmode=pgs",
				AuditFile:              "json_4",
				AuditURL:               "json_5",
				CryptoKey:              "json_6",
				SecretKeyForSign:       "json_3",
				Config:                 file.Name(),
				TrustedSubnet:          "192.168.0.1/24",
			},
			err: false,
		},
		{
			name:    "json ENV",
			flagset: nil,
			envset: map[string]string{
				"CONFIG": file.Name()},
			expected: ConfigServer{
				ServerHost:             NetAddress{Host: "localhost", Port: 8091},
				StoreInterval:          11,
				FileStorageName:        "json_1",
				RestoreFromFileStorage: true,
				StorageMode:            MetricaStorageModeAsync,
				DatabaseDSN:            "postgres://pgs/pgs?sslmode=pgs",
				AuditFile:              "json_4",
				AuditURL:               "json_5",
				CryptoKey:              "json_6",
				SecretKeyForSign:       "json_3",
				Config:                 file.Name(),
				TrustedSubnet:          "192.168.0.1/24",
			},
			err: false,
		},
		{
			name: "correct flag without ENV",
			flagset: []string{"-a=address1:1",
				"-i=301",
				"-f=TempFileStorage1",
				"-r",
				"-d=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
				"-k=asd",
				"--audit-file=FILENAME1",
				"--audit-url=URL1",
				"--crypto-key=private.pem",
				"-t=192.168.0.2/24",
			},
			envset: nil,
			expected: ConfigServer{
				ServerHost:             NetAddress{Host: "address1", Port: 1},
				StoreInterval:          301,
				FileStorageName:        "TempFileStorage1",
				RestoreFromFileStorage: true,
				DatabaseDSN:            "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
				StorageMode:            MetricaStorageModeAsync,
				AuditFile:              "FILENAME1",
				AuditURL:               "URL1",
				CryptoKey:              "private.pem",
				SecretKeyForSign:       "asd",
				TrustedSubnet:          "192.168.0.2/24",
			},
			err: false,
		},
		{
			name: "correct flag with ENV",
			flagset: []string{"-a=address1:1",
				"-i=301",
				"-f=TempFileStorage1",
				"-d=postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
				"-k=asd",
				"--audit-file=FILENAME1",
				"--audit-url=URL1",
				"--crypto-key=private_skip.pem",
				"-t=192.168.0.2/24",
			},
			envset: map[string]string{
				"ADDRESS":           "address2:2",
				"STORE_INTERVAL":    "302",
				"FILE_STORAGE_PATH": "TempFileStorage2",
				"RESTORE":           "true",
				"DATABASE_DSN":      "postgres://postgres2:postgres2@localhost:5432/postgres2?sslmode=disable",
				"AUDIT_FILE":        "FILENAME2",
				"AUDIT_URL":         "URL2",
				"CRYPTO_KEY":        "private.pem",
				"KEY":               "dsa",
				"TRUSTED_SUBNET":    "192.168.0.3/24",
			},
			expected: ConfigServer{
				ServerHost:             NetAddress{Host: "address2", Port: 2},
				StoreInterval:          302,
				FileStorageName:        "TempFileStorage2",
				RestoreFromFileStorage: true,
				DatabaseDSN:            "postgres://postgres2:postgres2@localhost:5432/postgres2?sslmode=disable",
				StorageMode:            MetricaStorageModeAsync,
				AuditFile:              "FILENAME2",
				AuditURL:               "URL2",
				CryptoKey:              "private.pem",
				SecretKeyForSign:       "dsa",
				TrustedSubnet:          "192.168.0.3/24",
			},
			err: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			//Установка аргументов
			os.Args = append([]string{app}, test.flagset...)

			//Установка переменных
			if test.envset != nil {
				for k, v := range test.envset {
					t.Setenv(k, v)
				}
			}

			//Тестовый вызов парсинга параметров
			config, err := ParseParamsServer()

			//Добавим определение маски
			if test.expected.TrustedSubnet != "" {
				_, test.expected.MaskSubnet, err = net.ParseCIDR(test.expected.TrustedSubnet)
				assert.NoError(t, err)
			}

			//Проверки
			if test.err {
				assert.Error(t, err, os.Args)
			} else {
				assert.NoError(t, err, os.Args)
				assert.Equal(t, test.expected, *config)
			}

		})
	}
}
