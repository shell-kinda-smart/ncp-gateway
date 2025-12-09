package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Note: We are hitting the standard /merchants endpoint now
	url := "http://localhost:8080/merchants"
	
	payload := map[string]string{
		"name":       "Pure Managed Store",
		"payout_btc": "bc1q-cold-storage-example",
		"payout_ltc": "ltc1-cold-storage-example",
		"payout_eth": "0xColdStorageETHExample",
	}
	
	jsonPayload, _ := json.Marshal(payload)
	
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		fmt.Printf("Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()
	
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("Response Status: %s\n", resp.Status)
	fmt.Printf("Response Body: %s\n", string(body))
}
