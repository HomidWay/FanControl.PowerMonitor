package aida

import (
	"PowerReader/internal/domain/sensors"
	"bytes"
	"encoding/xml"
	"fmt"
	"syscall"
	"unsafe"
)

// var regPath = `Software\FinalWire\AIDA64\SensorValues`

type AidaSensorData struct {
	Fans     []AidaSensor `xml:"fan"`
	Voltages []AidaSensor `xml:"volt"`
	Currents []AidaSensor `xml:"curr"`
	Powers   []AidaSensor `xml:"pwr"`
}

type AidaSensor struct {
	ID    string  `xml:"id"`
	Label string  `xml:"label"`
	Value float64 `xml:"value"`
}

type ReadMode int

const (
	ReadModeSharedMemory ReadMode = iota
)

type AIDA struct {
	readMode ReadMode
}

func NewAida() *AIDA {
	return &AIDA{
		readMode: ReadModeSharedMemory,
	}
}

func (a *AIDA) GetSensorData() (sensors.SensorData, error) {

	switch a.readMode {
	case ReadModeSharedMemory:
		return readPowerSensordShared()
	default:
		return sensors.SensorData{}, fmt.Errorf("unknown read mode: %d", a.readMode)
	}
}

func readPowerSensordShared() (sensors.SensorData, error) {

	const sharedMemName = "AIDA64_SensorValues"

	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	openFileMapping := kernel32.NewProc("OpenFileMappingW")
	mapViewOfFile := kernel32.NewProc("MapViewOfFile")
	unmapViewOfFile := kernel32.NewProc("UnmapViewOfFile")
	closeHandle := kernel32.NewProc("CloseHandle")

	nameUTF16, err := syscall.UTF16PtrFromString(sharedMemName)
	if err != nil {
		return sensors.SensorData{}, fmt.Errorf("failed to convert name to UTF-16: %w", err)
	}

	handle, _, err := openFileMapping.Call(
		syscall.FILE_MAP_READ,
		0,
		uintptr(unsafe.Pointer(nameUTF16)),
	)
	defer closeHandle.Call(handle)
	if err != nil && err.Error() != "The operation completed successfully." {
		return sensors.SensorData{}, fmt.Errorf("failed to open file mapping: %w", err)
	}

	if handle == 0 {
		return sensors.SensorData{}, fmt.Errorf("failed to open file mapping: %w", syscall.GetLastError())
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
		return sensors.SensorData{}, fmt.Errorf("failed to access data pointer: %w", err)
	}

	if dataPtr == 0 {
		return sensors.SensorData{}, fmt.Errorf("failed to map view of file: %w", syscall.GetLastError())
	}

	var data []byte

	for i := 0; ; i++ {
		byte := *(*byte)(unsafe.Pointer(dataPtr + uintptr(i)))
		if byte == 0 {
			break
		}
		data = append(data, byte)
	}

	sensorData := sensors.SensorData{}

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

			var aidaSensor AidaSensor
			err = decoder.DecodeElement(&aidaSensor, &t)
			if err != nil {
				fmt.Println("Error decoding element:%w", err)
				continue
			}

			sensor := sensors.Sensor{
				ID:    aidaSensor.ID,
				Label: aidaSensor.Label,
				Value: aidaSensor.Value,
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
