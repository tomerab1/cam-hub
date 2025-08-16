package onvif

import (
	"github.com/use-go/onvif/device"
)

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

func (client *OnvifClient) GetDatetime() *SystemDateTimeDto {
	resp, err := client.device.CallMethod(device.GetSystemDateAndTime{})

	if err != nil {
		client.logger.Error(err.Error())
		return nil
	}

	var parsed systemDateTimeResp
	parseResp(resp, &parsed, client.logger)

	dto := &SystemDateTimeDto{
		DateTimeType:    parsed.DateTimeType,
		DaylightSavings: parsed.DaylightSavings,
		TimeZoneTZ:      parsed.TimeZoneTZ,
	}

	dto.UTC.Year = parsed.UTCYear
	dto.UTC.Month = parsed.UTCMonth
	dto.UTC.Day = parsed.UTCDay
	dto.UTC.Hour = parsed.UTCHour
	dto.UTC.Minute = parsed.UTCMinute
	dto.UTC.Second = parsed.UTCSecond

	dto.Local.Year = parsed.LocalYear
	dto.Local.Month = parsed.LocalMonth
	dto.Local.Day = parsed.LocalDay
	dto.Local.Hour = parsed.LocalHour
	dto.Local.Minute = parsed.LocalMinute
	dto.Local.Second = parsed.LocalSecond

	return dto

}
