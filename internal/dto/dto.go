package dto

type TransactionDto struct {
  Value       int64  `json:"valor"`
  Type        string `json:"tipo"`
  Description string `json:"descricao"`
  ClientID    int    `json:"id"`
}
