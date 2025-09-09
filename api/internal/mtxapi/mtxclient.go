package mtxapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"tomerab.com/cam-hub/internal/repos"
)

const (
	RtspPort       = 554
	FFMPEGTemplate = `/usr/bin/ffmpeg -loglevel warning -rtsp_transport tcp -i rtsp://%s:%s@%s:%d/channel=1_stream=0.sdp?real_stream -map 0:v -map 0:a? -c:v libx264 -pix_fmt yuv420p -profile:v baseline -level:v 3.1 -preset veryfast -tune zerolatency -g 60 -keyint_min 60 -sc_threshold 0 -c:a libopus -ar 48000 -ac 2 -b:a 64k -f rtsp -rtsp_transport tcp rtsp://%s:$RTSP_PORT/$MTX_PATH`
)

type MtxClient struct {
	HttpClient   *http.Client
	Logger       *slog.Logger
	CamRepo      repos.CameraRepoIface
	CamCredsRepo repos.CameraCredsRepoIface
	Cache        repos.RedisIface
}

type camDetails struct {
	addr     string
	username string
	password string
}

type addPathRequest struct {
	RunOnDemand             string `json:"runOnDemand"`
	RunOnDemandRestart      bool   `json:"runOnDemandRestart"`
	RunOnDemandStartTimeout string `json:"runOnDemandStartTimeout"`
	RunOnDemandCloseAfter   string `json:"runOnDemandCloseAfter"`
}

type MtxErrorDto struct {
	Error string `json:"error"`
}

func (client *MtxClient) Publish(ctx context.Context, uuid string) (string, error) {
	whepURL := fmt.Sprintf("http://%s:8889/%s/whep", os.Getenv("MEDIAMTX_HOST"), uuid)
	if client.doesStreamExists(ctx, uuid) {
		return whepURL, nil
	}

	onDemandCmd, err := client.getRunOnDemandCmd(ctx, uuid)
	if err != nil {
		return "", err
	}

	reqBody := addPathRequest{
		RunOnDemand:             onDemandCmd,
		RunOnDemandRestart:      true,
		RunOnDemandStartTimeout: "15s",
		RunOnDemandCloseAfter:   "15s",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	mediaMtxUrl := fmt.Sprintf("%s/v3/config/paths/add/%s", os.Getenv("MEDIAMTX_ADDR"), uuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mediaMtxUrl, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var mtxErr MtxErrorDto
		json.NewDecoder(resp.Body).Decode(&mtxErr)
		return "", fmt.Errorf("mediamtx returned %s: %s", resp.Status, mtxErr.Error)
	}

	return whepURL, nil
}

func (client *MtxClient) doesStreamExists(ctx context.Context, uuid string) bool {
	mediaMtxUrl := fmt.Sprintf("%s/v3/config/paths/get/%s", os.Getenv("MEDIAMTX_ADDR"), uuid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaMtxUrl, nil)
	if err != nil {
		client.Logger.Error(err.Error())
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.HttpClient.Do(req)
	if err != nil {
		client.Logger.Error(err.Error())
		return false
	}

	fmt.Println("here", resp.StatusCode)
	return resp.StatusCode == http.StatusOK
}

func (client *MtxClient) getCameraDetails(ctx context.Context, uuid string) (*camDetails, error) {
	cam, err := client.CamRepo.FindOne(ctx, uuid)
	if err != nil {
		return nil, err
	}

	cleanAddr := strings.Split(cam.Addr, ":")
	addr := cleanAddr[0]
	creds, err := client.CamCredsRepo.FindOne(ctx, uuid)
	if err != nil {
		return nil, nil
	}

	return &camDetails{
		addr:     addr,
		username: creds.Username,
		password: creds.Password,
	}, nil
}

func (client *MtxClient) getRunOnDemandCmd(ctx context.Context, uuid string) (string, error) {
	details, err := client.getCameraDetails(ctx, uuid)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf(FFMPEGTemplate, details.username, details.password, details.addr, RtspPort, "127.0.0.1"), nil
}
