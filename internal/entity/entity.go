package entity

import "time"

type BankStatement struct {
  Balance          Balance            `json:"saldo"`
  LastTransactions []LastTransactions `json:"ultimas_transacoes"`
}

type LastTransactions struct {
  Type        string    `json:"tipo"`
  Description string    `json:"descricao"`
  Value       int64     `json:"valor"`
  CreatedAt   time.Time `json:"realizada_em"`
}

type Balance struct {
  Total         int64     `json:"total"`
  StatamentDate time.Time `json:"data_extrato"`
  Limit         int64     `json:"limite"`
}
