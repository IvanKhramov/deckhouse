/*
Copyright 2021 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package proxy

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

type CrowdClient struct {
	apiURL   string
	login    string
	password string

	allowedGroups map[string]struct{}
	httpClient    *http.Client
}

func NewCrowdClient(apiURL, login, password string, allowedGroups []string) *CrowdClient {
	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
		},
	}

	groups := make(map[string]struct{})
	for _, group := range allowedGroups {
		groups[group] = struct{}{}
	}

	return &CrowdClient{
		apiURL:        strings.TrimSuffix(apiURL, "/"),
		login:         login,
		password:      password,
		allowedGroups: groups,
		httpClient:    client,
	}
}

func (c *CrowdClient) MakeRequest(url, method string, jsonPayload interface{}) (string, error) {
	var body io.Reader
	if jsonPayload != nil {
		jsonData, err := json.Marshal(jsonPayload)
		if err != nil {
			return "", fmt.Errorf("crowd request error: %+v", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, fmt.Sprintf("%s/rest/usermanagement/1%s", c.apiURL, url), body)
	if err != nil {
		return "", fmt.Errorf("crowd request error: %+v", err)
	}

	req.SetBasicAuth(c.login, c.password)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("crowd request error: %+v", err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("crowd request error: %v", err)
	}

	if (resp.StatusCode != http.StatusOK) && (resp.StatusCode != http.StatusCreated) {
		return "", fmt.Errorf("crowd request was not successful: %v %v", resp.StatusCode, string(responseBody))
	}

	return string(responseBody), nil
}

func (c *CrowdClient) GetGroups(body string) ([]string, error) {
	var crowdGroups struct {
		Groups []struct{ Name string } `json:"groups"`
	}
	var groups []string

	if err := json.Unmarshal([]byte(body), &crowdGroups); err != nil {
		return groups, err
	}

	for _, value := range crowdGroups.Groups {
		if len(c.allowedGroups) > 0 {
			if _, ok := c.allowedGroups[value.Name]; ok {
				groups = append(groups, value.Name)
			}
		} else {
			groups = append(groups, value.Name)
		}
	}
	return groups, nil
}
