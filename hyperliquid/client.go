package hyperliquid

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

// IClient is the interface that wraps the basic Requst method.
//
// Request method sends a POST request to the HyperLiquid API.
// IsMainnet method returns true if the client is connected to the mainnet.
// debug method enables debug mode.
// SetPrivateKey method sets the private key for the client.
type IClient interface {
	IAPIService
	SetPrivateKey(privateKey string) error
	SetAccountAddress(address string)
	AccountAddress() string
	SetDebugActive()
	IsMainnet() bool
}

// Client is the default implementation of the Client interface.
//
// It contains the base URL of the HyperLiquid API, the HTTP client, the debug mode,
// the network type, the private key, and the logger.
// The debug method prints the debug messages.
type Client struct {
	baseUrl        string       // Base URL of the HyperLiquid API
	privateKey     string       // Private key for the client
	defualtAddress string       // Default address for the client
	isMainnet      bool         // Network type
	Debug          bool         // Debug mode
	httpClient     *http.Client // HTTP client
	keyManager     *PKeyManager // Private key manager
	Logger         *log.Logger  // Logger for debug messages
}

// Returns the private key manager connected to the API.
func (client *Client) KeyManager() *PKeyManager {
	return client.keyManager
}

// getAPIURL returns the API URL based on the network type.
func getURL(isMainnet bool) string {
	if isMainnet {
		return MAINNET_API_URL
	} else {
		return TESTNET_API_URL
	}
}

// NewClient returns a new instance of the Client struct.
func NewClient(isMainnet bool) *Client {
	logger := log.New()
	logger.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		PadLevelText:  true,
	})
	logger.SetOutput(os.Stdout)
	logger.SetLevel(log.DebugLevel)
	return &Client{
		baseUrl:        getURL(isMainnet),
		httpClient:     http.DefaultClient,
		Debug:          false,
		isMainnet:      isMainnet,
		privateKey:     "",
		defualtAddress: "",
		Logger:         logger,
		keyManager:     nil,
	}
}

// debug prints the debug messages.
func (client *Client) debug(format string, v ...any) {
	if client.Debug {
		client.Logger.Debugf(format, v...)
	}
}

// SetPrivateKey sets the private key for the client.
func (client *Client) SetPrivateKey(privateKey string) error {
	privateKey = strings.TrimPrefix(privateKey, "0x")
	client.privateKey = privateKey
	var err error
	client.keyManager, err = NewPKeyManager(privateKey)
	return err
}

// Some methods need public address to gather info (from infoAPI).
// In case you use PKeyManager from API section https://app.hyperliquid.xyz/API
// Then you can use this method to set the address.
func (client *Client) SetAccountAddress(address string) {
	client.defualtAddress = address
}

// Returns the public address connected to the API.
func (client *Client) AccountAddress() string {
	return client.defualtAddress
}

// IsMainnet returns true if the client is connected to the mainnet.
func (client *Client) IsMainnet() bool {
	return client.isMainnet
}

// SetDebugActive enables debug mode.
func (client *Client) SetDebugActive() {
	client.Debug = true
}

// Request sends a POST request to the HyperLiquid API.
func (client *Client) Request(endpoint string, payload any) ([]byte, error) {
	endpoint = strings.TrimPrefix(endpoint, "/") // Remove leading slash if present
	url := fmt.Sprintf("%s/%s", client.baseUrl, endpoint)
	client.debug("Request to %s", url)
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		client.debug("Error json.Marshal: %s", err)
		return nil, err
	}
	client.debug("Request payload: %s", string(jsonPayload))
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		client.debug("Error json.Marshal: %s", err)
		return nil, err
	}
	request, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		client.debug("Error http.NewRequest: %s", err)
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(request)
	if err != nil {
		client.debug("Error client.httpClient.Do: %s", err)
		return nil, err
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := response.Body.Close()
		// Only overwrite the retured error if the original error was nil and an
		// error occurred while closing the body.
		if err == nil && cerr != nil {
			err = cerr
		}
	}()
	client.debug("response: %#v", response)
	client.debug("response body: %s", string(data))
	client.debug("response status code: %d", response.StatusCode)
	if response.StatusCode >= http.StatusBadRequest {
		// If the status code is 400 or greater, return an error
		return nil, APIError{Message: fmt.Sprintf("HTTP %d: %s", response.StatusCode, data)}
	}
	return data, nil
}
