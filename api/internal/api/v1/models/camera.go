package models

// Camera represents a physical camera device and its associated metadata.
// It contains identification, hardware details, and network information.
//
// Fields:
//   - UUID: Unique identifier for the camera, stored as "id" in the database
//   - CameraName: User-friendly name of the camera, stored as "name" in the database
//   - Manufacturer: Name of the camera manufacturer
//   - Model: Camera model designation
//   - FirmwareVersion: Current version of the camera's firmware
//   - SerialNumber: Unique serial number assigned by manufacturer
//   - HardwareId: Hardware identifier of the camera
//   - Addr: Network address of the camera
//   - Version: Version number for record tracking
type Camera struct {
	UUID            string `json:"uuid" db:"id"`
	CameraName      string `json:"camera_name" db:"name"`
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmware_version"`
	SerialNumber    string `json:"serial_number"`
	HardwareId      string `json:"hardware_id"`
	Addr            string `json:"addr"`
	Version         int    `json:"version"`
}
