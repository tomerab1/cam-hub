package onvif

import (
	"fmt"

	"github.com/IOTechSystems/onvif/device"
)

func (client *OnvifClient) GetDatetime() SystemDateTimeDto {
	resp, err := client.device.CallMethod(device.GetSystemDateAndTime{})

	if err != nil {
		client.logger.Error(err.Error())
		return SystemDateTimeDto{}
	}

	var parsed device.GetSystemDateAndTimeResponse
	parseResp(resp, &parsed, client.logger)

	fmt.Println(parsed)

	return SystemDateTimeDto{}
}
