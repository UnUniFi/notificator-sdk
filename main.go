package main

var appName = "cosmos-notificator"

func main() {
	config, err := LoadConfig()
	if err != nil {
		panic(err)
	}
	notificator, err := NewNotificator(*config)
	if err != nil {
		panic(err)
	}
	defer notificator.Close()

	notificator.Start()
}
