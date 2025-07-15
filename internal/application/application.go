package application

import (
	"PowerReader/internal/domain/files"
	"PowerReader/internal/domain/sensors"
	"context"
	"fmt"
	"time"
)

type Application struct {
	sensorDataFetcher sensors.SensorDataFetcher
	fileWriter        *files.FileWritter
}

func NewApplication(sensorDataFetcher sensors.SensorDataFetcher, fileWriter *files.FileWritter) *Application {
	return &Application{
		sensorDataFetcher: sensorDataFetcher,
		fileWriter:        fileWriter,
	}
}

func (a *Application) StartUpdateLoop(interval time.Duration, ctx context.Context) {

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:

			sensors, err := a.sensorDataFetcher.GetSensorData()
			if err != nil {
				fmt.Printf("Failed to read sensor values: %s", err.Error())
				break
			}

			for i, sensor := range sensors.Powers {

				fmt.Printf("(%d)%s: %f\n", i, sensor.ID, sensor.Value)

				value := sensor.Value / 10

				err := a.fileWriter.WirteSensorFile(
					sensor.ID,
					[]byte(fmt.Sprintf("%f", value)),
				)

				if err != nil {
					fmt.Printf("Failed to write sensor values: %s", err.Error())
				}
			}

		case <-ctx.Done():
			return
		}
	}
}
