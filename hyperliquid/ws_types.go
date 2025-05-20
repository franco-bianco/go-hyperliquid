package hyperliquid

// SubscriptionMessage represents a WebSocket subscription message
type SubscriptionMessage struct {
	Method       string       `json:"method"`
	Subscription Subscription `json:"subscription"`
}

// UnsubscriptionMessage represents a WebSocket unsubscription message
type UnsubscriptionMessage struct {
	Method       string       `json:"method"`
	Subscription Subscription `json:"subscription"`
}

// Subscription represents the subscription details
type Subscription struct {
	Type            string `json:"type"`
	User            string `json:"user,omitempty"`
	Coin            string `json:"coin,omitempty"`
	Interval        string `json:"interval,omitempty"`
	NSigFigs        int    `json:"nSigFigs,omitempty"`
	Mantissa        int    `json:"mantissa,omitempty"`
	AggregateByTime bool   `json:"aggregateByTime,omitempty"`
	Dex             string `json:"dex,omitempty"`
}

// PostMessage represents a WebSocket post message
type PostMessage struct {
	Method  string      `json:"method"`
	ID      int         `json:"id"`
	Request PostRequest `json:"request"`
}

// PostRequest represents the post request details
type PostRequest struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WsResponse represents a general WebSocket response
type WsResponse struct {
	Channel string      `json:"channel"`
	Data    interface{} `json:"data"`
}

// PostResponse represents a WebSocket post response
type PostResponse struct {
	ID       int              `json:"id"`
	Response PostResponseData `json:"response"`
}

// PostResponseData represents the post response data
type PostResponseData struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WsTrade represents a trade update
type WsTrade struct {
	Coin  string    `json:"coin"`
	Side  string    `json:"side"`
	Px    string    `json:"px"`
	Sz    string    `json:"sz"`
	Hash  string    `json:"hash"`
	Time  int64     `json:"time"`
	Tid   int64     `json:"tid"`
	Users [2]string `json:"users"`
}

// WsBook represents order book snapshot updates
type WsBook struct {
	Coin   string      `json:"coin"`
	Time   int64       `json:"time"`
	Levels [][]WsLevel `json:"levels"`
}

// WsBbo represents best bid and offer updates
type WsBbo struct {
	Coin string      `json:"coin"`
	Time int64       `json:"time"`
	Bbo  [2]*WsLevel `json:"bbo"`
}

// WsLevel represents a price level in the order book
type WsLevel struct {
	Px string `json:"px"`
	Sz string `json:"sz"`
	N  int    `json:"n"`
}

// Notification represents a notification message
type Notification struct {
	Notification string `json:"notification"`
}

// AllMids represents all mid prices
type AllMids struct {
	Mids map[string]string `json:"mids"`
}

// Candle represents a candle update
type Candle struct {
	T  int64   `json:"t"`
	T2 int64   `json:"T"`
	S  string  `json:"s"`
	I  string  `json:"i"`
	O  float64 `json:"o"`
	C  float64 `json:"c"`
	H  float64 `json:"h"`
	L  float64 `json:"l"`
	V  float64 `json:"v"`
	N  int     `json:"n"`
}

// WsUserEvent represents user events
type WsUserEvent struct {
	Fills         []WsFill          `json:"fills,omitempty"`
	Funding       *WsUserFunding    `json:"funding,omitempty"`
	Liquidation   *WsLiquidation    `json:"liquidation,omitempty"`
	NonUserCancel []WsNonUserCancel `json:"nonUserCancel,omitempty"`
}

// WsUserFills represents user fills
type WsUserFills struct {
	IsSnapshot bool     `json:"isSnapshot,omitempty"`
	User       string   `json:"user"`
	Fills      []WsFill `json:"fills"`
}

// WsFill represents a fill
type WsFill struct {
	Coin          string           `json:"coin"`
	Px            string           `json:"px"`
	Sz            string           `json:"sz"`
	Side          string           `json:"side"`
	Time          int64            `json:"time"`
	StartPosition string           `json:"startPosition"`
	Dir           string           `json:"dir"`
	ClosedPnl     string           `json:"closedPnl"`
	Hash          string           `json:"hash"`
	Oid           int64            `json:"oid"`
	Crossed       bool             `json:"crossed"`
	Fee           string           `json:"fee"`
	Tid           int64            `json:"tid"`
	Liquidation   *FillLiquidation `json:"liquidation,omitempty"`
	FeeToken      string           `json:"feeToken"`
	BuilderFee    string           `json:"builderFee,omitempty"`
}

// FillLiquidation represents liquidation information
type FillLiquidation struct {
	LiquidatedUser string  `json:"liquidatedUser,omitempty"`
	MarkPx         float64 `json:"markPx"`
	Method         string  `json:"method"`
}

// WsUserFunding represents funding information
type WsUserFunding struct {
	Time        int64  `json:"time"`
	Coin        string `json:"coin"`
	Usdc        string `json:"usdc"`
	Szi         string `json:"szi"`
	FundingRate string `json:"fundingRate"`
}

