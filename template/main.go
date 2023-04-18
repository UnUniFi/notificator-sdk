package main

import (
	"fmt"

	sdk "github.com/UnUniFi/notificator-sdk"
	"github.com/tendermint/tendermint/abci/types"
)

var appName = "cosmos-notificator"

func main() {
	config, err := sdk.LoadConfig("./config.json")
	if err != nil {
		panic(err)
	}
	notificator, err := sdk.NewNotificator(appName, *config)
	if err != nil {
		panic(err)
	}
	defer notificator.Close()

	notificator.RegisterEventHandler(EventHogeType, handleEventHoge)

	notificator.Start()
}

const EventHogeType = "ununifi.nftmarket.EventListNft"

func handleEventHoge(attributes []types.EventAttribute) error {
	for _, attr := range attributes {
		fmt.Printf("attr.Key: %s, attr.Value: %s\n", attr.Key, attr.Value)
	}

	fmt.Println("hello world")
	return nil
}
