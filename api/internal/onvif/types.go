package onvif

type systemDateTimeResp struct {
	DateTimeType    string `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>DateTimeType"`
	DaylightSavings bool   `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>DaylightSavings"`
	TimeZoneTZ      string `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>TimeZone>TZ"`

	UTCYear   int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>UTCDateTime>Date>Year"`
	UTCMonth  int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>UTCDateTime>Date>Month"`
	UTCDay    int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>UTCDateTime>Date>Day"`
	UTCHour   int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>UTCDateTime>Time>Hour"`
	UTCMinute int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>UTCDateTime>Time>Minute"`
	UTCSecond int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>UTCDateTime>Time>Second"`

	LocalYear   int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>LocalDateTime>Date>Year"`
	LocalMonth  int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>LocalDateTime>Date>Month"`
	LocalDay    int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>LocalDateTime>Date>Day"`
	LocalHour   int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>LocalDateTime>Time>Hour"`
	LocalMinute int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>LocalDateTime>Time>Minute"`
	LocalSecond int `xml:"Body>GetSystemDateAndTimeResponse>SystemDateAndTime>LocalDateTime>Time>Second"`
}

type wsDiscoveryResp struct {
	Matches []struct {
		Match struct {
			UUID  string `xml:"EndpointReference>Address"`
			Xaddr string `xml:"XAddrs"`
		} `xml:"ProbeMatch"`
	} `xml:"Body>ProbeMatches"`
}
