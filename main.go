package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"
)

var regPath = `Software\FinalWire\AIDA64\SensorValues`

type SensorData struct {
	Fans     []Sensor `xml:"fan"`
	Voltages []Sensor `xml:"volt"`
	Currents []Sensor `xml:"curr"`
	Powers   []Sensor `xml:"pwr"`
}

type Sensor struct {
	ID    string  `xml:"id"`
	Label string  `xml:"label"`
	Value float64 `xml:"value"`
}

func main() {

	for {

		sensorData, err := readPowerSensordShared()
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		fmt.Printf("Sensors captured at: %s\n", time.Now().UTC())

		for _, powerSensor := range sensorData.Powers {
			fmt.Printf("%s: %f\n", powerSensor.Label, powerSensor.Value)

			filePath, err := getWritablePath(powerSensor.ID)
			if err != nil {
				log.Printf("Error getting writable path: %v", err)
				time.Sleep(time.Second * 5)
				continue
			}

			divided := powerSensor.Value / 10

			err = os.WriteFile(filePath, []byte(fmt.Sprintf("%f\n", divided)), 0644)
			if err != nil {
				continue
			}
		}

		time.Sleep(time.Second * 5)
	}
}

func readPowerSensordShared() (SensorData, error) {

	const sharedMemName = "AIDA64_SensorValues"

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	openFileMapping := kernel32.NewProc("OpenFileMappingW")
	mapViewOfFile := kernel32.NewProc("MapViewOfFile")
	unmapViewOfFile := kernel32.NewProc("UnmapViewOfFile")
	closeHandle := kernel32.NewProc("CloseHandle")

	nameUTF16, err := syscall.UTF16PtrFromString(sharedMemName)
	if err != nil {
		return SensorData{}, fmt.Errorf("failed to convert name to UTF-16: %w", err)
	}

	handle, _, err := openFileMapping.Call(
		syscall.FILE_MAP_READ,
		0,
		uintptr(unsafe.Pointer(nameUTF16)),
	)
	defer closeHandle.Call(handle)
	if err != nil && err.Error() != "The operation completed successfully." {
		return SensorData{}, fmt.Errorf("failed to open file mapping: %w", err)
	}

	if handle == 0 {
		return SensorData{}, fmt.Errorf("failed to open file mapping: %w", syscall.GetLastError())
	}

	dataPtr, _, err := mapViewOfFile.Call(
		handle,
		syscall.FILE_MAP_READ,
		0,
		0,
		0,
	)
	defer unmapViewOfFile.Call(dataPtr)

	if err != nil && err.Error() != "The operation completed successfully." {
		return SensorData{}, fmt.Errorf("failed to access data pointer: %w", err)
	}

	if dataPtr == 0 {
		return SensorData{}, fmt.Errorf("failed to map view of file: %w", syscall.GetLastError())
	}

	var data []byte

	for i := 0; ; i++ {
		byte := *(*byte)(unsafe.Pointer(dataPtr + uintptr(i)))
		if byte == 0 {
			break
		}
		data = append(data, byte)
	}

	sensorData := SensorData{}

	decoder := xml.NewDecoder(bytes.NewReader(data))

	for {
		token, err := decoder.Token()
		if err != nil {
			break // End of XML stream
		}

		switch t := token.(type) {
		case xml.StartElement:

			if t.Name.Local != "fan" && t.Name.Local != "volt" && t.Name.Local != "pwr" && t.Name.Local != "curr" {
				continue // Skip elements that are not relevant to us
			}

			var sensor Sensor
			err = decoder.DecodeElement(&sensor, &t)
			if err != nil {
				fmt.Println("Error decoding element:%w", err)
				continue
			}

			switch t.Name.Local {
			case "fan":
				sensorData.Fans = append(sensorData.Fans, sensor)
			case "volt":
				sensorData.Voltages = append(sensorData.Voltages, sensor)
			case "pwr":
				sensorData.Powers = append(sensorData.Powers, sensor)
			case "curr":
				sensorData.Currents = append(sensorData.Currents, sensor)
			}
		}
	}

	return sensorData, nil
}

func getWritablePath(fileName string) (string, error) {
	// First try creating file next to executable
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		filePath := filepath.Join(exeDir, fmt.Sprintf("%s.sensor", fileName))
		if _, err := os.Stat(filePath); err == nil || os.IsNotExist(err) {
			return filePath, nil
		}
	}

	// Fall back to LOCALAPPDATA if executable path fails
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		return "", fmt.Errorf("LOCALAPPDATA environment variable not set")
	}

	appDir := filepath.Join(appData, "PowerSensor")
	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		return "", err
	}

	return filepath.Join(appDir, fmt.Sprintf("%s.sensor", fileName)), nil
}
