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

func (w *WarpAPI) GetAccount(authToken, deviceID string) (model.IdentityAccount, error) {
	reqUrl := fmt.Sprintf("%s/reg/%s/account", apiBase, deviceID)
	method := "GET"

	req, err := http.NewRequest(method, reqUrl, nil)
	if err != nil {
		return model.IdentityAccount{}, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return model.IdentityAccount{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.IdentityAccount{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.IdentityAccount{}, err
	}

	var rspData = model.IdentityAccount{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return model.IdentityAccount{}, err
	}

	return rspData, nil
}

func (w *WarpAPI) GetBoundDevices(authToken, deviceID string) ([]model.IdentityDevice, error) {
	reqUrl := fmt.Sprintf("%s/reg/%s/account/devices", apiBase, deviceID)
	method := "GET"

	req, err := http.NewRequest(method, reqUrl, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rspData = []model.IdentityDevice{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return nil, err
	}

	return rspData, nil
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
	reqUrl := fmt.Sprintf("%s/reg/%s", apiBase, deviceID)
	method := "GET"

	req, err := http.NewRequest(method, reqUrl, nil)
	if err != nil {
		return model.Identity{}, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return model.Identity{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.Identity{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Identity{}, err
	}

	var rspData = model.Identity{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return model.Identity{}, err
	}

	return rspData, nil
}

func (w *WarpAPI) Register(publicKey string) (model.Identity, error) {
	reqUrl := fmt.Sprintf("%s/reg", apiBase)
	method := "POST"

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

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return model.Identity{}, err
	}

	req, err := http.NewRequest(method, reqUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return model.Identity{}, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return model.Identity{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.Identity{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Identity{}, err
	}

	var rspData = model.Identity{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return model.Identity{}, err
	}

	return rspData, nil
}

func (w *WarpAPI) ResetAccountLicense(authToken, deviceID string) (model.License, error) {
	reqUrl := fmt.Sprintf("%s/reg/%s/account/license", apiBase, deviceID)
	method := "POST"

	req, err := http.NewRequest(method, reqUrl, nil)
	if err != nil {
		return model.License{}, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return model.License{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.License{}, fmt.Errorf("API request failed with response: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.License{}, err
	}

	var rspData = model.License{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return model.License{}, err
	}

	return rspData, nil
}

func (w *WarpAPI) UpdateAccount(authToken, deviceID, license string) (model.IdentityAccount, error) {
	reqUrl := fmt.Sprintf("%s/reg/%s/account", apiBase, deviceID)
	method := "PUT"

	jsonBody, err := json.Marshal(map[string]interface{}{"license": license})
	if err != nil {
		return model.IdentityAccount{}, err
	}

	req, err := http.NewRequest(method, reqUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return model.IdentityAccount{}, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return model.IdentityAccount{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.IdentityAccount{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.IdentityAccount{}, err
	}

	var rspData = model.IdentityAccount{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return model.IdentityAccount{}, err
	}

	return rspData, nil
}

func (w *WarpAPI) UpdateBoundDevice(authToken, deviceID, otherDeviceID, name string, active bool) (model.IdentityDevice, error) {
	reqUrl := fmt.Sprintf("%s/reg/%s/account/devices/%s", apiBase, deviceID, otherDeviceID)
	method := "PATCH"

	data := map[string]interface{}{
		"active": active,
		"name":   name,
	}

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return model.IdentityDevice{}, err
	}

	req, err := http.NewRequest(method, reqUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return model.IdentityDevice{}, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return model.IdentityDevice{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.IdentityDevice{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.IdentityDevice{}, err
	}

	var rspData = model.IdentityDevice{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return model.IdentityDevice{}, err
	}

	return rspData, nil
}

func (w *WarpAPI) UpdateSourceDevice(authToken, deviceID string, data map[string]interface{}) (model.Identity, error) {
	reqUrl := fmt.Sprintf("%s/reg/%s", apiBase, deviceID)
	method := "PATCH"

	jsonBody, err := json.Marshal(data)
	if err != nil {
		return model.Identity{}, err
	}

	req, err := http.NewRequest(method, reqUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return model.Identity{}, err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return model.Identity{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return model.Identity{}, fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	// convert response to byte array
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return model.Identity{}, err
	}

	var rspData = model.Identity{}
	if err := json.Unmarshal(responseData, &rspData); err != nil {
		return model.Identity{}, err
	}

	return rspData, nil
}

func (w *WarpAPI) DeleteDevice(authToken, deviceID string) error {
	reqUrl := fmt.Sprintf("%s/reg/%s", apiBase, deviceID)
	method := "DELETE"

	req, err := http.NewRequest(method, reqUrl, nil)
	if err != nil {
		return err
	}

	// Set headers
	for k, v := range defaultHeaders() {
		req.Header.Set(k, v)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)

	// Create HTTP client and execute request
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status: %s", resp.Status)
	}

	return nil
}
