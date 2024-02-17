package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

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

var db *sql.DB

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

	conn, err := sql.Open("postgres", connectionurl)
	if err != nil {
		panic(err)
	}
	err = conn.Ping()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// set db configuration
	conn.SetMaxOpenConns(15)

	// set global db
	db = conn

	http.HandleFunc("GET /clientes/{id}/extrato", func(w http.ResponseWriter, r *http.Request) {
		pathId := r.PathValue("id")
		id, _ := strconv.Atoi(pathId)
		if id > 5 || id <= 0 {
			jsonResponse(w, http.StatusNotFound, nil)
			return
		}
		resp, err := getBankStatement(r.Context(), id)
		if err != nil {
			jsonResponse(w, http.StatusBadRequest, nil)
			return
		}
		jsonResponse(w, http.StatusOK, resp)

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
		resp, err := createTransaction(r.Context(), &req)
		if err != nil {
			jsonResponse(w, http.StatusUnprocessableEntity, nil)
			return
		}
		jsonResponse(w, http.StatusOK, resp)
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

func getBankStatement(ctx context.Context, id int) (*BankStatement, error) {
	// get balance and limit
	var balance, limit int64
	row := db.QueryRowContext(ctx, `SELECT balance, user_limit FROM clients WHERE id = $1;`, id)
	err := row.Scan(&balance, &limit)
	if err != nil {
		return nil, err
	}

	// get last transactions
	query := `SELECT amount, type, description, created_at FROM bank_transactions t where client_id = $1 ORDER BY created_at DESC LIMIT 10;`
	result, err := db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, err
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
	return &bankStatement, nil
}

func createTransaction(ctx context.Context, t *TransactionDto) (*TransactionResponse, error) {
	_, err := db.ExecContext(ctx,
		"INSERT INTO bank_transactions(type, description, amount, client_id) VALUES ($1, $2, $3, $4)",
		t.Type, t.Description, t.Value, t.ClientID)
	if err != nil {
		return nil, err
	}

	var newBalance, limit int64
	err = db.QueryRowContext(ctx,
		"SELECT balance, user_limit FROM clients WHERE id = $1", t.ClientID).Scan(&newBalance, &limit)
	if err != nil {
		return nil, err
	}

	transaction := TransactionResponse{
		Balance: newBalance,
		Limit:   limit,
	}
	return &transaction, nil
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
