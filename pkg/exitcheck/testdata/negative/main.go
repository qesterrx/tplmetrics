package main

import "os"

func main() {
	// Проверка обнаружения во вложенных функциях
	defer func() {
		os.Exit(1) // want "os\\.Exit\\(\\) call discover in main\\.main"
	}()

	go func() {
		os.Exit(2) // want "os\\.Exit\\(\\) call discover in main\\.main"
	}()

	// Анонимная функция
	func() {
		os.Exit(3) // want "os\\.Exit\\(\\) call discover in main\\.main"
	}()

	// Обычный вызов
	os.Exit(0) // want "os\\.Exit\\(\\) call discover in main\\.main"
}
