package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseParamsAgent(t *testing.T) {
	app := os.Args[0]

	tests := []struct {
		flagset []string
		err     bool //Сокращенный вариант, не стал добавлять сюда все переменные, мне главное проверить ошибки
	}{
		{flagset: []string{"-a=localhost:8080", "-p=2", "-r=10"}, err: false},
		{flagset: []string{"-a=localhost:8080", "-p=2", "-r=0"}, err: true},
		{flagset: []string{"-a=localhost:8080", "-p=2", "-r=-1"}, err: true},
		{flagset: []string{"-a=localhost:8080", "-p=0", "-r=10"}, err: true},
		{flagset: []string{"-a=localhost:8080", "-p=-1", "-r=10"}, err: true},
	}

	for _, test := range tests {
		os.Args = test.flagset
		flag.CommandLine = flag.NewFlagSet(app, flag.ExitOnError)

		_, err := ParseParamsAgent()
		if test.err {
			assert.Error(t, err, os.Args)
		} else {
			assert.NoError(t, err, os.Args)
		}
	}

}
