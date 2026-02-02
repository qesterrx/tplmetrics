package config

import (
	"fmt"
	"strconv"
	"strings"
)

type NetAddress struct {
	Host string
	Port int
}

func (na *NetAddress) String() string {
	return na.Host + ":" + strconv.Itoa(na.Port)
}

func (na *NetAddress) Set(value string) error {
	//fmt.Println("NetAddress.Set value=", value)

	splt := strings.Split(value, ":")
	if len(splt) != 2 {
		return fmt.Errorf("param must have format host:port")
	}

	port, err := strconv.Atoi(splt[1])
	if err != nil {
		return err
	}

	na.Host = splt[0]
	na.Port = port

	return nil
}
