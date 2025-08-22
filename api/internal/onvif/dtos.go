package onvif

type SystemDateTimeDto struct {
	DateTimeType    string `json:"datimeType"`
	DaylightSavings bool   `json:"daylingSavings"`
	TimeZoneTZ      string `json:"timezone"`

	UTC struct {
		Year   int `json:"year"`
		Month  int `json:"month"`
		Day    int `json:"day"`
		Hour   int `json:"hour"`
		Minute int `json:"minute"`
		Second int `json:"second"`
	} `json:"utc"`

	Local struct {
		Year   int `json:"year"`
		Month  int `json:"month"`
		Day    int `json:"day"`
		Hour   int `json:"hour"`
		Minute int `json:"minute"`
		Second int `json:"second"`
	} `json:"local"`
}

// Ws discovery
type WsDiscoveryMatch struct {
	UUID  string `json:"uuid"`
	Xaddr string `json:"addr"`
}

type WsDiscoveryDto struct {
	Matches []WsDiscoveryMatch `json:"matches"`
}
