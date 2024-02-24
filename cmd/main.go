package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
	"github.com/wiliamvj/rinha-backend-2024/internal/dto"
	"github.com/wiliamvj/rinha-backend-2024/internal/response"
	"github.com/wiliamvj/rinha-backend-2024/internal/service"
)

var db *pgxpool.Pool

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// dbHost := "localhost"
	// dbUser := "rinha_user"
	// dbPassword := "rinha_pass"
	// dbName := "rinha_db"

	connectionurl := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbUser, dbPassword, dbName)

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
	pgxConfig.MaxConns = 10

	// set global db
	db = conn

	http.HandleFunc("GET /clientes/{id}/extrato", func(w http.ResponseWriter, r *http.Request) {
		pathId := r.PathValue("id")
		id, _ := strconv.Atoi(pathId)
		if id > 5 || id <= 0 {
			response.JsonResponse(w, http.StatusNotFound, nil)
			return
		}
		service.GetBankStatement(r.Context(), id, w, db)
	})
	http.HandleFunc("POST /clientes/{id}/transacoes", func(w http.ResponseWriter, r *http.Request) {
		pathId := r.PathValue("id")
		id, _ := strconv.Atoi(pathId)
		if id > 5 || id <= 0 {
			response.JsonResponse(w, http.StatusNotFound, nil)
			return
		}
		var req dto.TransactionDto
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			response.JsonResponse(w, http.StatusUnprocessableEntity, nil)
			return
		}
		descriptionLen := len(req.Description)
		if (req.Type != "c" && req.Type != "d") || (descriptionLen < 1 || descriptionLen > 10) {
			response.JsonResponse(w, http.StatusUnprocessableEntity, nil)
			return
		}
		req.ClientID = id
		service.CreateTransaction(r.Context(), req, w, db)
	})

	port := os.Getenv("HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("Server is running on port :%s", port)
	err = http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
	if err != nil {
		panic(err)
	}
}
