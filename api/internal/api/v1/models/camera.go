package models

type Camera struct {
	UUID            string `json:"uuid" db:"id"`
	CameraName      string `json:"camera_name" db:"name"`
	Manufacturer    string `json:"manufacturer" db:"manufacturer"`
	Model           string `json:"model" db:"model"`
	FirmwareVersion string `json:"firmware_version" db:"firmwareversion"`
	SerialNumber    string `json:"serial_number" db:"serialnumber"`
	HardwareId      string `json:"hardware_id" db:"hardwareid"`
	Addr            string `json:"addr" db:"addr"`
	IsPaired        bool   `json:"is_paired" db:"ispaired"`
}
