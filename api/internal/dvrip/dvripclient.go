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
	cmdOpMachine              = 1450
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

func (cli *DvripClient) PairWifi(ssid, psk string) error {
	node, err := cli.Get("NetWork.Wifi")
	if err != nil {
		return fmt.Errorf("failed to get Wifi: %w", err)
	}

	if err := cli.setWifiPairingInfo(node, ssid, psk); err != nil {
		return fmt.Errorf("failed to set Wifi: %w", err)
	}

	return nil
}

func (cli *DvripClient) Reboot() error {
	payload := map[string]any{
		"Name": "OPMachine",
		"OPMachine": map[string]any{
			"Action": "Reboot",
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	if _, err = cli.client.WriteCmd(cmdOpMachine, b); err != nil {
		return err
	}

	resp, err := cli.client.ReadJSON()
	if err != nil {
		return fmt.Errorf("read reboot ack: %w", err)
	}

	if v, ok := resp["Ret"].(float64); ok && int(v) != 100 {
		return fmt.Errorf("reboot rejected, Ret=%d", int(v))
	}

	return nil
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
) error {
	node["Enable"] = true
	node["SSID"] = ssid
	node["Keys"] = psk

	return cli.Set("NetWork.Wifi", node)
}
