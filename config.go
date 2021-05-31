package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var config struct {
	API struct {
		Host string
		Port string
	}
	SQL struct {
		Host     string
		Port     string
		User     string
		Password string
	}
}

func loadConfig() error {
	f, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		return err
	}
	return yaml.Unmarshal(f, &config)
}

func createConfigFile() {
	if _, err := os.Stat("config.yaml"); err == nil {
		fmt.Println("the 'config.yaml' already exists, do you really want to overwrite? (y/N)")
		var rsp string
		fmt.Scan(&rsp)
		if strings.ToLower(rsp) == "y" {
			writeFile()
		}
		return
	}
	writeFile()
}

func writeFile() {
	b, _ := yaml.Marshal(config)
	ioutil.WriteFile("config.yaml", b, 0766)
}
