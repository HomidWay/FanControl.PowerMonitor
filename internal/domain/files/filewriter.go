package files

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileWritter struct{}

func NewFileWritter() *FileWritter {
	return &FileWritter{}
}

func (f *FileWritter) WirteSensorFile(fileName string, data []byte) error {

	err := writeAtExePath(fileName, data)
	if err == nil {
		return nil
	}

	err = writeAtAppData(fileName, data)
	if err != nil {
		return err
	}

	return nil
}

func writeAtExePath(fileName string, data []byte) error {
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		filePath := filepath.Join(exeDir, fmt.Sprintf("%s.sensor", fileName))
		if _, err := os.Stat(filePath); err != nil && !os.IsNotExist(err) {
			return err
		}

		err = os.WriteFile(
			filePath,
			data,
			0644,
		)

		if err != nil {
			return err
		}

		return nil
	}
}

func writeAtAppData(fileName string, data []byte) error {
	appData := os.Getenv("LOCALAPPDATA")
	if appData == "" {
		return fmt.Errorf("LOCALAPPDATA environment variable not set")
	}

	appDir := filepath.Join(appData, "PowerSensor")
	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		return err
	}

	err = os.WriteFile(
		filepath.Join(appDir, fmt.Sprintf("%s.sensor", fileName)),
		data,
		0644,
	)

	if err != nil {
		return err
	}

	return nil
}