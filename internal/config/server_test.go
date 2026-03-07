package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseParamsServer(t *testing.T) {
	app := os.Args[0]

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
				RestoreFromFileStorage: false},
			err: false,
		},
		{
			name:    "correct flag without ENV",
			flagset: []string{"-a=address1:1", "-i=301", "-f=TempFileStorage1", "-r"},
			envset:  nil,
			expected: ConfigServer{
				ServerHost:             NetAddress{Host: "address1", Port: 1},
				StoreInterval:          301,
				FileStorageName:        "TempFileStorage1",
				RestoreFromFileStorage: true},
			err: false,
		},
		{
			name:    "correct flag with ENV",
			flagset: []string{"-a=address1:1", "-i=301", "-f=TempFileStorage1"},
			envset:  map[string]string{"ADDRESS": "address2:2", "STORE_INTERVAL": "302", "FILE_STORAGE_PATH": "TempFileStorage2", "RESTORE": "true"},
			expected: ConfigServer{
				ServerHost:             NetAddress{Host: "address2", Port: 2},
				StoreInterval:          302,
				FileStorageName:        "TempFileStorage2",
				RestoreFromFileStorage: true},
			err: false,
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

			config, err := ParseParamsServer()

			if test.err {
				assert.Error(t, err, os.Args)
			} else {
				assert.NoError(t, err, os.Args)
				assert.Equal(t, test.expected, *config)
			}

		})
	}
}
