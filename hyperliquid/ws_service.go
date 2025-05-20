package hyperliquid

import (
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// IWebSocketAPI is an interface for the WebSocket service
type IWebSocketAPI interface {
	IClient // Base client interface

	// Connection methods
	Connect() error
	Disconnect() error
	IsConnected() bool

	// Subscription methods
	Subscribe(subscription Subscription, callback func(data interface{})) error
	Unsubscribe(subscription Subscription) error

	// Post request methods
	Post(requestType string, payload interface{}) (interface{}, error)
}

// WebSocketAPI is the default implementation of the IWebSocketAPI interface
type WebSocketAPI struct {
	Client
	conn         *websocket.Conn
	wsURL        string
	connected    bool
	handlers     map[string]func(data interface{})
	postHandlers map[int]chan interface{}
	idCounter    atomic.Int32
	mu           sync.RWMutex
	connMu       sync.Mutex
	done         chan struct{}
}

// NewWebSocketAPI returns a new instance of the WebSocketAPI struct
func NewWebSocketAPI(isMainnet bool) *WebSocketAPI {
	api := WebSocketAPI{
		Client:       *NewClient(isMainnet),
		connected:    false,
		handlers:     make(map[string]func(data interface{})),
		postHandlers: make(map[int]chan interface{}),
		done:         make(chan struct{}),
	}

	if isMainnet {
		api.wsURL = MAINNET_WS_URL
	} else {
		api.wsURL = TESTNET_WS_URL
	}

	return &api
}

// Endpoint implements the IAPIService interface
func (api *WebSocketAPI) Endpoint() string {
	return ""
}

// Connect establishes a connection to the WebSocket server
func (api *WebSocketAPI) Connect() error {
	api.connMu.Lock()
	defer api.connMu.Unlock()

	if api.connected {
		return nil
	}

	api.debug("connecting to %s", api.wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(api.wsURL, nil)
	if err != nil {
		api.debug("error connecting to websocket: %s", err)
		return err
	}

	api.conn = conn
	api.connected = true
	api.done = make(chan struct{})

	go api.readLoop()

	return nil
}

// Disconnect closes the connection to the WebSocket server
func (api *WebSocketAPI) Disconnect() error {
	api.connMu.Lock()
	defer api.connMu.Unlock()

	if !api.connected {
		return nil
	}

	close(api.done)
	api.debug("disconnecting from websocket")

	api.mu.Lock()
	api.handlers = make(map[string]func(data interface{}))
	for id, ch := range api.postHandlers {
		close(ch)
		delete(api.postHandlers, id)
	}
	api.mu.Unlock()

	err := api.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		api.debug("error sending close message: %s", err)
	}

	time.Sleep(100 * time.Millisecond)
	err = api.conn.Close()
	if err != nil {
		api.debug("error closing websocket: %s", err)
		return err
	}

	api.connected = false
	return nil
}

// IsConnected returns true if the client is connected to the WebSocket server
func (api *WebSocketAPI) IsConnected() bool {
	api.connMu.Lock()
	defer api.connMu.Unlock()
	return api.connected
}

// Subscribe subscribes to a WebSocket feed
func (api *WebSocketAPI) Subscribe(subscription Subscription, callback func(data interface{})) error {
	if !api.IsConnected() {
		err := api.Connect()
		if err != nil {
			return err
		}
	}

	api.mu.Lock()
	channelKey := subscription.Type
	if subscription.User != "" {
		channelKey = fmt.Sprintf("%s-%s", channelKey, subscription.User)
	}
	if subscription.Coin != "" {
		channelKey = fmt.Sprintf("%s-%s", channelKey, subscription.Coin)
	}
	if subscription.Interval != "" {
		channelKey = fmt.Sprintf("%s-%s", channelKey, subscription.Interval)
	}
	api.handlers[channelKey] = callback
	api.mu.Unlock()

	subMsg := SubscriptionMessage{
		Method:       "subscribe",
		Subscription: subscription,
	}

	return api.sendMessage(subMsg)
}

// Unsubscribe unsubscribes from a WebSocket feed
func (api *WebSocketAPI) Unsubscribe(subscription Subscription) error {
	if !api.IsConnected() {
		return fmt.Errorf("not connected")
	}

	api.mu.Lock()
	channelKey := subscription.Type
	if subscription.User != "" {
		channelKey = fmt.Sprintf("%s-%s", channelKey, subscription.User)
	}
	if subscription.Coin != "" {
		channelKey = fmt.Sprintf("%s-%s", channelKey, subscription.Coin)
	}
	if subscription.Interval != "" {
		channelKey = fmt.Sprintf("%s-%s", channelKey, subscription.Interval)
	}
	delete(api.handlers, channelKey)
	api.mu.Unlock()

	unsubMsg := UnsubscriptionMessage{
		Method:       "unsubscribe",
		Subscription: subscription,
	}

	return api.sendMessage(unsubMsg)
}

// Post sends a post request over WebSocket
func (api *WebSocketAPI) Post(requestType string, payload interface{}) (interface{}, error) {
	if !api.IsConnected() {
		err := api.Connect()
		if err != nil {
			return nil, err
		}
	}

	id := int(api.idCounter.Add(1))
	responseChan := make(chan interface{}, 1)

	api.mu.Lock()
	api.postHandlers[id] = responseChan
	api.mu.Unlock()

	defer func() {
		api.mu.Lock()
		delete(api.postHandlers, id)
		api.mu.Unlock()
	}()

	postMsg := PostMessage{
		Method: "post",
		ID:     id,
		Request: PostRequest{
			Type:    requestType,
			Payload: payload,
		},
	}

	err := api.sendMessage(postMsg)
	if err != nil {
		return nil, err
	}

	select {
	case response := <-responseChan:
		// Check if the response is an error
		if err, ok := response.(error); ok {
			return nil, err
		}
		return response, nil
	case <-time.After(15 * time.Second):
		return nil, fmt.Errorf("request timeout")
	}
}

// readLoop reads messages from the WebSocket connection and processes them
func (api *WebSocketAPI) readLoop() {
	defer func() {
		api.connMu.Lock()
		api.connected = false
		api.connMu.Unlock()
	}()

	for {
		select {
		case <-api.done:
			return
		default:
			_, message, err := api.conn.ReadMessage()
			if err != nil {
				api.debug("error reading message: %s", err)
				return
			}

			api.processMessage(message)
		}
	}
}

// processMessage processes incoming WebSocket messages
func (api *WebSocketAPI) processMessage(message []byte) {
	var response WsResponse
	err := json.Unmarshal(message, &response)
	if err != nil {
		api.debug("error unmarshaling message: %s", err)
		return
	}

	fmt.Println(string(message))

	if response.Channel == "post" {
		var postResponseData map[string]interface{}
		err = json.Unmarshal(message, &postResponseData)
		if err != nil {
			api.debug("error unmarshaling post response: %s", err)
			return
		}

		data, ok := postResponseData["data"].(map[string]interface{})
		if !ok {
			api.debug("invalid post response data format")
			return
		}

		id, ok := data["id"].(float64)
		if !ok {
			api.debug("invalid post response id format")
			return
		}

		responseObj, ok := data["response"].(map[string]interface{})
		if !ok {
			api.debug("invalid post response object format")
			return
		}

		respType, _ := responseObj["type"].(string)

		api.mu.RLock()
		ch, ok := api.postHandlers[int(id)]
		api.mu.RUnlock()

		if ok {
			if respType == "error" {
				errMsg, _ := responseObj["payload"].(string)
				if errMsg == "" {
					errMsg = "unknown error"
				}
				ch <- fmt.Errorf("%s", errMsg)
			} else {
				payload := responseObj["payload"]
				ch <- payload
			}
		}
		return
	}

	if response.Channel == "subscriptionResponse" {
		api.debug("subscription confirmed for channel: %s", response.Channel)
		return
	}

	channelKey := response.Channel
	var dataMap map[string]interface{}
	jsonData, _ := json.Marshal(response.Data)
	json.Unmarshal(jsonData, &dataMap)

	if coin, ok := dataMap["coin"].(string); ok {
		channelKey = fmt.Sprintf("%s-%s", channelKey, coin)
	}
	if user, ok := dataMap["user"].(string); ok {
		channelKey = fmt.Sprintf("%s-%s", channelKey, user)
	}

	switch response.Channel {
	case "orderUpdates":
		var orders []WsOrder
		jsonData, _ := json.Marshal(response.Data)
		json.Unmarshal(jsonData, &orders)

		api.mu.RLock()
		handler, ok := api.handlers[channelKey]

		if !ok {
			prefixToMatch := "orderUpdates-"
			for hKey, h := range api.handlers {
				if len(hKey) > len(prefixToMatch) && hKey[:len(prefixToMatch)] == prefixToMatch {
					handler = h
					ok = true
					break
				}
			}
		}

		if ok {
			handler(orders)
			api.mu.RUnlock()
			return
		}
		api.mu.RUnlock()
	}

	api.mu.RLock()
	handler, ok := api.handlers[channelKey]
	api.mu.RUnlock()

	if ok {
		handler(response.Data)
	}
}

// sendMessage sends a message over the WebSocket connection
func (api *WebSocketAPI) sendMessage(message interface{}) error {
	api.connMu.Lock()
	defer api.connMu.Unlock()

	if !api.connected {
		return fmt.Errorf("not connected")
	}

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	api.debug("sending message: %s", string(data))
	return api.conn.WriteMessage(websocket.TextMessage, data)
}

// SubscribeToAllMids subscribes to all mids
func (api *WebSocketAPI) SubscribeToAllMids(callback func(data AllMids)) error {
	return api.Subscribe(Subscription{Type: "allMids"}, func(data interface{}) {
		var mids AllMids
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &mids)
		callback(mids)
	})
}

