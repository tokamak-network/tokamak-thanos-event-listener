package redis

type Config struct {
	Addresses  string `json:"addresses"`
	Password   string `json:"password"`
	MasterName string `json:"master_name"`
}
