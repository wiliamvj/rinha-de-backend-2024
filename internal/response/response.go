package response

import (
  "encoding/json"
  "net/http"
)

type TransactionResponse struct {
  Limit   int64 `json:"limite"`
  Balance int64 `json:"saldo"`
}

func JsonResponse(w http.ResponseWriter, status int, data interface{}) {
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  json.NewEncoder(w).Encode(data)
}
