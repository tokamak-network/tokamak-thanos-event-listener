package thanosnotif

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type Data struct {
	cfg *Config
}

type EthCallResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  string `json:"result"`
}

type TokenInfo struct {
	Address  string
	Symbol   string
	Decimals int
}

func (data *Data) tokenInfoMap() (map[string]TokenInfo, error) {
	tokenInfoMap := make(map[string]TokenInfo)

	for _, addr := range data.cfg.TokenAddresses {
		tokenAddress := strings.Trim(addr, "[]")
		symbol, decimals, err := data.getTokenInfo(tokenAddress)
		if err != nil {
			return nil, err
		}

		tokenInfoMap[tokenAddress] = TokenInfo{
			Address:  tokenAddress,
			Symbol:   symbol,
			Decimals: decimals,
		}
	}

	return tokenInfoMap, nil
}

func (data *Data) getTokenInfo(tokenAddress string) (string, int, error) {
	client := &http.Client{}

	// get symbol data
	symbolData := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"%s","data":"0x95d89b41"}, "latest"],"id":1}`, tokenAddress)

	symbolResp, err := data.makeRequest(client, symbolData)
	if err != nil {
		return "", 0, err
	}

	symbol, err := decodeHexString(symbolResp.Result)
	if err != nil {
		return "", 0, err
	}

	// get decimals data
	decimalsData := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_call","params":[{"to":"%s","data":"0x313ce567"}, "latest"],"id":2}`, tokenAddress)

	decimalsResp, err := data.makeRequest(client, decimalsData)
	if err != nil {
		return "", 0, err
	}

	decimals, err := decodeDecimals(decimalsResp.Result)
	if err != nil {
		return "", 0, err
	}

	return symbol, decimals, nil
}

func (data *Data) makeRequest(client *http.Client, tokenData string) (EthCallResponse, error) {
	req, err := http.NewRequest("POST", data.cfg.L1Rpc, strings.NewReader(tokenData))
	if err != nil {
		return EthCallResponse{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return EthCallResponse{}, err
	}
	defer resp.Body.Close()

	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return EthCallResponse{}, err
	}

	var ethCallResp EthCallResponse
	err = json.Unmarshal(bodyText, &ethCallResp)
	if err != nil {
		return EthCallResponse{}, err
	}

	return ethCallResp, nil
}

func decodeHexString(hexStr string) (string, error) {
	if len(hexStr) < 2 {
		return "", fmt.Errorf("hex string is too short")
	}
	resultHex := hexStr[2:]
	resultBytes, err := hex.DecodeString(resultHex)
	if err != nil {
		return "", err
	}

	decodedString := strings.TrimRight(string(resultBytes), "\x00")
	return decodedString, nil
}

func decodeDecimals(hexStr string) (int, error) {
	if len(hexStr) < 2 {
		return 0, fmt.Errorf("hex string is too short")
	}
	resultHex := hexStr[2:]

	if len(resultHex) > 0 {
		decimalsInt, err := strconv.ParseInt(resultHex, 16, 64)
		if err != nil {
			return 0, err
		}
		return int(decimalsInt), nil
	}
	return 0, fmt.Errorf("invalid decimals value")
}
