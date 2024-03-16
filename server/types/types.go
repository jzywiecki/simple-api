package types

type ResponseToHttp struct {
	Response            Response `json:"response"`
	Average             float64  `json:"average"`
	Median              float64  `json:"median"`
	StandardDeviation   float64  `json:"standard_deviation"`
	Max                 float64  `json:"max"`
	Min                 float64  `json:"min"`
	PriceOfCoinOtherApi string   `json:"price_of_coin_other_api"`
}

type Response struct {
	Data   []CryptoListing `json:"data"`
	Status Status          `json:"status"`
}

type Status struct {
	Timestamp    string `json:"timestamp"`
	ErrorCode    int    `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Elapsed      int    `json:"elapsed"`
	CreditCount  int    `json:"credit_count"`
	Notice       string `json:"notice"`
}

type CryptoListing struct {
	ID                            int              `json:"id"`
	Name                          string           `json:"name"`
	Symbol                        string           `json:"symbol"`
	Slug                          string           `json:"slug"`
	CMCRank                       int              `json:"cmc_rank"`
	NumMarketPairs                int              `json:"num_market_pairs"`
	CirculatingSupply             float64          `json:"circulating_supply"`
	TotalSupply                   float64          `json:"total_supply"`
	MaxSupply                     float64          `json:"max_supply,omitempty"` // Omit empty fields
	InfiniteSupply                *bool            `json:"infinite_supply,omitempty"`
	LastUpdated                   string           `json:"last_updated"`
	DateAdded                     string           `json:"date_added"`
	Tags                          []string         `json:"tags"`
	Platform                      *Platform        `json:"platform,omitempty"` // Use a pointer for optional field
	SelfReportedCirculatingSupply *float64         `json:"self_reported_circulating_supply,omitempty"`
	SelfReportedMarketCap         *float64         `json:"self_reported_market_cap,omitempty"`
	Quote                         map[string]Quote `json:"quote"`
}

type Platform struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Symbol       string `json:"symbol"`
	Slug         string `json:"slug"`
	TokenAddress string `json:"token_address"`
}

type Quote struct {
	Price                 float64 `json:"price"`
	Volume24h             float64 `json:"volume_24h"`
	VolumeChange24h       float64 `json:"volume_change_24h"`
	PercentChange1h       float64 `json:"percent_change_1h"`
	PercentChange24h      float64 `json:"percent_change_24h"`
	PercentChange7d       float64 `json:"percent_change_7d"`
	MarketCap             float64 `json:"market_cap"`
	MarketCapDominance    float64 `json:"market_cap_dominance"`
	FullyDilutedMarketCap float64 `json:"fully_diluted_market_cap"`
	LastUpdated           string  `json:"last_updated"`
}

type Coin struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}
