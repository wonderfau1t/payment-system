package storage

type Transaction struct {
	Id              int     `json:"id"`
	Timestamp       int64   `json:"timestamp"`
	TransactionHash string  `json:"transactionHash"`
	Sender          string  `json:"sender"`
	Recipient       string  `json:"recipient"`
	Amount          float64 `json:"amount"`
}
