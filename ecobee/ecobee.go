package ecobee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

type NewToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
}

type EcobeeResponse struct {
	Thermostats []EcobeeThermostat `json:"thermostatList"`
}

type EcobeeThermostat struct {
	Name            string        `json:"name"`
	Runtime         EcobeeRuntime `json:"runtime"`
	EquipmentStatus string        `json:"equipmentStatus"`
	Time            string        `json:"thermostatTime"`
}

type EcobeeRuntime struct {
	ActualTemp        int    `json:"actualTemperature"`
	ActualHumidity    int    `json:"actualHumidity"`
	DesiredHeat       int    `json:"desiredHeat"`
	DesiredCool       int    `json:"desiredCool"`
	DesiredHumidity   int    `json:"desiredHumidity"`
	DesiredDehumidity int    `json:"desiredDehumidity"`
	DesiredFanMode    string `json:"desiredFanMode"`
}

type Client struct {
	BaseUrl      string
	ApiKey       string
	RefreshToken string
	AccessToken  string
	RefreshAt    int64
}

func NewClient(baseUrl string, apiKey string, refreshToken string) *Client {
	return &Client{
		BaseUrl:      baseUrl,
		ApiKey:       apiKey,
		RefreshToken: refreshToken,
		AccessToken:  "",
		RefreshAt:    0,
	}
}

func (c *Client) Refresh() error {
	if time.Now().Unix() < c.RefreshAt {
		return nil
	}

	query := map[string]string{
		"grant_type": "refresh_token",
		"code":       c.RefreshToken,
		"client_id":  c.ApiKey,
	}

	var data NewToken
	err := c.Get("token", query, &data, false)
	if err != nil {
		return err
	}

	c.AccessToken = data.AccessToken
	c.RefreshAt = time.Now().Unix() + data.ExpiresIn - 30
	return nil
}

func (c *Client) Get(endpoint string, params map[string]string, output interface{}, refreshCheck bool) error {
	if refreshCheck {
		err := c.Refresh()
		if err != nil {
			return err
		}
	}

	req, err := http.NewRequest(http.MethodGet, c.BaseUrl+endpoint, nil)
	if err != nil {
		return err
	}
	q := url.Values{}
	for key, val := range params {
		q.Add(key, val)
	}
	req.URL.RawQuery = q.Encode()
	req.Header = http.Header{
		"Authorization": {"Bearer " + c.AccessToken},
		"Content-type":  {"text/json"},
		"User-Agent":    {"AwairFan 1.0"},
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if statusCode := res.StatusCode; statusCode < http.StatusOK || statusCode > 299 {
		reqDump, _ := httputil.DumpRequestOut(req, true)
		resDump, _ := httputil.DumpResponse(res, true)
		return fmt.Errorf("status %d returned from remote with request %s and response %s", statusCode, reqDump, resDump)
	}
	if err := json.NewDecoder(res.Body).Decode(output); err != nil {
		return err
	}

	return nil
}

func (c *Client) Post(endpoint string, data string) error {
	err := c.Refresh()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseUrl+endpoint, bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	q := url.Values{}
	q.Add("format", "json")
	req.URL.RawQuery = q.Encode()
	req.Header = http.Header{
		"Authorization": {"Bearer " + c.AccessToken},
		"Content-type":  {"application/json"},
		"User-Agent":    {"AwairFan 1.0"},
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if statusCode := res.StatusCode; statusCode < http.StatusOK || statusCode > 299 {
		resDump, _ := httputil.DumpResponse(res, true)
		return fmt.Errorf("status %d returned from remote with response: %s", statusCode, resDump)
	}

	return nil
}

func (c *Client) Status() (*EcobeeThermostat, error) {
	query := map[string]string{
		"format": "json",
		"body":   "{\"selection\":{\"selectionType\":\"registered\",\"selectionMatch\":\"actual\",\"includeRuntime\":true, \"includeEquipmentStatus\":true}}",
	}

	var data EcobeeResponse
	err := c.Get("1/thermostat", query, &data, true)
	if err != nil {
		return nil, err
	}
	// @TODO: deal with more than one thermostat
	return &data.Thermostats[0], nil
}

func (c *Client) FanOn(thermostat EcobeeThermostat, currentTime string, holdMins int) error {
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, currentTime)
	if err != nil {
		log.Fatal(err)
	}
	t = t.Add(time.Minute * time.Duration(holdMins))
	body := `
		{
			"selection": {
				"selectionType": "registered",
				"selectionMatch": ""
			},
			"functions": [
				{
					"type": "setHold",
					"params": {
						"holdType":"dateTime",
						"endDate": "2006-01-02",
						"endTime": "15:04:05",
						"heatHoldTemp": "999",
						"coolHoldTemp": "000",
						"fan": "on"
					}
				}
			]
		}`
	body = strings.Replace(body, "2006-01-02", t.Format("2006-01-02"), 1)
	body = strings.Replace(body, "15:04:05", t.Format("15:04:05"), 1)
	body = strings.Replace(body, "999", fmt.Sprint(thermostat.Runtime.DesiredHeat), 1)
	body = strings.Replace(body, "000", fmt.Sprint(thermostat.Runtime.DesiredCool), 1)

	err = c.Post("1/thermostat", body)
	return err
}

func (c *Client) Resume() error {
	body := `
		{
			"selection": {
				"selectionType": "registered",
				"selectionMatch": ""
			},
			"functions": [
				{
					"type": "resumeProgram",
					"params": {
						"resumeAll": true
					}
				}
			]
		}`
	err := c.Post("1/thermostat", body)
	return err
}
