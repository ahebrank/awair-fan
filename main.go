package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/ahebrank/awair-fan/awair"
	"github.com/ahebrank/awair-fan/ecobee"
)

type Config struct {
	AwairUrl           string
	VocLimit           int
	PmLimit            int
	CO2Limit           int
	FanTime            int
	EcobeeBaseUrl      string
	EcobeeApiKey       string
	EcobeeRefreshToken string
}

func main() {
	var conf Config
	_, err := toml.DecodeFile("conf.toml", &conf)
	if err != nil {
		log.Fatal(err)
	}

	fanClient := ecobee.NewClient(conf.EcobeeBaseUrl, conf.EcobeeApiKey, conf.EcobeeRefreshToken)
	status, err := fanClient.Status()
	if err != nil {
		log.Fatal(err)
	}
	if status.EquipmentStatus != "" && status.EquipmentStatus != "fan" {
		log.Fatal("equipment already running")
	}

	sensorClient := awair.NewClient(conf.AwairUrl)

	data, err := sensorClient.Get()
	if err != nil {
		log.Fatal(err)
	}

	// Check limits
	if data.CO2 > conf.CO2Limit || data.VOC > conf.VocLimit || data.PM25 > conf.PmLimit {
		fmt.Printf("Fan on. CO2: %d; VOC: %d; PM25: %d\n", data.CO2, data.VOC, data.PM25)
		// Only hold if fan not already running.
		if status.EquipmentStatus == "" {
			err = fanClient.FanOn(*status, status.Time, 30)
			if err != nil {
				log.Fatal(err)
			}
		}
	} else {
		// Only resume if the fan is the only thing running.
		if status.EquipmentStatus == "fan" {
			err = fanClient.Resume()
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
