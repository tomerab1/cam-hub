package onvif

import (
	"github.com/use-go/onvif/device"
)

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
