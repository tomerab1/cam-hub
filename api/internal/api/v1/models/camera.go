package models

type Camera struct {
	UUID            string `json:"uuid" db:"id"`
	CameraName      string `json:"camera_name" db:"name"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmware_version"`
	SerialNumber    string `json:"serial_number"`
	HardwareId      string `json:"hardware_id"`
	Addr            string `json:"addr"`
}
