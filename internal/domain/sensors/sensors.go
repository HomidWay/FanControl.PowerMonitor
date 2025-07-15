package sensors

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

type SensorDataFetcher interface {
	GetSensorData() (SensorData, error)
}
