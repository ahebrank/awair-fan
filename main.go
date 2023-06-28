package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/ahebrank/awair-fan/awair"
)

type Config struct {
	AwairUrl string
	VocLimit int
	PmLimit  int
	FanTime  int
}

func main() {
	var conf Config
	_, err := toml.DecodeFile("conf.toml", &conf)
	if err != nil {
		log.Fatal(err)
	}

	sensorClient := awair.NewClient(conf.AwairUrl)

	data, err := sensorClient.Get()
	if err != nil {
		log.Fatal(err)
	}

	// Check limits
	if data.VOC > conf.VocLimit || data.PM25 > conf.PmLimit {
		fmt.Println("Fan should go on.")
	}
}
