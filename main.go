package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

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

type TransactionDto struct {
	Value       int64  `json:"valor"`
	Type        string `json:"tipo"`
	Description string `json:"descricao"`
	ClientID    int    `json:"id"`
}

type Balance struct {
	Total         int64     `json:"total"`
	StatamentDate time.Time `json:"data_extrato"`
	Limit         int64     `json:"limite"`
}

type TransactionResponse struct {
	Limit   int64 `json:"limite"`
	Balance int64 `json:"saldo"`
}

var db *pgxpool.Pool

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	//dbHost := "localhost"
	//dbUser := "rinha_user"
	//dbPassword := "rinha_pass"
	//dbName := "rinha_db"

	connectionurl := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)

	//conn, err := sql.Open("postgres", connectionurl)
	//if err != nil {
	//	panic(err)
	//}
	//err = conn.Ping()
	//if err != nil {
	//	panic(err)
	//}

	pgxconn, err := pgxpool.ParseConfig(connectionurl)
	if err != nil {
		panic(err)
	}
	conn, err := pgxpool.NewWithConfig(context.Background(), pgxconn)
	if err != nil {
		panic(err)
	}
	err = conn.Ping(context.Background())
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// set db configuration
	pgxConfig := conn.Config()
	pgxConfig.MaxConns = 9

	// set global db
	db = conn

	http.HandleFunc("GET /clientes/{id}/extrato", func(w http.ResponseWriter, r *http.Request) {
		pathId := r.PathValue("id")
		id, _ := strconv.Atoi(pathId)
		if id > 5 || id <= 0 {
			jsonResponse(w, http.StatusNotFound, nil)
			return
		}
		getBankStatement(r.Context(), id, w)
	})
	http.HandleFunc("POST /clientes/{id}/transacoes", func(w http.ResponseWriter, r *http.Request) {
		pathId := r.PathValue("id")
		id, _ := strconv.Atoi(pathId)
		if id > 5 || id <= 0 {
			jsonResponse(w, http.StatusNotFound, nil)
			return
		}
		var req TransactionDto
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			jsonResponse(w, http.StatusUnprocessableEntity, nil)
			return
		}
		descriptionLen := len(req.Description)
		if (req.Type != "c" && req.Type != "d") || (descriptionLen < 1 || descriptionLen > 10) {
			jsonResponse(w, http.StatusUnprocessableEntity, nil)
			return
		}
		req.ClientID = id
		createTransaction(r.Context(), &req, w)
	})

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println(fmt.Sprintf("Server is running on port :%s", port))
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}

func getBankStatement(ctx context.Context, id int, w http.ResponseWriter) {
	balanceQuery := `SELECT balance, user_limit FROM clients WHERE id = $1;`
	transactionsQuery := `SELECT amount, type, description, created_at FROM bank_transactions t WHERE client_id = $1 ORDER BY created_at DESC LIMIT 10;`

	var balance, limit int64
	row := db.QueryRow(ctx, balanceQuery, id)
	err := row.Scan(&balance, &limit)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, nil)
		return
	}

	// get last transactions
	result, err := db.Query(ctx, transactionsQuery, id)
	if err != nil {
		jsonResponse(w, http.StatusBadRequest, nil)
		return
	}
	var transactions []LastTransactions

	for result.Next() {
		t := LastTransactions{}
		_ = result.Scan(&t.Value, &t.Type, &t.Description, &t.CreatedAt)
		transactions = append(transactions, t)
	}
	bankStatement := BankStatement{
		Balance: Balance{
			Total:         balance,
			StatamentDate: time.Now().UTC(),
			Limit:         limit,
		},
		LastTransactions: transactions,
	}
	jsonResponse(w, http.StatusOK, bankStatement)
}

func createTransaction(ctx context.Context, t *TransactionDto, w http.ResponseWriter) {
	_, err := db.Exec(ctx,
		"INSERT INTO bank_transactions(type, description, amount, client_id) VALUES ($1, $2, $3, $4)",
		t.Type, t.Description, t.Value, t.ClientID)
	if err != nil {
		jsonResponse(w, http.StatusUnprocessableEntity, nil)
		return
	}
	var newBalance, limit int64
	err = db.QueryRow(ctx,
		"SELECT balance, user_limit FROM clients WHERE id = $1", t.ClientID).Scan(&newBalance, &limit)
	if err != nil {
		jsonResponse(w, http.StatusUnprocessableEntity, nil)
		return
	}
	transaction := TransactionResponse{
		Balance: newBalance,
		Limit:   limit,
	}
	jsonResponse(w, http.StatusOK, transaction)
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
