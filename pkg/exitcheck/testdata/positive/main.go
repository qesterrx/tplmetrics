package main

import "os"

func main() {
	helper()
}

func helper() {
	// Этот вызов не должен быть обнаружен, так как не в main
	os.Exit(1)
}
