package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
)

const (
	//ROOT default directory
	ROOT string = "/tmp/mywork"
	//WD default output directory, clear speration from ROOT
	WD string = "/home"
	//EXT default extension of target files
	EXT string = ".go"
	//Target default output file name
	Target string = "merge"
)

//Config configuration
type Config struct {
	Root   string `json:"root"`
	WD     string `json:"work directory"`
	EXT    string `json:"extension"`
	Target string `json:"target file"`
}

func loadConfig(file string) (Config, error) {
	var config Config

	j, err := ioutil.ReadFile(file)
	if err != nil {
		return config, err
	}

	config = Config{}
	err = json.Unmarshal(j, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func defaultConfig() Config {
	ctg := Config{
		Root:   ROOT,
		WD:     WD,
		EXT:    EXT,
		Target: Target,
	}
	jb, err := json.Marshal(ctg)
	if err != nil {
		log.Println(err)
	} else {
		_, err := fmt.Print(string(jb))
		if err != nil {
			log.Println(err)
		}
	}
	return ctg
}
