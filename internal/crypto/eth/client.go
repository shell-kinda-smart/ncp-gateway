package eth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
)

type Client struct {
	URL string
}

func NewClient(url string) *Client {
	return &Client{URL: url}
}

type rpcRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	Id      int           `json:"id"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  interface{}     `json:"error"`
}

func (c *Client) Call(method string, params ...interface{}) (json.RawMessage, error) {
	reqBody := rpcRequest{
		Jsonrpc: "2.0",
		Method:  method,
		Params:  params,
		Id:      1,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	resp, err := http.Post(c.URL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body) // Changed to io.ReadAll and renamed variable to respBody
	if resp.StatusCode != 200 { // Added status code check
		return nil, fmt.Errorf("RPC Error: %d %s", resp.StatusCode, string(respBody))
	}
	var rpcResp rpcResponse
	if err := json.Unmarshal(respBody, &rpcResp); err != nil { // Changed to respBody
		return nil, err
	}

	if rpcResp.Error != nil {
		return nil, fmt.Errorf("RPC Error: %v", rpcResp.Error)
	}

	return rpcResp.Result, nil
}

func (c *Client) GetBlockNumber() (int64, error) {
	res, err := c.Call("eth_blockNumber")
	if err != nil {
		return 0, err
	}
	var hexStr string
	json.Unmarshal(res, &hexStr)
	return strconv.ParseInt(hexStr, 0, 64)
}

// Block includes transactions
type Block struct {
	Number       string        `json:"number"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
}

type Transaction struct {
	Hash        string `json:"hash"`
	From        string `json:"from"`
	To          string `json:"to"`
	Value       string `json:"value"` // Hex
	BlockNumber string `json:"blockNumber"`
}

func (c *Client) GetBlockByNumber(number int64) (*Block, error) {
	hexNum := fmt.Sprintf("0x%x", number)
	res, err := c.Call("eth_getBlockByNumber", hexNum, true) // true = full tx objects
	if err != nil {
		return nil, err
	}
	// If null (block not found/pending), handle it
	if string(res) == "null" {
		return nil, fmt.Errorf("block not found")
	}

	var block Block
	if err := json.Unmarshal(res, &block); err != nil {
		return nil, err
	}
	return &block, nil
}

func HexToBigInt(hexStr string) *big.Int {
	i := new(big.Int)
	i.SetString(hexStr, 0) // auto detect base (0x means hex)
	return i
}
