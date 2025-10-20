package dvripclient

import (
	"encoding/json"
	"fmt"

	"github.com/xaionaro-go/go2rtc/pkg/dvrip"
)

const (
	cmdConfigSet              = 1040
	cmdConfigGet              = 1042
	cmdDelUser                = 1486
	dialTemplate              = "dvrip://%s:%s@%s:34567"
	ModeAuto        LightMode = "Auto"
	ModeNone        LightMode = "None"
	ModeIntelligent LightMode = "Intelligent"
)

type LightMode = string

type DvripClient struct {
	client dvrip.Client
}

type LightModeParams struct {
	Mode     LightMode
	Duration int // Only appiled to 'Intelligent' mode
}

type networkInfo struct {
	hostHex    string
	gatewayHex string
	submaskHex string
}

func New(camAddr, username, pass string) (*DvripClient, error) {
	var cli DvripClient
	fmt.Println(camAddr, username, pass)
	if err := cli.client.Dial(fmt.Sprintf(dialTemplate, username, pass, camAddr)); err != nil {
		_ = cli.client.Close()
		return nil, fmt.Errorf("dial: %w", err)
	}

	return &cli, nil
}

func (cli *DvripClient) DelUser(username string) error {
	body, err := json.Marshal(map[string]any{"Name": username})
	if err != nil {
		return err
	}

	if _, err := cli.client.WriteCmd(cmdDelUser, body); err != nil {
		return err
	}

	_, err = cli.client.ReadJSON()
	return err
}

func (cli *DvripClient) Get(node string) (map[string]any, error) {
	body, err := json.Marshal(map[string]any{"Name": node})
	if err != nil {
		return nil, err
	}

	if _, err := cli.client.WriteCmd(cmdConfigGet, body); err != nil {
		return nil, err
	}

	resp, err := cli.client.ReadJSON()
	if err != nil {
		return nil, err
	}

	if m, ok := resp[node].(map[string]any); ok {
		return m, nil
	}

	return nil, fmt.Errorf("node (%s) can not be found", node)
}

func (cli *DvripClient) Set(node string, obj map[string]any) error {
	body, err := json.Marshal(map[string]any{
		"Name": node,
		node:   obj,
	})
	if err != nil {
		return err
	}

	if _, err := cli.client.WriteCmd(cmdConfigSet, body); err != nil {
		return err
	}

	_, err = cli.client.ReadJSON()
	return err
}

func (cli *DvripClient) Close() error {
	return cli.client.Close()
}

func (cli *DvripClient) SetLightMode(params LightModeParams) error {
	switch params.Mode {
	case ModeAuto:
		fallthrough
	case ModeNone:
		cli.Set("Camera.WhiteLight", map[string]any{
			"WorkMode": params.Mode,
		})
	case ModeIntelligent:
		cli.Set("Camera.WhiteLight", map[string]any{
			"WorkMode": params.Mode,
			"MoveTrigLight": map[string]any{
				"Duration": params.Duration,
			},
		})
	}

	return nil
}

func (cli *DvripClient) getNetworkInfo() (*networkInfo, error) {
	node, err := cli.Get("NetWork.NetCommon")
	if err != nil {
		return nil, err
	}

	var hostHex, gatewayHex, submaskHex string
	if val, ok := node["HostIP"].(string); ok {
		hostHex = val
	}
	if val, ok := node["GateWay"].(string); ok {
		gatewayHex = val
	}
	if val, ok := node["Submask"].(string); ok {
		submaskHex = val
	}

	return &networkInfo{
		hostHex:    hostHex,
		gatewayHex: gatewayHex,
		submaskHex: submaskHex,
	}, nil
}

func (cli *DvripClient) PairWifi(ssid, psk string) error {
	node, err := cli.Get("NetWork.Wifi")
	if err != nil {
		return err
	}

	netInfo, err := cli.getNetworkInfo()
	if err != nil {
		return err
	}

	if err := cli.setWifiPairingInfo(node, ssid, psk, netInfo); err != nil {
		return err
	}

	return cli.verifyWifiPairing(ssid, psk, netInfo)
}

func (cli *DvripClient) verifyWifiPairing(
	ssid,
	psk string,
	netInfo *networkInfo,
) error {
	node, err := cli.Get("NetWork.Wifi")
	if err != nil {
		return err
	}

	val := node["SSID"] == ssid &&
		node["Keys"] == psk &&
		node["HostIP"] == netInfo.hostHex &&
		node["GateWay"] == netInfo.gatewayHex &&
		node["Submask"] == netInfo.submaskHex

	if !val {
		return fmt.Errorf("dvrip: Wifi pairing failed")
	}

	return nil
}

func (cli *DvripClient) setWifiPairingInfo(
	node map[string]any,
	ssid,
	psk string,
	netInfo *networkInfo,
) error {
	node["Enable"] = true
	node["SSID"] = ssid
	node["Keys"] = psk
	node["HostIP"] = netInfo.hostHex
	node["GateWay"] = netInfo.gatewayHex
	node["Submask"] = netInfo.submaskHex

	return cli.Set("NetWork.Wifi", node)
}
