package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type Config struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"` // Optional default
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	config := loadConfig()

	for {
		clearScreen()
		fmt.Println("===========================================")
		fmt.Println("   🚀 Crypto Gateway API Tester (CLI)")
		fmt.Println("===========================================")
		fmt.Printf("Target: %s\n", config.BaseURL)
		fmt.Println("-------------------------------------------")
		fmt.Println("1. 🛠️  Run DB Setup (Reset & Seed)")
		fmt.Println("2. 🏪 Create Merchant (API)")
		fmt.Println("3. 👛 Generate Wallet (API)")
		fmt.Println("4. ⚙️  Update Config URL")
		fmt.Println("5. 🚪 Exit")
		fmt.Println("-------------------------------------------")
		fmt.Print("Select option: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			runDBSetup()
		case "2":
			createMerchantAPI(reader, config)
		case "3":
			createWalletAPI(reader, config)
		case "4":
			updateConfig(reader, &config)
		case "5":
			fmt.Println("Bye!")
			os.Exit(0)
		default:
			fmt.Println("Invalid option!")
		}
		
		fmt.Println("\nPress Enter to continue...")
		reader.ReadString('\n')
	}
}

func loadConfig() Config {
	// Default
	cfg := Config{BaseURL: "http://localhost:8080"}
	
	data, err := ioutil.ReadFile("test_config.json")
	if err == nil {
		json.Unmarshal(data, &cfg)
	}
	return cfg
}

func saveConfig(cfg Config) {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	_ = ioutil.WriteFile("test_config.json", data, 0644)
}

func updateConfig(r *bufio.Reader, cfg *Config) {
	fmt.Printf("Enter new Base URL [%s]: ", cfg.BaseURL)
	input, _ := r.ReadString('\n')
	input = strings.TrimSpace(input)
	if input != "" {
		cfg.BaseURL = input
		saveConfig(*cfg)
	}
}

func runDBSetup() {
	fmt.Println("\n📦 Running DB Setup script...")
	
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Use Default Test Data? (y/n) [y]: ")
	ans, _ := reader.ReadString('\n')
	ans = strings.TrimSpace(strings.ToLower(ans))

	args := []string{"run", "../setup/setup_merchant.go"}

	if ans == "n" {
		fmt.Println("Enter Setup Details:")
		
		fmt.Print("BTC Address (Hit Enter to Skip): ")
		btc, _ := reader.ReadString('\n')
		btc = strings.TrimSpace(btc)
		if btc != "" { args = append(args, "-btc", btc) }

		fmt.Print("BTC XPUB (Leave empty for Test Vector): ")
		bxpub, _ := reader.ReadString('\n')
		bxpub = strings.TrimSpace(bxpub)
		if bxpub != "" { args = append(args, "-btc-xpub", bxpub) }

		fmt.Print("LTC Address (Hit Enter to Skip): ")
		ltc, _ := reader.ReadString('\n')
		ltc = strings.TrimSpace(ltc)
		if ltc != "" { args = append(args, "-ltc", ltc) }

		fmt.Print("LTC XPUB (Leave empty for Test Vector): ")
		lxpub, _ := reader.ReadString('\n')
		lxpub = strings.TrimSpace(lxpub)
		if lxpub != "" { args = append(args, "-ltc-xpub", lxpub) }

		fmt.Print("ETH Address (Hit Enter to Skip): ")
		eth, _ := reader.ReadString('\n')
		eth = strings.TrimSpace(eth)
		if eth != "" { args = append(args, "-eth", eth) }

		fmt.Print("ETH XPUB (Leave empty for Test Vector): ")
		expub, _ := reader.ReadString('\n')
		expub = strings.TrimSpace(expub)
		if expub != "" { args = append(args, "-eth-xpub", expub) }

		fmt.Print("Webhook URL: ")
		wh, _ := reader.ReadString('\n')
		wh = strings.TrimSpace(wh)
		if wh != "" { args = append(args, "-webhook", wh) }
	}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error running setup: %v\n", err)
	}
}

func createMerchantAPI(r *bufio.Reader, cfg Config) {
	fmt.Println("\n--- Create Merchant ---")
	fmt.Print("Name [Test Store]: ")
	name, _ := r.ReadString('\n')
	name = strings.TrimSpace(name)
	if name == "" { name = "Test Store" }

	// Defaults
	defBTC := "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"
	defLTC := "xpub6DDn5pSVZKEuWmeJ8p84mpgVubu5bcu76EPp5Ue4irjBf7TDMr8zTkxtaFKL87f4sjK1MPArHgvRVeiwpXnX8GMtjthNRtMsaRXchTGZgc7"
	defETH := "xpub661MyMwAqRbcFtXgS5sYJABqqG9YLmC4Q1Rdap9gSE8NqtwybGhePY2gZ29ESFjqJoCu1Rupje8YtGqsefD265TMg7usUDFdp6W1EGMcet8"

	fmt.Printf("BTC XPUB (Enter for test): ")
	btc, _ := r.ReadString('\n')
	btc = strings.TrimSpace(btc)
	if btc == "" { btc = defBTC }

	fmt.Printf("LTC XPUB (Enter for YOUR DEFAULT): ")
	ltc, _ := r.ReadString('\n')
	ltc = strings.TrimSpace(ltc)
	if ltc == "" { ltc = defLTC }

	fmt.Printf("ETH XPUB (Enter for test): ")
	eth, _ := r.ReadString('\n')
	eth = strings.TrimSpace(eth)
	if eth == "" { eth = defETH }
	
	payload := map[string]string{
		"name": name,
		"btc_xpub": btc,
		"ltc_xpub": ltc,
		"eth_xpub": eth,
	}
	
	sendPost(cfg.BaseURL+"/merchants", payload, "")
}

func createWalletAPI(r *bufio.Reader, cfg Config) {
	fmt.Println("\n--- Generate Wallet ---")
	fmt.Print("API Key (leave blank to use hardcoded test key): ")
	key, _ := r.ReadString('\n')
	key = strings.TrimSpace(key)
	if key == "" { key = "test_api_key_123456789" }

	fmt.Print("User ID [user_01]: ")
	uid, _ := r.ReadString('\n')
	uid = strings.TrimSpace(uid)
	if uid == "" { uid = "user_01" }

	payload := map[string]string{"user_id": uid}
	sendPost(cfg.BaseURL+"/wallets", payload, key)
}

func sendPost(url string, data interface{}, apiKey string) {
	body, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Request Failed: %v\n", err)
		return
	}
	defer resp.Body.Close()

	respBody, _ := ioutil.ReadAll(resp.Body)
	fmt.Printf("\nStatus: %s\nResponse: %s\n", resp.Status, string(respBody))
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}
