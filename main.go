package main

import (
	"flag"
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
	sensorOnlyFlag := flag.Bool("s", false, "Sensor output only, no fan trigger")
	confPathFlag := flag.String("c", "./conf.toml", "Path to configuration file")
	flag.Parse()

	var conf Config
	_, err := toml.DecodeFile(*confPathFlag, &conf)
	if err != nil {
		log.Fatal(err)
	}

	sensorClient := awair.NewClient(conf.AwairUrl)
	data, err := sensorClient.Get()
	if err != nil {
		log.Fatal(err)
	}

	fanClient := ecobee.NewClient(conf.EcobeeBaseUrl, conf.EcobeeApiKey, conf.EcobeeRefreshToken)
	status, err := fanClient.Status()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("CO2: %d (limit %d); VOC: %d (limit %d); PM25: %d (limit %d)\n", data.CO2, conf.CO2Limit, data.VOC, conf.VocLimit, data.PM25, conf.PmLimit)

	if status.EquipmentStatus != "" && status.EquipmentStatus != "fan" {
		log.Fatal("equipment already running")
	}

	if !*sensorOnlyFlag {
		// Check limits
		if data.CO2 > conf.CO2Limit || data.VOC > conf.VocLimit || data.PM25 > conf.PmLimit {
			fmt.Println("Fan on.")
			err = fanClient.FanOn(*status, status.Time, conf.FanTime)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			// Only resume if the fan is the only thing running.
			if status.EquipmentStatus == "fan" {
				fmt.Println("Resuming program.")
				err = fanClient.Resume()
				if err != nil {
					log.Fatal(err)
				}
			} else {
				log.Fatal("EquipmentStatus is not 'fan' only; skipping program resume.")
			}
		}
	}
}
