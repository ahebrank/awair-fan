package awair

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type AwairData struct {
	Timestamp     time.Time `json:"timestamp"`
	Score         float64   `json:"score"`
	DewPoint      float64   `json:"dew_point"`
	Temp          float64   `json:"temp"`
	Humid         float64   `json:"humid"`
	CO2           int       `json:"co2"`
	VOC           int       `json:"voc"`
	VOCBaseline   int       `json:"voc_baseline"`
	VOCH2Raw      int       `json:"voc_h2_raw"`
	VOCEthanolRaw int       `json:"voc_ethanol_raw"`
	PM25          int       `json:"pm25"`
	PM10Est       int       `json:"pm10_est"`
}

// Client contains the Awair's LAN IP for making requests to the
// local API
type Client struct {
	Url *url.URL
}

// NewClient returns a new Awair local client
func NewClient(awairUrl string) *Client {
	parsedUrl, err := url.Parse(awairUrl)
	if err != nil {
		return nil
	}
	return &Client{
		Url: parsedUrl,
	}
}

// Latest data from the Awair local HTTP server
func (c *Client) Get() (*AwairData, error) {
	req := http.Request{
		Method: "GET",
		URL:    c.Url,
	}
	var data AwairData
	err := getJSON(&req, &data)
	if err != nil {
		return nil, err
	}
	return &data, nil
}

func getJSON(req *http.Request, output interface{}) error {
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if statusCode := res.StatusCode; statusCode < http.StatusOK || statusCode > 299 {
		return fmt.Errorf("non-200 returned from remote")
	}
	if err := json.NewDecoder(res.Body).Decode(output); err != nil {
		return err
	}
	return nil
}
