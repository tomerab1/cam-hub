package onvif

import (
	"github.com/IOTechSystems/onvif/ptz"
	"github.com/IOTechSystems/onvif/xsd"
	"github.com/IOTechSystems/onvif/xsd/onvif"
	dto "tomerab.com/cam-hub/internal/onvif/ptz"
	"tomerab.com/cam-hub/internal/utils"
)

func (client *OnvifClient) MoveCamera(moveDto dto.MoveCameraDto) error {
	tok := onvif.ReferenceToken(moveDto.Token)
	timeout := xsd.Duration("PT1S")
	resp, err := client.device.CallMethod(ptz.ContinuousMove{
		ProfileToken: &tok,
		Velocity: &onvif.PTZSpeed{
			PanTilt: optVec2D(moveDto.Translation),
			Zoom:    optVec1D(moveDto.Zoom),
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

func optVec2D(vec *utils.Vec2D) *onvif.Vector2D {
	if vec == nil {
		return nil
	}
	return &onvif.Vector2D{
		X: float64(vec.X),
		Y: float64(vec.Y),
	}
}

func optVec1D(f *float32) *onvif.Vector1D {
	if f == nil {
		return nil
	}

	return &onvif.Vector1D{X: float64(*f)}
}