// SubscribeToNotification subscribes to notifications for a user
func (api *WebSocketAPI) SubscribeToNotification(address string, callback func(data Notification)) error {
	return api.Subscribe(Subscription{Type: "notification", User: address}, func(data interface{}) {
		var notification Notification
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &notification)
		callback(notification)
	})
}

// SubscribeToCandle subscribes to candle updates for a specific coin and interval
func (api *WebSocketAPI) SubscribeToCandle(coin string, interval string, callback func(data []Candle)) error {
	return api.Subscribe(Subscription{Type: "candle", Coin: coin, Interval: interval}, func(data interface{}) {
		var candles []Candle
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &candles)
		callback(candles)
	})
}

// SubscribeToL2Book subscribes to order book updates for a specific coin
func (api *WebSocketAPI) SubscribeToL2Book(coin string, callback func(data WsBook)) error {
	return api.Subscribe(Subscription{Type: "l2Book", Coin: coin}, func(data interface{}) {
		var book WsBook
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &book)
		callback(book)
	})
}

// SubscribeToTrades subscribes to trades for a specific coin
func (api *WebSocketAPI) SubscribeToTrades(coin string, callback func(data []WsTrade)) error {
	return api.Subscribe(Subscription{Type: "trades", Coin: coin}, func(data interface{}) {
		var trades []WsTrade
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &trades)
		callback(trades)
	})
}

