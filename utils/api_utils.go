package utils

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smlsynctodede/config"
	"time"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func GetFullAPIURL(serviceName string) string {
	for _, part := range config.PartServices {
		if part.ServiceName == serviceName {
			return config.AppConfig.API.BaseURL + part.PartName
		}
	}
	return ""
}

func SendDataToAPI(serviceName string, apiKey string, data interface{}) ([]byte, error) {
	apiURL := GetFullAPIURL(serviceName)
	if apiURL == "" {
		return nil, fmt.Errorf("unknown service name: %s", serviceName)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("error marshalling data: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return body, fmt.Errorf("API authentication failed (401): %s", string(body))
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return body, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
