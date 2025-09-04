package v1

import "time"

type PairDeviceReq struct {
	UUID       string `json:"uuid"`
	Addr       string `json:"addr"`
	CameraName string `json:"camera_name"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type UnpairDeviceReq struct {
	UUID string `json:"uuid"`
}

type DiscoveryEvent struct {
	Type string    `json:"type"`
	UUID string    `json:"uuid"`
	Addr string    `json:"addr"`
	At   time.Time `json:"at"`
}
