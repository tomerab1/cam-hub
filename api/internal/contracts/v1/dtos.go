package v1

type PairDeviceReq struct {
	UUID       string `json:"uuid"`
	Addr       string `json:"addr"`
	CameraName string `json:"camera_name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}
