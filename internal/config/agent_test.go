package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseParamsAgent(t *testing.T) {
	app := os.Args[0]

	tests := []struct {
		name     string
		flagset  []string
		envset   map[string]string
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
				ClientErrorCount: ClientErrorCount,
				SecretKeyForSign: "",
				RateLimit:        1,
			},
			err: false,
		},
		{
			name:    "correct flag without ENV",
			flagset: []string{"-a=address1:1", "-p=123", "-r=456", "-k=qwerty", "-l=5"},
			envset:  nil,
			expected: ConfigAgent{
				ServerHost:       NetAddress{Host: "address1", Port: 1},
				PoolInterval:     123,
				ReportInterval:   456,
				ClientErrorCount: ClientErrorCount,
				SecretKeyForSign: "qwerty",
				RateLimit:        5,
			},
			err: false,
		},
		{
			name:    "correct flag with ENV",
			flagset: []string{"-a=address1:1", "-p=123", "-r=456"},
			envset: map[string]string{
				"ADDRESS":         "address2:2",
				"POLL_INTERVAL":   "321",
				"REPORT_INTERVAL": "654",
				"KEY":             "asdfg",
				"RATE_LIMIT":      "10",
			},
			expected: ConfigAgent{
				ServerHost:       NetAddress{Host: "address2", Port: 2},
				PoolInterval:     321,
				ReportInterval:   654,
				ClientErrorCount: ClientErrorCount,
				SecretKeyForSign: "asdfg",
				RateLimit:        10,
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
