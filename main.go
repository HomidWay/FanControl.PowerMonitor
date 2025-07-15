package main

import (
	"PowerReader/internal/interfaces/apprunner"
)



func main() {
	AppRunner := apprunner.NewAppRunner()
	AppRunner.Run()
}
