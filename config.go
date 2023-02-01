package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Config struct {
	RestHost          string `json:"rest_host"`
	TendermintRpcHost string `json:"tendermint_rpc_host"`
	MailgunDomain     string `json:"mailgun_domain"`
	MailgunApiKey     string `json:"mailgun_api_key"`
	MailgunSender     string `json:"mailgun_sender"`
}

func LoadConfig(appName string) (*Config, error) {
	path := os.ExpandEnv(fmt.Sprintf("$HOME/.%s/config.json", appName))
	bz, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = json.Unmarshal(bz, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
