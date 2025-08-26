package onvif

import (
	"github.com/IOTechSystems/onvif/device"
	"github.com/IOTechSystems/onvif/xsd"
	"github.com/IOTechSystems/onvif/xsd/onvif"
	dev "tomerab.com/cam-hub/internal/onvif/device"
)

type DeviceInformation struct {
	Manufacturer    string `xml:"Manufacturer"`
	Model           string `xml:"Model"`
	FirmwareVersion string `xml:"FirmwareVersion"`
	SerialNumber    string `xml:"SerialNumber"`
	HardwareId      string `xml:"HardwareId"`
}

type getDeviceInfoResp struct {
	DeviceInformation `xml:"GetDeviceInformationResponse"`
}

func (client *OnvifClient) GetDeviceInfo() (dev.GetDeviceInfoDto, error) {
	resp, err := client.device.CallMethod(device.GetDeviceInformation{})

	if err != nil {
		client.logger.Error(err.Error())
		return dev.GetDeviceInfoDto{}, err
	}

	var devInfo getDeviceInfoResp
	if err = parseResp(resp, &devInfo); err != nil {
		client.logger.Error(err.Error())
		return dev.GetDeviceInfoDto{}, err
	}

	return dev.GetDeviceInfoDto{
		Manufacturer:    devInfo.Manufacturer,
		Model:           devInfo.Model,
		FirmwareVersion: devInfo.FirmwareVersion,
		SerialNumber:    devInfo.SerialNumber,
		HardwareId:      devInfo.HardwareId,
	}, nil
}

func (client *OnvifClient) CreateUser(createUserDto dev.CreateUserDto) error {
	lvl := onvif.UserLevel("User")
	resp, err := client.device.CallMethod(device.CreateUsers{
		User: []onvif.UserRequest{
			{
				Username:  createUserDto.Username,
				Password:  createUserDto.Password,
				UserLevel: &lvl,
			},
		},
	})

	if err != nil {
		return err
	}

	var createUsersResp device.CreateUsersResponse
	if err := parseResp(resp, &createUsersResp); err != nil {
		return err
	}

	return nil
}

func (client *OnvifClient) DeleteUser(deleteUserDto dev.DeleteUserDto) error {
	resp, err := client.device.CallMethod(device.DeleteUsers{
		Username: []xsd.String{xsd.String(deleteUserDto.Username)},
	})

	if err != nil {
		return err
	}

	var createUsersResp device.DeleteUsersResponse
	if err := parseResp(resp, &createUsersResp); err != nil {
		return err
	}

	return nil
}

func (client *OnvifClient) DemoteClient(demoteUserDto dev.DemoteUserDto) error {
	lvl := onvif.UserLevel("User")
	resp, err := client.device.CallMethod(device.SetUser{
		User: []onvif.UserRequest{
			{
				Username:  demoteUserDto.Username,
				Password:  demoteUserDto.Password,
				UserLevel: &lvl,
			},
		},
	})

	if err != nil {
		return err
	}

	var demoteUserResp device.SetUserResponse
	if err := parseResp(resp, &demoteUserResp); err != nil {
		return err
	}

	return nil
}
