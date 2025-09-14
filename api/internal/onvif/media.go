package onvif

import (
	"errors"

	"github.com/IOTechSystems/onvif/media"
	"github.com/IOTechSystems/onvif/xsd/onvif"
	"tomerab.com/cam-hub/internal/utils"
)

var (
	ErrNoPtz = errors.New("this device does not support ptz")
)

type PtzProfileToken struct {
	Token string
}

func (client *OnvifClient) GetPtzProfile() (*PtzProfileToken, error) {
	resp, err := client.device.CallMethod(media.GetProfiles{})
	if err != nil {
		return nil, err
	}

	var getProfilesResp media.GetProfilesResponse
	if err := parseResp(resp, &getProfilesResp); err != nil {
		return nil, err
	}

	firstElem, err := utils.FindFirstIdx(getProfilesResp.Profiles, func(idx int, profile onvif.Profile) bool {
		return profile.PTZConfiguration != nil
	})

	if err != nil {
		return nil, ErrNoPtz
	}

	return &PtzProfileToken{
		Token: string(getProfilesResp.Profiles[firstElem].Token),
	}, nil
}
