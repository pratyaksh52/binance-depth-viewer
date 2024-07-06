package binance

type DepthResult struct {
	Asks [][]string `json:"a"`
	Bids [][]string `json:"b"`
}

type DepthResponse struct {
	Stream string      `json:"stream"`
	Data   DepthResult `json:"data"`
}
