package domain

// Block represents a single block in the blockchain
type Block struct {
	Number       string        `json:"number"`
	Transactions []Transaction `json:"transactions"`
}
