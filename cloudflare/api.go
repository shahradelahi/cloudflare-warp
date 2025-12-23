package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/shahradelahi/cloudflare-warp/cloudflare/model"
	"github.com/shahradelahi/cloudflare-warp/cloudflare/network"
)

const (
	apiBase string = "https://api.cloudflareclient.com/v0a1922"
)

func defaultHeaders() map[string]string {
	return map[string]string{
		"Content-Type":      "application/json; charset=UTF-8",
		"User-Agent":        "okhttp/3.12.1",
		"CF-Client-Version": "a-6.30-3596",
	}
}

type WarpAPI struct {
	client *http.Client
}

func NewWarpAPI() *WarpAPI {
	tlsDialer := network.Dialer{}
	// Create a custom HTTP transport
	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return tlsDialer.TLSDial(network, addr)
		},
	}

	return &WarpAPI{
		client: &http.Client{Transport: transport},
	}
}

func (w *WarpAPI) request(method, url, authToken string, body interface{}, out interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}

	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}

	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}

	return nil
}

func (w *WarpAPI) GetAccount(authToken, deviceID string) (model.IdentityAccount, error) {
	var rspData model.IdentityAccount
	err := w.request("GET", fmt.Sprintf("%s/reg/%s/account", apiBase, deviceID), authToken, nil, &rspData)
	return rspData, err
}

func (w *WarpAPI) GetBoundDevices(authToken, deviceID string) ([]model.IdentityDevice, error) {
	var rspData []model.IdentityDevice
	err := w.request("GET", fmt.Sprintf("%s/reg/%s/account/devices", apiBase, deviceID), authToken, nil, &rspData)
	return rspData, err
}

func (w *WarpAPI) GetSourceBoundDevice(authToken, deviceID string) (*model.IdentityDevice, error) {
	devices, err := w.GetBoundDevices(authToken, deviceID)
	if err != nil {
		return nil, err
	}

	for _, device := range devices {
		if device.ID == deviceID {
			return &device, nil
		}
	}

	return nil, errors.New("no matching bound device found")
}

func (w *WarpAPI) GetSourceDevice(authToken, deviceID string) (model.Identity, error) {
	var rspData model.Identity
	err := w.request("GET", fmt.Sprintf("%s/reg/%s", apiBase, deviceID), authToken, nil, &rspData)
	return rspData, err
}

func (w *WarpAPI) Register(publicKey string) (model.Identity, error) {
	data := map[string]interface{}{
		"install_id":   "",
		"fcm_token":    "",
		"tos":          time.Now().Format(time.RFC3339Nano),
		"key":          publicKey,
		"type":         "Android",
		"model":        "PC",
		"locale":       "en_US",
		"warp_enabled": true,
	}

	var rspData model.Identity
	err := w.request("POST", fmt.Sprintf("%s/reg", apiBase), "", data, &rspData)
	return rspData, err
}

func (w *WarpAPI) ResetAccountLicense(authToken, deviceID string) (model.License, error) {
	var rspData model.License
	err := w.request("POST", fmt.Sprintf("%s/reg/%s/account/license", apiBase, deviceID), authToken, nil, &rspData)
	return rspData, err
}

func (w *WarpAPI) UpdateAccount(authToken, deviceID, license string) (model.IdentityAccount, error) {
	var rspData model.IdentityAccount
	err := w.request("PUT", fmt.Sprintf("%s/reg/%s/account", apiBase, deviceID), authToken, map[string]interface{}{"license": license}, &rspData)
	return rspData, err
}

func (w *WarpAPI) UpdateBoundDevice(authToken, deviceID, otherDeviceID, name string, active bool) (model.IdentityDevice, error) {
	data := map[string]interface{}{
		"active": active,
		"name":   name,
	}

	var rspData model.IdentityDevice
	err := w.request("PATCH", fmt.Sprintf("%s/reg/%s/account/devices/%s", apiBase, deviceID, otherDeviceID), authToken, data, &rspData)
	return rspData, err
}

func (w *WarpAPI) UpdateSourceDevice(authToken, deviceID string, data map[string]interface{}) (model.Identity, error) {
	var rspData model.Identity
	err := w.request("PATCH", fmt.Sprintf("%s/reg/%s", apiBase, deviceID), authToken, data, &rspData)
	return rspData, err
}

func (w *WarpAPI) DeleteDevice(authToken, deviceID string) error {
	return w.request("DELETE", fmt.Sprintf("%s/reg/%s", apiBase, deviceID), authToken, nil, nil)
}
