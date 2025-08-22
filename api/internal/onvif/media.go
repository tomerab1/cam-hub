package onvif

import (
	"encoding/json"

	"github.com/IOTechSystems/onvif/imaging"
	"github.com/IOTechSystems/onvif/xsd/onvif"
)

func (client *OnvifClient) GetProfiles() []byte {
	resp, err := client.device.CallMethod(imaging.SetImagingSettings{
		VideoSourceToken: onvif.ReferenceToken("V_SRC_000"),
	})

	if err != nil {
		client.logger.Error(err.Error())
	}

	var profiles imaging.GetImagingSettingsResponse
	parseResp(resp, &profiles, client.logger)

	b, err := json.Marshal(profiles)
	if err != nil {
		client.logger.Error(err.Error())
		return []byte{}
	}

	return b
}
