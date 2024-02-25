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
  query := `
    select c.balance, c.u_limit,t.value, t.type, t.description, t.created_at
    from client c
    left join (
      select * from bank_transaction
      where client_id = $1
      order by created_at desc
      limit 10
    ) t on c.id = client_id
    where c.id = $1;
  `

  var balance, limit int64
  var transactions []entity.LastTransactions

  result, err := db.Query(ctx, query, id)
  if err != nil {
    response.JsonResponse(w, http.StatusBadRequest, nil)
    return
  }

  for result.Next() {
    t := entity.LastTransactions{}
    _ = result.Scan(&balance, &limit, &t.Value, &t.Type, &t.Description, &t.CreatedAt)
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

func CreateTransaction(ctx context.Context, t dto.TransactionDto, w http.ResponseWriter, db *pgxpool.Pool) {
  var newBalance, limit int
  err := db.QueryRow(ctx, "select * from new_transaction($1, $2, $3, $4)", t.ClientID, t.Value, t.Description, t.Type).Scan(&newBalance, &limit)
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
