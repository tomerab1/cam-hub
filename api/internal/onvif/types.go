package onvif

type GetNetworkInterfacesResp struct {
	Token string `xml:"Body>GetNetworkInterfacesResponse>NetworkInterfaces>token"`
}
