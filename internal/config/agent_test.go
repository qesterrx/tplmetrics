package config

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseParamsAgent(t *testing.T) {
	app := os.Args[0]

	//Создаем временный файл с json настройками
	jsonConfig := configAgentJSON{
		ServerHost:       func() *string { s := "localhost:8091"; return &s }(),
		PoolInterval:     func() *int { i := 11; return &i }(),
		ReportInterval:   func() *int { i := 12; return &i }(),
		RateLimit:        func() *int { i := 13; return &i }(),
		SecretKeyForSign: func() *string { s := "json"; return &s }(),
		CryptoKey:        func() *string { s := "json"; return &s }(),
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
		jsoncfg  string
		expected ConfigAgent
		err      bool //Сокращенный вариант, не стал добавлять сюда все переменные, мне главное проверить ошибки
	}{
		{
			name:    "default flag",
			flagset: nil,
			envset:  nil,
			expected: ConfigAgent{
				ServerHost:       NetAddress{Host: "localhost", Port: 8080},
				PoolInterval:     2,
				ReportInterval:   10,
				SecretKeyForSign: "",
				RateLimit:        1,
				CryptoKey:        "",
			},
			err: false,
		},
		{
			name:    "json flag",
			flagset: []string{"-c=" + file.Name()},
			envset:  nil,
			expected: ConfigAgent{
				ServerHost:       NetAddress{Host: "localhost", Port: 8091},
				PoolInterval:     11,
				ReportInterval:   12,
				SecretKeyForSign: "json",
				RateLimit:        13,
				CryptoKey:        "json",
				Config:           file.Name(),
			},
			err: false,
		},
		{
			name:    "json ENV",
			flagset: nil,
			envset: map[string]string{
				"CONFIG": file.Name()},
			expected: ConfigAgent{
				ServerHost:       NetAddress{Host: "localhost", Port: 8091},
				PoolInterval:     11,
				ReportInterval:   12,
				SecretKeyForSign: "json",
				RateLimit:        13,
				CryptoKey:        "json",
				Config:           file.Name(),
			},
			err: false,
		},
		{
			name:    "correct flag without ENV",
			flagset: []string{"-a=address1:1", "-p=123", "-r=456", "-k=qwerty", "-l=5", "--crypto-key=asd"},
			envset:  nil,
			expected: ConfigAgent{
				ServerHost:       NetAddress{Host: "address1", Port: 1},
				PoolInterval:     123,
				ReportInterval:   456,
				SecretKeyForSign: "qwerty",
				RateLimit:        5,
				CryptoKey:        "asd",
			},
			err: false,
		},
		{
			name:    "correct flag with ENV",
			flagset: []string{"-a=address1:1", "-p=123", "-r=456", "-k=qwerty", "-l=5", "--crypto-key=asd"},
			envset: map[string]string{
				"ADDRESS":         "address2:2",
				"POLL_INTERVAL":   "321",
				"REPORT_INTERVAL": "654",
				"KEY":             "asdfg",
				"RATE_LIMIT":      "10",
				"CRYPTO_KEY":      "dsa",
			},
			expected: ConfigAgent{
				ServerHost:       NetAddress{Host: "address2", Port: 2},
				PoolInterval:     321,
				ReportInterval:   654,
				SecretKeyForSign: "asdfg",
				RateLimit:        10,
				CryptoKey:        "dsa",
			},
			err: false,
		},
		{
			name:     "error flag r=0",
			flagset:  []string{"-a=localhost:8080", "-p=2", "-r=0"},
			envset:   nil,
			expected: ConfigAgent{},
			err:      true,
		},
		{
			name:     "error flag r is negative",
			flagset:  []string{"-a=localhost:8080", "-p=2", "-r=-1"},
			envset:   nil,
			expected: ConfigAgent{},
			err:      true,
		},
		{
			name:     "error flag p=0",
			flagset:  []string{"-a=localhost:8080", "-p=0", "-r=10"},
			envset:   nil,
			expected: ConfigAgent{},
			err:      true,
		},
		{
			name:     "error flag p is negative",
			flagset:  []string{"-a=localhost:8080", "-p=-1", "-r=10"},
			envset:   nil,
			expected: ConfigAgent{},
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			os.Args = append([]string{app}, test.flagset...)

			if test.envset != nil {
				for k, v := range test.envset {
					t.Setenv(k, v)
				}
			}

			config, err := ParseParamsAgent()

			if test.err {
				assert.Error(t, err, os.Args)
			} else {
				assert.NoError(t, err, os.Args)

				assert.Equal(t, test.expected, *config)
			}

		})
	}

}
