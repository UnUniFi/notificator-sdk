package notificator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/tendermint/tendermint/abci/types"
	tmClient "github.com/tendermint/tendermint/rpc/client/http"
	tmTypes "github.com/tendermint/tendermint/types"

	"github.com/gorilla/mux"
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
	router := mux.NewRouter()
	router.HandleFunc("/email-address", postEmailAddressHandlerFactory(notificator)).Methods("OPTIONS")
	router.HandleFunc("/email-address", postEmailAddressHandlerFactory(notificator)).Methods("POST")

	http.Handle("/", router)
	http.ListenAndServe(fmt.Sprintf(":%d", notificator.Config.Port), nil)
	ctx := context.Background()
	// github.com/tendermint/tendermint v0.34.20
	client, err := tmClient.New(notificator.Config.TendermintRpcHost, "websocket")

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
	// health check
	// query := "tm.event = 'NewBlock'"
	query := "tm.event='Tx'"
	out, err := client.Subscribe(ctx, "test", query, 1000)
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

type PostEmailAddressRequest struct {
	BlockchainAddress string `json:"blockchain_address"`
	EmailAddress      string `json:"email_address"`
}

func postEmailAddressHandlerFactory(notificator Notificator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if r.Method != "POST" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.Header.Get("Content-Type") != "application/json" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		length, err := strconv.Atoi(r.Header.Get("Content-Length"))
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		body := make([]byte, length)
		length, err = r.Body.Read(body)
		if err != nil && err != io.EOF {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var req PostEmailAddressRequest
		err = json.Unmarshal(body[:length], &req)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		fmt.Printf("%v\n", req)

		err = notificator.SetEmailAddress(req.BlockchainAddress, req.EmailAddress)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte{})
		return
	}
}
