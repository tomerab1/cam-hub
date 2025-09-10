package onvif

import (
	"github.com/IOTechSystems/onvif/ptz"
	"github.com/IOTechSystems/onvif/xsd"
	"github.com/IOTechSystems/onvif/xsd/onvif"
	dto "tomerab.com/cam-hub/internal/onvif/ptz"
)

func (client *OnvifClient) MoveCamera(moveDto dto.MoveCameraDto) error {
	tok := onvif.ReferenceToken(moveDto.Token)
	timeout := xsd.Duration("PT1S")
	resp, err := client.device.CallMethod(ptz.ContinuousMove{
		ProfileToken: &tok,
		Velocity: &onvif.PTZSpeed{
			PanTilt: &onvif.Vector2D{X: moveDto.Translation.X, Y: moveDto.Translation.Y},
		},
		Timeout: &timeout,
	})
	if err != nil {
		return err
	}

	var relativeMoveResp ptz.RelativeMoveResponse
	if err := parseResp(resp, &relativeMoveResp); err != nil {
		return err
	}

	return nil
}
