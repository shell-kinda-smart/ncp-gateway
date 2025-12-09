package price

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// PriceService handles fetching crypto prices
type PriceService struct {
	Client *http.Client
}

func NewPriceService() *PriceService {
	return &PriceService{
		Client: &http.Client{Timeout: 5 * time.Second},
	}
}

// GetUSDPrice returns the current USD price for a given coin (BTC, LTC, ETH)
// Uses CoinGecko API (Free Tier)
func (s *PriceService) GetUSDPrice(coin string) (float64, error) {
	// Map coin symbol to CoinGecko ID
	coingeckoID := ""
	switch strings.ToUpper(coin) {
	case "BTC":
		coingeckoID = "bitcoin"
	case "LTC":
		coingeckoID = "litecoin"
	case "ETH":
		coingeckoID = "ethereum"
	default:
		return 0, fmt.Errorf("unsupported coin: %s", coin)
	}

	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", coingeckoID)
	
	resp, err := s.Client.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("api returned status: %d", resp.StatusCode)
	}

	// Response format: {"litecoin": {"usd": 95.20}}
	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if priceData, ok := result[coingeckoID]; ok {
		if usd, ok := priceData["usd"]; ok {
			return usd, nil
		}
	}

	return 0, fmt.Errorf("price not found in response")
}
