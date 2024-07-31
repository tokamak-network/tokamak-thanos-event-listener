package types

type Token struct {
	Symbol   string `json:"symbol"`
	Decimals int    `json:"decimals"`
	Address  string `json:"address"`
}
