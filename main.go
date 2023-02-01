package main

var appName = "notificator"

func main() {
	config, err := LoadConfig()
	if err != nil {
		return
	}
	notificator := NewNotificator(config)
}
