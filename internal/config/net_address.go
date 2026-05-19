package config

import (
	"fmt"
	"strconv"
	"strings"
)

// NetAddress - дополнительная структура для хранения адреса в виде набора полей
// Host
// Port
type NetAddress struct {
	Host string
	Port int
}

// Возвращает значение Host:Port
func (na *NetAddress) String() string {
	return na.Host + ":" + strconv.Itoa(na.Port)
}

// Функция установки значения, распарсивает строку Host:Port на отдельные составляющие
// Требуется для работы пакета flag
func (na *NetAddress) Set(value string) error {

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

// Функция установки значения, распарсивает строку Host:Port на отдельные составляющие
// Использует функию
//
//	func (na *NetAddress) Set(value string) error
//
// Требуется для работы github.com/jessevdk/go-flags
func (na *NetAddress) UnmarshalFlag(value string) error {

	return na.Set(value)

}
