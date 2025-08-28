// internal/bot/price.go
package bot

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CoinGeckoResponse structure for parsing API response
type CoinGeckoResponse struct {
	Bitcoin struct {
		Usd float64 `json:"usd"`
	} `json:"bitcoin"`
}

// getCryptoPrice gets the current price of a cryptocurrency
func GetCryptoPrice(coinSymbol string) (float64, error) {
	if coinSymbol != "BTC" {
		return 0, fmt.Errorf("пока поддерживается только BTC")
	}
	url := "https://api.coingecko.com/api/v3/simple/price?ids=bitcoin&vs_currencies=usd"
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	var data CoinGeckoResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, err
	}
	return data.Bitcoin.Usd, nil
}
