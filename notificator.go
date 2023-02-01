package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tendermint/tendermint/abci/types"
	tmClient "github.com/tendermint/tendermint/rpc/client/http"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

type Notificator struct {
	Config        Config
	DB            *leveldb.DB
	EventHandlers map[string]func(attributes []types.EventAttribute) error
}

func NewNotificator(appName string, config Config) (*Notificator, error) {
	path := os.ExpandEnv(fmt.Sprintf("$HOME/.%s/db", appName))
	db, err := leveldb.OpenFile(path, &opt.Options{})
	if err != nil {
		return nil, err
	}

	return &Notificator{
		Config:        config,
		DB:            db,
		EventHandlers: make(map[string]func(attributes []types.EventAttribute) error),
	}, nil
}

func (notificator Notificator) Close() {
	notificator.DB.Close()
}

func (notificator Notificator) RegisterEventHandler(event string, handler func(atributes []types.EventAttribute) error) {
	notificator.EventHandlers[event] = handler
}

func (notificator Notificator) Start() {
	client, err := tmClient.New(notificator.Config.TendermintRpcHost)
	if err != nil {
		fmt.Printf("failed to initialize a client: %s\n", err)
		os.Exit(1)
	}

	if err := client.Start(); err != nil {
		fmt.Printf("failed to start a client: %s\n", err)
		os.Exit(1)
	}

	defer client.Stop() //nolint:errcheck

	// Subscribe to all tendermint transactions
	query := "tm.event = 'Tx'"
	out, err := client.Subscribe(context.Background(), "test", query, 1000)
	if err != nil {
		fmt.Printf("failed to subscribe to query: %s\n", err)
		os.Exit(1)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	for {
		select {
		case result := <-out:
			tx, ok := result.Data.(tmTypes.EventDataTx)
			if !ok {
				fmt.Printf("new tx: error while extracting event data from new tx\n")
			}
			// Iterate over each event in the transaction
			for _, event := range tx.Result.Events {
				eventType := event.GetType()

				if notificator.EventHandlers[eventType] == nil {
					continue
				}

				err := notificator.EventHandlers[eventType](event.GetAttributes())

				if err != nil {
					fmt.Printf("error in event handler: %s\n", err.Error())
				}
			}
		case <-quit:
			os.Exit(0)
		}
	}
}
