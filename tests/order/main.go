package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/franco-bianco/go-hyperliquid/hyperliquid/hyperliquid"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	godotenv.Load()

	ws := hyperliquid.NewWebSocketAPI(true)
	err := ws.Connect()
	if err != nil {
		log.Fatalf("failed to connect: %s", err)
	}
	defer ws.Disconnect()

	userAddress := os.Getenv("VAULT_ADDRESS")
	err = ws.SubscribeToOrderUpdates(userAddress, func(orders []hyperliquid.WsOrder) {
		data, _ := json.Marshal(orders)
		fmt.Println(string(data))
	})

	if err != nil {
		log.Fatalf("failed to subscribe: %s", err)
	}

	log.Info("listening for order updates...")
	log.Info("press CTRL+C to exit")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nshutting down...")
}
