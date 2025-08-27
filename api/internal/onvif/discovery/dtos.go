package discovery

type WsDiscoveryMatch struct {
	UUID  string `json:"uuid"`
	Xaddr string `json:"addr"`
}

type WsDiscoveryDto struct {
	Matches []WsDiscoveryMatch `json:"matches"`
}
