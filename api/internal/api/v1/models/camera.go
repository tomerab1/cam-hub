package models

type Camera struct {
	UUID            string `json:"uuid"`
	CameraName      string `json:"camera_name"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmware_version"`
	SerialNumber    string `json:"serial_number"`
	HardwareId      string `json:"hardware_id"`
	Addr            string `json:"addr"`
}
