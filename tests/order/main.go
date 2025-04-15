package main

import (
	"os"

	"github.com/franco-bianco/go-hyperliquid/hyperliquid/hyperliquid"
	"github.com/joho/godotenv"
	"github.com/k0kubun/pp"
	"github.com/sirupsen/logrus"
)

func main() {
	godotenv.Load()

	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		ForceColors:     true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	apiWalletPrivateKey := os.Getenv("API_WALLET_PRIVATE_KEY")
	vaultAddress := os.Getenv("VAULT_ADDRESS")

	hyperliquidClient := hyperliquid.NewHyperliquid(&hyperliquid.HyperliquidClientConfig{
		IsMainnet:      true,
		PrivateKey:     apiWalletPrivateKey,
		AccountAddress: vaultAddress,
	})
	hyperliquidClient.SetDebugActive()

	orderRes, err := hyperliquidClient.ExchangeAPI.Order(hyperliquid.OrderRequest{
		Coin:    "ETH",
		IsBuy:   false,
		Sz:      0.1,
		LimitPx: 1700,
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifGtc,
			},
		},
	}, hyperliquid.GroupingNa)
	if err != nil {
		log.Fatalf("error placing order: %s", err)
	}

	pp.Println(orderRes)
}