// WsLiquidation represents liquidation information
type WsLiquidation struct {
	Lid                    int64  `json:"lid"`
	Liquidator             string `json:"liquidator"`
	LiquidatedUser         string `json:"liquidated_user"`
	LiquidatedNtlPos       string `json:"liquidated_ntl_pos"`
	LiquidatedAccountValue string `json:"liquidated_account_value"`
}

// WsNonUserCancel represents non-user cancel information
type WsNonUserCancel struct {
	Coin string `json:"coin"`
	Oid  int64  `json:"oid"`
}

// WsOrder represents order information
type WsOrder struct {
	Order           WsBasicOrder `json:"order"`
	Status          string       `json:"status"`
	StatusTimestamp int64        `json:"statusTimestamp"`
}

// WsBasicOrder represents basic order information
type WsBasicOrder struct {
	Coin      string `json:"coin"`
	Side      string `json:"side"`
	LimitPx   string `json:"limitPx"`
	Sz        string `json:"sz"`
	Oid       int64  `json:"oid"`
	Timestamp int64  `json:"timestamp"`
	OrigSz    string `json:"origSz"`
	Cloid     string `json:"cloid,omitempty"`
}

// WsActiveAssetCtx represents active asset context
type WsActiveAssetCtx struct {
	Coin string        `json:"coin"`
	Ctx  PerpsAssetCtx `json:"ctx"`
}

// WsActiveSpotAssetCtx represents active spot asset context
type WsActiveSpotAssetCtx struct {
	Coin string       `json:"coin"`
	Ctx  SpotAssetCtx `json:"ctx"`
}

// SharedAssetCtx represents shared asset context properties
type SharedAssetCtx struct {
	DayNtlVlm float64 `json:"dayNtlVlm"`
	PrevDayPx float64 `json:"prevDayPx"`
	MarkPx    float64 `json:"markPx"`
	MidPx     float64 `json:"midPx,omitempty"`
}

// PerpsAssetCtx represents perpetuals asset context
type PerpsAssetCtx struct {
	SharedAssetCtx
	Funding      float64 `json:"funding"`
	OpenInterest float64 `json:"openInterest"`
	OraclePx     float64 `json:"oraclePx"`
}

// SpotAssetCtx represents spot asset context
type SpotAssetCtx struct {
	SharedAssetCtx
	CirculatingSupply float64 `json:"circulatingSupply"`
}

// WsActiveAssetData represents active asset data
type WsActiveAssetData struct {
	User             string     `json:"user"`
	Coin             string     `json:"coin"`
	Leverage         Leverage   `json:"leverage"`
	MaxTradeSzs      [2]float64 `json:"maxTradeSzs"`
	AvailableToTrade [2]float64 `json:"availableToTrade"`
}

// WsTwapSliceFill represents TWAP slice fill
type WsTwapSliceFill struct {
	Fill   WsFill `json:"fill"`
	TwapId int64  `json:"twapId"`
}

// WsUserTwapSliceFills represents user TWAP slice fills
type WsUserTwapSliceFills struct {
	IsSnapshot     bool              `json:"isSnapshot,omitempty"`
	User           string            `json:"user"`
	TwapSliceFills []WsTwapSliceFill `json:"twapSliceFills"`
}

// TwapState represents TWAP state
type TwapState struct {
	Coin        string  `json:"coin"`
	User        string  `json:"user"`
	Side        string  `json:"side"`
	Sz          float64 `json:"sz"`
	ExecutedSz  float64 `json:"executedSz"`
	ExecutedNtl float64 `json:"executedNtl"`
	Minutes     int     `json:"minutes"`
	ReduceOnly  bool    `json:"reduceOnly"`
	Randomize   bool    `json:"randomize"`
	Timestamp   int64   `json:"timestamp"`
}

// TwapStatus represents TWAP status
type TwapStatus struct {
	Status      string `json:"status"`
	Description string `json:"description"`
}

// WsTwapHistory represents TWAP history
type WsTwapHistory struct {
	State  TwapState  `json:"state"`
	Status TwapStatus `json:"status"`
	Time   int64      `json:"time"`
}

// WsUserTwapHistory represents user TWAP history
type WsUserTwapHistory struct {
	IsSnapshot bool            `json:"isSnapshot,omitempty"`
	User       string          `json:"user"`
	History    []WsTwapHistory `json:"history"`
}

// WsUserFundings represents user fundings
type WsUserFundings struct {
	IsSnapshot bool            `json:"isSnapshot,omitempty"`
	User       string          `json:"user"`
	Fundings   []WsUserFunding `json:"fundings"`
}

// WsUserNonFundingLedgerUpdates represents user non-funding ledger updates
type WsUserNonFundingLedgerUpdates struct {
	IsSnapshot bool               `json:"isSnapshot,omitempty"`
	User       string             `json:"user"`
	Updates    []NonFundingUpdate `json:"updates"`
}
