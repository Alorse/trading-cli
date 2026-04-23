package models

// ExchangeInfo describes an exchange and its capabilities.
type ExchangeInfo struct {
	Name string `json:"name"`
	Type string `json:"type"` // "crypto" or "stock"
}
