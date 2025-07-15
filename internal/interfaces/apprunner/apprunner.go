package apprunner

import (
	"PowerReader/internal/application"
	"PowerReader/internal/domain/files"
	"PowerReader/internal/infrastructure/aida"
	"context"
	"fmt"
	"time"
)

type AppRunner struct{}

func NewAppRunner() *AppRunner {
	return &AppRunner{}
}

func (r *AppRunner) Run() {

	fmt.Println("App starting")

	cxt := context.Background()

	fileWriter := files.NewFileWritter()
	sensorDataFetcher := aida.NewAida()

	application := application.NewApplication(sensorDataFetcher, fileWriter)
	application.StartUpdateLoop(5*time.Second, cxt)
}
