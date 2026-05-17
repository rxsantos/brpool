package bitcoin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"brpool/internal/pool/types"

	// Sonic is an extremely fast JSON parser
	// created by ByteDance.
	// It replaces Go's default encoding/json.
	"github.com/bytedance/sonic"
	// "github.com/bytedance/sonic/ast"
)

// RPCClient represents a JSON-RPC client
// used for communication with Bitcoin Core.
//
// Example:
// http://username:password@127.0.0.1:8332
type RPCClient struct {
	// Bitcoin Core RPC endpoint URL
	URL      string
	// RPC user defined in bitcoin.conf
	User     string
	// RPC password defined in bitcoin.conf
	Password string
	// You could use a custom client here
	// to set timeout, keepalive, etc.
	//
	// Example:
	// Client *http.Client
}

// Structure sent to Bitcoin Core.
//
// Bitcoin Core uses the JSON-RPC protocol.
//
// Example sent:
//
// {
// "jsonrpc":"1.0",
// "id":"pool",
// "method":"getblocktemplate",
// "params":[]
// }
type rpcRequest struct {
	// JSON-RPC protocol version
	Jsonrpc string        `json:"jsonrpc"`
	// Arbitrary request ID used to correlate response
	ID      string        `json:"id"`
	// RPC method name
	//
	// Examples:
	// getblocktemplate
	// submitblock
	// getblockchaininfo
	Method  string        `json:"method"`
	// Parameters sent to the RPC method
	Params  []interface{} `json:"params"`
}

// Bitcoin Core response structure.
//
// Example:
//
// {
// "result": {...},
// "error": null,
// "id": "pool"
// }
type rpcResponse struct {
	// Result contains the raw JSON returned by Bitcoin Core.
	//
	// We use json.RawMessage to avoid unnecessary parsing initially.
	//
	// This improves performance because:
	//
	// 1. First we parse only the RPC envelope
	// 2. Then we parse only the "result" field
	//
	// It's similar to lazy decoding.
	Result json.RawMessage `json:"result"`
	// RPC error field.
	//
	// If something goes wrong in Bitcoin Core,
	// the error will appear here.
	//
	// Example:
	// {"code":-1,"message":"Block decode failed"}
	Error  interface{}     `json:"error"`
	// Same ID sent in the request
	ID     string          `json:"id"`
}

// NewRPC creates a new instance of the RPC client.
func NewRPC(cfg types.Config) *RPCClient {
	return &RPCClient{
		URL:      fmt.Sprintf("http://%s:%d", cfg.RPCHost, cfg.RPCPort),
		User:     cfg.RPCUser,
		Password: cfg.RPCPass,
		// Client:   &http.Client{},
	}
}

// This call executes any Bitcoin Core RPC method.
//
// This is the central function of RPC communication.
//
// All methods pass through here:
//
// - getblocktemplate
// - submitblock
// - getblockchaininfo
// - getnetworkhashps
// etc.
func (r *RPCClient) call(method string, params []interface{}, result interface{}) error {
	reqBody := rpcRequest{
		Jsonrpc: "1.0",
		ID:      "pool",
		Method:  method,
		Params:  params,
	}

	// Converts Go struct to JSON
	//
	// Sonic does this much faster
	// than encoding/json.
	b, err := sonic.Marshal(reqBody)
	if err != nil {
		return err
	}

	// Creates an HTTP POST request.
	req, err := http.NewRequest("POST", r.URL, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	// Defines the content type
	req.Header.Set("Content-Type", "application/json")
	// Define HTTP Basic authentication
	//
	// Bitcoin Core uses:
	// rpcuser
	// rpcpassword
	req.SetBasicAuth(r.User, r.Password)

	// Executes HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	// Closes connection automatically
	defer resp.Body.Close()

	// Read the entire Bitcoin Core response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Structure that will receive RPC response
	var rpcResp rpcResponse

	// Parse JSON -> struct Go
	if err := sonic.Unmarshal(body, &rpcResp); err != nil {
		return err
	}

	// Checks if Bitcoin Core returned an RPC error.
	if rpcResp.Error != nil {
		return fmt.Errorf("rpc error: %v", rpcResp.Error)
	}

	// Some methods don't need to return a result.
	//
	// Example:
	// submitblock
	if result == nil {
		return nil
	}

	// Parses ONLY the "result" field
	// into the desired struct.
	//
	// Example:
	//
	// result -> types.BlockTemplate
	//
	// This avoids unnecessary parsing.
	return sonic.Unmarshal(rpcResp.Result, result)

}

// GetBlockTemplate requests a new candidate block from Bitcoin Core.
//
// This is the heart of the mining pool.
//
// Bitcoin Core returns:
//
// - transactions
// - previous block hash
// - bits
// - version
// - height
// - coinbasevalue
// etc.
func (r *RPCClient) GetBlockTemplate() (*types.BlockTemplate, error) {
	// Structure that will receive the template
	var tmpl types.BlockTemplate
	// Call RPC: 
	// getblocktemplate
	err := r.call("getblocktemplate", 
		// RPC Parameters
		[]interface{}{
			map[string]interface{}{
			// Announces support for SegWit
			//
			// Practically mandatory today.			
				"rules": []string{"segwit"},
			},
		}, 
		// Where to save the result
		&tmpl)

	// Returns template and possible error.		
	return &tmpl, err
}

// SubmitBlock sends a mined block to Bitcoin Core.
//
// If the block is valid and resolves to the network target:
//
// Bitcoin Core will propagate the block to the entire Bitcoin network.
func (r *RPCClient) SubmitBlock(blockHex string) error {
	// submitblock typically does not return useful data.
	var result interface{}
	// Calls RPC:
	// submitblock "<hex of the block>"
	return r.call("submitblock", []interface{}{blockHex}, &result)
}