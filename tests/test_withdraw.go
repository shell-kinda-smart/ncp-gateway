package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// 1. Create Merchant First to get API Key
	createUrl := "http://localhost:8080/merchants"
	createPayload := map[string]string{
		"name": "Withdraw Tester",
	}
	jsonCreate, _ := json.Marshal(createPayload)
	resp, _ := http.Post(createUrl, "application/json", bytes.NewBuffer(jsonCreate))
	body, _ := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	
	var mResponse map[string]string
	json.Unmarshal(body, &mResponse)
	apiKey := mResponse["api_key"]
	fmt.Printf("Created Merchant. API Key: %s\n", apiKey)

	// 2. Call Withdraw
	withdrawUrl := "http://localhost:8080/withdraw"
	withdrawPayload := map[string]string{
		"coin": "BTC",
		"address": "bc1q-cold-storage-destination",
		"amount": "0.5",
	}
	jsonWithdraw, _ := json.Marshal(withdrawPayload)
	
	req, _ := http.NewRequest("POST", withdrawUrl, bytes.NewBuffer(jsonWithdraw))
	req.Header.Set("X-API-Key", apiKey)
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	wResp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Withdraw Request Failed: %v\n", err)
		return
	}
	defer wResp.Body.Close()
	
	wBody, _ := ioutil.ReadAll(wResp.Body)
	fmt.Printf("Withdraw Status: %s\n", wResp.Status)
	fmt.Printf("Withdraw Body: %s\n", string(wBody))
}
