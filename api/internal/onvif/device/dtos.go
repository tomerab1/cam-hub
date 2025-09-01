package device

type GetDeviceInfoDto struct {
	Manufacturer    string `json:"manufacturer"`
	Model           string `json:"model"`
	FirmwareVersion string `json:"firmware_version"`
	SerialNumber    string `json:"serial_number"`
	HardwareId      string `json:"hardware_id"`
}

type CreateUserDto struct {
	Username  string `json:"username"`
	Password  string `json:"password"`
	UserLevel string `json:"user_level"` // enum { 'Administrator', 'Operator', 'User', 'Anonymous', 'Extended' }
}

type DemoteUserDto struct {
	Username string
	Password string
}

type DeleteUserDto struct {
	Username string `json:"username"`
}
