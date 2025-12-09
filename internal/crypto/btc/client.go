package btc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	URL      string
	User     string
	Password string
}

func NewClient(url, user, password string) *Client {
	return &Client{
		URL:      url,
		User:     user,
		Password: password,
	}
}

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Id      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type RPCResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *RPCError       `json:"error"`
	Id     int             `json:"id"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Call executes a raw JSON-RPC request
func (c *Client) Call(method string, params ...interface{}) (json.RawMessage, error) {
	reqBody := RPCRequest{
		Jsonrpc: "1.0",
		Id:      1,
		Method:  method,
		Params:  params,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.User, c.Password)
	req.Header.Set("Content-Type", "text/plain;")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	bodyBytes, _ := io.ReadAll(resp.Body)
	// Restore body for decoding
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, fmt.Errorf("decode error: %v | Body: %s", err, string(bodyBytes))
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc error: %d %s", rpcResp.Error.Code, rpcResp.Error.Message)
	}

	return rpcResp.Result, nil
}

// GetBlockCount returns the current block height
func (c *Client) GetBlockCount() (int64, error) {
	res, err := c.Call("getblockcount")
	if err != nil {
		return 0, err
	}
	var count int64
	err = json.Unmarshal(res, &count)
	return count, err
}

// GetBlockHash returns the hash of the block at the given height
func (c *Client) GetBlockHash(height int64) (string, error) {
	res, err := c.Call("getblockhash", height)
	if err != nil {
		return "", err
	}
	var hash string
	err = json.Unmarshal(res, &hash)
	return hash, err
}

// GetBlockRaw returns the raw hex-encoded block
func (c *Client) GetBlockRaw(hash string) (string, error) {
	// verbosity=0 returns hex-encoded data
	res, err := c.Call("getblock", hash, 0)
	if err != nil {
		return "", err
	}
	var hexData string
	err = json.Unmarshal(res, &hexData)
	return hexData, err
}

// SendRawTransaction broadcasts a hex-encoded transaction
func (c *Client) SendRawTransaction(hexTx string) (string, error) {
	res, err := c.Call("sendrawtransaction", hexTx)
	if err != nil {
		return "", err
	}
	var txID string
	err = json.Unmarshal(res, &txID)
	return txID, err
}

type UTXO struct {
	TxID          string  `json:"txid"`
	Vout          int     `json:"vout"`
	Address       string  `json:"address"`
	ScriptPubKey  string  `json:"scriptPubKey"`
	Amount        float64 `json:"amount"`
	Confirmations int     `json:"confirmations"`
	Safe          bool    `json:"safe"`
	Spendable     bool    `json:"spendable"`
}

// ListUnspent returns UTXOs. 
// addresses can be nil to list all (if wallet is loaded).
func (c *Client) ListUnspent(minConf, maxConf int, addresses []string) ([]UTXO, error) {
	var params []interface{}
	params = append(params, minConf)
	params = append(params, maxConf)
	if addresses != nil {
		params = append(params, addresses)
	}
	
	res, err := c.Call("listunspent", params...)
	if err != nil {
		return nil, err
	}
	
	var utxos []UTXO
	err = json.Unmarshal(res, &utxos)
	return utxos, err
}

type ImportDescriptorRequest struct {
	Desc      string `json:"desc"`
	Timestamp interface{} `json:"timestamp"` // "now" or number
	Active    bool   `json:"active"`
	Internal  bool   `json:"internal"`
	Range     []int  `json:"range,omitempty"` // [0, 1000]
	Label     string `json:"label,omitempty"`
}

// ImportDescriptors imports watch-only descriptors
func (c *Client) ImportDescriptors(requests []ImportDescriptorRequest) error {
	// Params is a list of requests
	var params []interface{}
	params = append(params, requests)
	
	// We ignore the result for now, assumes success if no error. 
	// Real implementation should check the detailed result array.
	_, err := c.Call("importdescriptors", params...)
	return err
}

// EstimateSmartFee returns fee rate in BTC/kB
func (c *Client) EstimateSmartFee(confTarget int) (float64, error) {
	res, err := c.Call("estimatesmartfee", confTarget)
	if err != nil {
		return 0, err
	}
	
	var result struct {
		Feerate float64 `json:"feerate"`
		Errors  []string `json:"errors"`
	}
	if err := json.Unmarshal(res, &result); err != nil {
		return 0, err
	}
	
	// Default fallback if node has no data (regtest/testnet often 0)
	if result.Feerate == 0 {
		return 0.00001, nil // 1 sat/vbyte roughly
	}
	
	return result.Feerate, nil
}
