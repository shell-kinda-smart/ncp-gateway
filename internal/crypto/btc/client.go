package btc

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
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