// SubscribeToOrderUpdates subscribes to order updates for a specific user
func (api *WebSocketAPI) SubscribeToOrderUpdates(address string, callback func(data []WsOrder)) error {
	return api.Subscribe(Subscription{Type: "orderUpdates", User: address}, func(data interface{}) {
		var orders []WsOrder
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &orders)
		callback(orders)
	})
}

// SubscribeToUserEvents subscribes to user events for a specific user
func (api *WebSocketAPI) SubscribeToUserEvents(address string, callback func(data WsUserEvent)) error {
	return api.Subscribe(Subscription{Type: "userEvents", User: address}, func(data interface{}) {
		var events WsUserEvent
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &events)
		callback(events)
	})
}

// SubscribeToUserFills subscribes to user fills for a specific user
func (api *WebSocketAPI) SubscribeToUserFills(address string, aggregateByTime bool, callback func(data WsUserFills)) error {
	return api.Subscribe(Subscription{Type: "userFills", User: address, AggregateByTime: aggregateByTime}, func(data interface{}) {
		var fills WsUserFills
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &fills)
		callback(fills)
	})
}

// SubscribeToUserFundings subscribes to user fundings for a specific user
func (api *WebSocketAPI) SubscribeToUserFundings(address string, callback func(data WsUserFundings)) error {
	return api.Subscribe(Subscription{Type: "userFundings", User: address}, func(data interface{}) {
		var fundings WsUserFundings
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &fundings)
		callback(fundings)
	})
}

// SubscribeToUserNonFundingLedgerUpdates subscribes to user non-funding ledger updates for a specific user
func (api *WebSocketAPI) SubscribeToUserNonFundingLedgerUpdates(address string, callback func(data WsUserNonFundingLedgerUpdates)) error {
	return api.Subscribe(Subscription{Type: "userNonFundingLedgerUpdates", User: address}, func(data interface{}) {
		var updates WsUserNonFundingLedgerUpdates
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &updates)
		callback(updates)
	})
}

// SubscribeToActiveAssetCtx subscribes to active asset context for a specific coin
func (api *WebSocketAPI) SubscribeToActiveAssetCtx(coin string, callback func(data interface{})) error {
	return api.Subscribe(Subscription{Type: "activeAssetCtx", Coin: coin}, func(data interface{}) {
		callback(data)
	})
}

// SubscribeToActiveAssetData subscribes to active asset data for a specific user and coin
func (api *WebSocketAPI) SubscribeToActiveAssetData(address string, coin string, callback func(data WsActiveAssetData)) error {
	return api.Subscribe(Subscription{Type: "activeAssetData", User: address, Coin: coin}, func(data interface{}) {
		var assetData WsActiveAssetData
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &assetData)
		callback(assetData)
	})
}

// SubscribeToUserTwapSliceFills subscribes to user TWAP slice fills for a specific user
func (api *WebSocketAPI) SubscribeToUserTwapSliceFills(address string, callback func(data WsUserTwapSliceFills)) error {
	return api.Subscribe(Subscription{Type: "userTwapSliceFills", User: address}, func(data interface{}) {
		var twapFills WsUserTwapSliceFills
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &twapFills)
		callback(twapFills)
	})
}

// SubscribeToUserTwapHistory subscribes to user TWAP history for a specific user
func (api *WebSocketAPI) SubscribeToUserTwapHistory(address string, callback func(data WsUserTwapHistory)) error {
	return api.Subscribe(Subscription{Type: "userTwapHistory", User: address}, func(data interface{}) {
		var twapHistory WsUserTwapHistory
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &twapHistory)
		callback(twapHistory)
	})
}

// SubscribeToBbo subscribes to BBO for a specific coin
func (api *WebSocketAPI) SubscribeToBbo(coin string, callback func(data WsBbo)) error {
	return api.Subscribe(Subscription{Type: "bbo", Coin: coin}, func(data interface{}) {
		var bbo WsBbo
		jsonData, _ := json.Marshal(data)
		json.Unmarshal(jsonData, &bbo)
		callback(bbo)
	})
}
