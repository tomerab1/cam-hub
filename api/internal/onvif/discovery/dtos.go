package discovery

type WsDiscoveryMatch struct {
	UUID  string `json:"uuid,omitempty"`
	Xaddr string `json:"addr,omitempty"`
}

type WsDiscoveryDto struct {
	Matches []WsDiscoveryMatch `json:"matches"`
}
