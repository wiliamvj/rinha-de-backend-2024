package service

import (
  "context"
  "net/http"
  "time"

  "github.com/jackc/pgx/v5/pgxpool"
  "github.com/wiliamvj/rinha-backend-2024/internal/dto"
  "github.com/wiliamvj/rinha-backend-2024/internal/entity"
  "github.com/wiliamvj/rinha-backend-2024/internal/response"
)

func GetBankStatement(ctx context.Context, id int, w http.ResponseWriter, db *pgxpool.Pool) {
  balanceQuery := `SELECT balance, u_limit FROM client WHERE id = $1;`
  transactionsQuery := `SELECT value, type, description, created_at FROM bank_transaction t WHERE client_id = $1 ORDER BY created_at DESC LIMIT 10;`

  var balance, limit int64
  row := db.QueryRow(ctx, balanceQuery, id)
  err := row.Scan(&balance, &limit)
  if err != nil {
    response.JsonResponse(w, http.StatusBadRequest, nil)
    return
  }

  result, err := db.Query(ctx, transactionsQuery, id)
  if err != nil {
    response.JsonResponse(w, http.StatusBadRequest, nil)
    return
  }
  var transactions []entity.LastTransactions

  for result.Next() {
    t := entity.LastTransactions{}
    _ = result.Scan(&t.Value, &t.Type, &t.Description, &t.CreatedAt)
    transactions = append(transactions, t)
  }
  bankStatement := entity.BankStatement{
    Balance: entity.Balance{
      Total:         balance,
      StatamentDate: time.Now().UTC(),
      Limit:         limit,
    },
    LastTransactions: transactions,
  }
  response.JsonResponse(w, http.StatusOK, bankStatement)
}

func CreateTransaction(ctx context.Context, t *dto.TransactionDto, w http.ResponseWriter, db *pgxpool.Pool) {
  var newBalance, limit int
  err := db.QueryRow(ctx, "SELECT * FROM new_transaction($1, $2, $3, $4)", t.ClientID, t.Value, t.Description, t.Type).Scan(&newBalance, &limit)
  if err != nil {
    response.JsonResponse(w, http.StatusUnprocessableEntity, nil)
    return
  }
  transaction := response.TransactionResponse{
    Balance: int64(newBalance),
    Limit:   int64(limit),
  }
  response.JsonResponse(w, http.StatusOK, transaction)
}
