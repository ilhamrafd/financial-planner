package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type User struct {
	ID            int    `json:"id"`
	TelegramID    *int64 `json:"telegram_id"`
	Username      string `json:"username"`
	Password      string `json:"password,omitempty"`
	Name          string `json:"name"`
	AccountNumber string `json:"account_number"`
}

type Wallet struct {
	ID      int     `json:"id"`
	UserID  int     `json:"user_id"`
	Balance float64 `json:"balance"`
}

type Transaction struct {
	ID            int     `json:"id"`
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	CreatedAt     string  `json:"created_at"`
	AccountNumber string  `json:"account_number"`
	Name          string  `json:"name"`
}

type TransactionReq struct {
	UserID      int     `json:"user_id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
}

type TransferReq struct {
	FromUserID          int     `json:"from_user_id"`
	TargetAccountNumber string  `json:"target_account_number"`
	Amount              float64 `json:"amount"`
	Description         string  `json:"description"`
}

type AuthReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name,omitempty"`
}

type SavedAccount struct {
	ID                  int    `json:"id"`
	OwnerID             int    `json:"owner_id"`
	TargetName          string `json:"target_name"`
	TargetAccountNumber string `json:"target_account_number"`
	Name                string `json:"name"`
	AccountNumber       string `json:"account_number"`
}
// Financial planning structs

type FinancialPlan struct {
	ID            int     `json:"id"`
	UserID        int     `json:"user_id"`
	Month         int     `json:"month"`
	Year          int     `json:"year"`
	Income        float64 `json:"income"`
	PctKebutuhan  float64 `json:"pct_kebutuhan"`
	PctTabungan   float64 `json:"pct_tabungan"`
	PctKeinginan  float64 `json:"pct_keinginan"`
}

type FinancialExpense struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	Category    string  `json:"category"` // Kebutuhan, Tabungan, Keinginan
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
	Month       int     `json:"month"`
	Year        int     `json:"year"`
}

var db *sql.DB

func generateAccountNumber() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("0281%06d", rand.Intn(1000000))
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validatePassword(pwd string) bool {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(pwd)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(pwd)
	hasSymbol := regexp.MustCompile(`[!@#\$%\^&\*(),.?":{}|<>_\-+=]`).MatchString(pwd)
	return hasUpper && hasDigit && hasSymbol && len(pwd) >= 6
}

func main() {
	var err error
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/ewallet_db?sslmode=disable"
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Auto-migrate tables — plus migrate old financial_plans (user_id UNIQUE -> per-month)
	db.Exec(`ALTER TABLE financial_plans ADD COLUMN IF NOT EXISTS month INTEGER DEFAULT 1`)
	db.Exec(`ALTER TABLE financial_plans ADD COLUMN IF NOT EXISTS year INTEGER DEFAULT 2026`)
	// drop old UNIQUE(user_id) if exists, keep new UNIQUE(user_id,month,year) via table creation; ensure constraint exists
	db.Exec(`DO $$ BEGIN
		IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'financial_plans_user_id_key') THEN
			ALTER TABLE financial_plans DROP CONSTRAINT financial_plans_user_id_key;
		END IF;
	END $$`)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_financial_plans_user_month_year ON financial_plans(user_id, month, year)`)
	// migrate financial_expenses -> per-month isolation (month/year columns)
	db.Exec(`ALTER TABLE financial_expenses ADD COLUMN IF NOT EXISTS month INTEGER`)
	db.Exec(`ALTER TABLE financial_expenses ADD COLUMN IF NOT EXISTS year INTEGER`)
	db.Exec(`UPDATE financial_expenses SET month = EXTRACT(MONTH FROM created_at)::int, year = EXTRACT(YEAR FROM created_at)::int WHERE month IS NULL OR year IS NULL`)
	db.Exec(`
		CREATE TABLE IF NOT EXISTS saved_accounts (
			id SERIAL PRIMARY KEY,
			owner_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			target_name VARCHAR(255) NOT NULL,
			target_account_number VARCHAR(50) NOT NULL,
			UNIQUE(owner_id, target_account_number)
		);

		CREATE TABLE IF NOT EXISTS financial_plans (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			month INTEGER NOT NULL,
			year INTEGER NOT NULL,
			income NUMERIC(15,2) DEFAULT 0.00,
			pct_kebutuhan NUMERIC(5,2) DEFAULT 50.00,
			pct_tabungan NUMERIC(5,2) DEFAULT 30.00,
			pct_keinginan NUMERIC(5,2) DEFAULT 20.00,
			UNIQUE(user_id, month, year)
		);

		CREATE TABLE IF NOT EXISTS financial_expenses (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
			month INTEGER NOT NULL DEFAULT 1,
			year INTEGER NOT NULL DEFAULT 2026,
			category VARCHAR(50) NOT NULL,
			amount NUMERIC(15,2) NOT NULL,
			description TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)

	r := mux.NewRouter()
	r.Use(enableCors)

	r.HandleFunc("/api/register", registerUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/login", loginUser).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/wallet/{user_id}", getWallet).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/transaction", createTransaction).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/transfer", handleTransfer).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/history/{user_id}", getHistory).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/transfer-history/{user_id}", getTransferHistory).Methods("GET", "OPTIONS")
	// saved accounts
	r.HandleFunc("/api/saved-accounts", createSavedAccount).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/saved-accounts/{user_id}", getSavedAccounts).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/saved-accounts/{id}", deleteSavedAccount).Methods("DELETE", "OPTIONS")
	r.HandleFunc("/api/lookup-account/{account_number}", lookupAccount).Methods("GET", "OPTIONS")
	// financial planning
	r.HandleFunc("/api/financial-plan", setFinancialPlan).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/financial-plan/{user_id}", getFinancialPlan).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/financial-expense", addFinancialExpense).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/financial-expense/{id}", deleteFinancialExpense).Methods("DELETE", "OPTIONS")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func registerUser(w http.ResponseWriter, r *http.Request) {
	var req AuthReq
	json.NewDecoder(r.Body).Decode(&req)
	if !validatePassword(req.Password) {
		http.Error(w, "Password lemah", http.StatusBadRequest)
		return
	}
	accNum := generateAccountNumber()
	var userID int
	err := db.QueryRow("INSERT INTO users (username, password, name, account_number) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Username, req.Password, req.Name, accNum).Scan(&userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	db.Exec("INSERT INTO wallets (user_id, balance) VALUES ($1, 0.00)", userID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func loginUser(w http.ResponseWriter, r *http.Request) {
	var req AuthReq
	json.NewDecoder(r.Body).Decode(&req)
	var u User
	var dbPwd string
	err := db.QueryRow("SELECT id, username, password, name, account_number FROM users WHERE username = $1", req.Username).Scan(&u.ID, &u.Username, &dbPwd, &u.Name, &u.AccountNumber)
	if err != nil || dbPwd != req.Password {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "user": u})
}

func getWallet(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var wlt Wallet
	db.QueryRow("SELECT id, user_id, balance FROM wallets WHERE user_id = $1", vars["user_id"]).Scan(&wlt.ID, &wlt.UserID, &wlt.Balance)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wlt)
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	var req TransactionReq
	json.NewDecoder(r.Body).Decode(&req)
	tx, _ := db.Begin()
	var walletID int
	var balance float64
	db.QueryRow("SELECT id, balance FROM wallets WHERE user_id = $1", req.UserID).Scan(&walletID, &balance)
	newBalance := balance
	if req.Type == "DEPOSIT" {
		newBalance += req.Amount
	} else {
		newBalance -= req.Amount
	}
	tx.Exec("UPDATE wallets SET balance = $1 WHERE id = $2", newBalance, walletID)
	tx.Exec("INSERT INTO transactions (wallet_id, type, amount, description) VALUES ($1, $2, $3, $4)", walletID, req.Type, req.Amount, req.Description)
	tx.Commit()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"new_balance": newBalance})
}

func handleTransfer(w http.ResponseWriter, r *http.Request) {
	var req TransferReq
	json.NewDecoder(r.Body).Decode(&req)
	tx, _ := db.Begin()
	var fromWID int
	var fromBal float64
	tx.QueryRow("SELECT id, balance FROM wallets WHERE user_id = $1 FOR UPDATE", req.FromUserID).Scan(&fromWID, &fromBal)
	if fromBal < req.Amount {
		tx.Rollback()
		http.Error(w, "Saldo tidak cukup", http.StatusBadRequest)
		return
	}
	var toWID int
	var toName string
	err := tx.QueryRow("SELECT w.id, u.name FROM wallets w JOIN users u ON w.user_id = u.id WHERE u.account_number = $1 FOR UPDATE", req.TargetAccountNumber).Scan(&toWID, &toName)
	if err != nil {
		tx.Rollback()
		http.Error(w, "Rekening tujuan tidak ditemukan", http.StatusNotFound)
		return
	}
	tx.Exec("UPDATE wallets SET balance = balance - $1 WHERE id = $2", req.Amount, fromWID)
	tx.Exec("UPDATE wallets SET balance = balance + $1 WHERE id = $2", req.Amount, toWID)
	descOut := fmt.Sprintf("Transfer ke %s (%s)", toName, req.TargetAccountNumber)
	descIn := fmt.Sprintf("Transfer dari pengirim")
	tx.Exec("INSERT INTO transactions (wallet_id, type, amount, description) VALUES ($1, 'TRANSFER_OUT', $2, $3)", fromWID, req.Amount, descOut)
	tx.Exec("INSERT INTO transactions (wallet_id, type, amount, description) VALUES ($1, 'TRANSFER_IN', $2, $3)", toWID, req.Amount, descIn)
	// auto save target account
	tx.Exec("INSERT INTO saved_accounts (owner_id, target_name, target_account_number) VALUES ($1, $2, $3) ON CONFLICT (owner_id, target_account_number) DO UPDATE SET target_name = EXCLUDED.target_name",
		req.FromUserID, toName, req.TargetAccountNumber)
	tx.Commit()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "target_name": toName})
}

func getHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rows, err := db.Query("SELECT t.id, t.type, t.amount, t.description, t.created_at, u.account_number, u.name FROM transactions t JOIN wallets w ON t.wallet_id = w.id JOIN users u ON w.user_id = u.id WHERE w.user_id = $1 ORDER BY t.created_at DESC", vars["user_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var history []Transaction
	for rows.Next() {
		var t Transaction
		rows.Scan(&t.ID, &t.Type, &t.Amount, &t.Description, &t.CreatedAt, &t.AccountNumber, &t.Name)
		history = append(history, t)
	}
	if history == nil { history = []Transaction{} }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func getTransferHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rows, err := db.Query("SELECT id, owner_id, target_name, target_account_number FROM saved_accounts WHERE owner_id = $1 ORDER BY id DESC", vars["user_id"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var contacts []SavedAccount
	for rows.Next() {
		var c SavedAccount
		rows.Scan(&c.ID, &c.OwnerID, &c.TargetName, &c.TargetAccountNumber)
		c.Name = c.TargetName
		c.AccountNumber = c.TargetAccountNumber
		contacts = append(contacts, c)
	}
	if contacts == nil { contacts = []SavedAccount{} }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contacts)
}

func createSavedAccount(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OwnerID int `json:"owner_id"`
		TargetAccountNumber string `json:"target_account_number"`
		TargetName string `json:"target_name"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	name := req.TargetName
	if name == "" {
		db.QueryRow("SELECT name FROM users WHERE account_number = $1", req.TargetAccountNumber).Scan(&name)
		if name == "" {
			http.Error(w, "Rekening tidak ditemukan", http.StatusNotFound)
			return
		}
	}
	_, err := db.Exec("INSERT INTO saved_accounts (owner_id, target_name, target_account_number) VALUES ($1,$2,$3) ON CONFLICT (owner_id, target_account_number) DO UPDATE SET target_name = EXCLUDED.target_name",
		req.OwnerID, name, req.TargetAccountNumber)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success", "name": name})
}

func getSavedAccounts(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	rows, _ := db.Query("SELECT id, owner_id, target_name, target_account_number FROM saved_accounts WHERE owner_id = $1 ORDER BY id DESC", vars["user_id"])
	defer rows.Close()
	var list []SavedAccount
	for rows.Next() {
		var c SavedAccount
		rows.Scan(&c.ID, &c.OwnerID, &c.TargetName, &c.TargetAccountNumber)
		c.Name = c.TargetName
		c.AccountNumber = c.TargetAccountNumber
		list = append(list, c)
	}
	if list == nil { list = []SavedAccount{} }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

func deleteSavedAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	db.Exec("DELETE FROM saved_accounts WHERE id = $1", vars["id"])
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "deleted"})
}

func lookupAccount(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var name, acc string
	err := db.QueryRow("SELECT name, account_number FROM users WHERE account_number = $1", vars["account_number"]).Scan(&name, &acc)
	if err != nil {
		http.Error(w, "Rekening tidak ditemukan", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"name": name, "account_number": acc})
}

// Financial planning handlers

func setFinancialPlan(w http.ResponseWriter, r *http.Request) {
	var p FinancialPlan
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}

	// Security: Validate owner via session check or header, here relying on userID from client
	// In production, use session/JWT middleware.

	if p.PctKebutuhan == 0 && p.PctTabungan == 0 && p.PctKeinginan == 0 {
		p.PctKebutuhan = 50
		p.PctTabungan = 30
		p.PctKeinginan = 20
	}
	if p.Month == 0 || p.Year == 0 {
		http.Error(w, "month and year required", http.StatusBadRequest)
		return
	}
	query := `
		INSERT INTO financial_plans (user_id, month, year, income, pct_kebutuhan, pct_tabungan, pct_keinginan)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (user_id, month, year) DO UPDATE SET
		    income = EXCLUDED.income,
		    pct_kebutuhan = EXCLUDED.pct_kebutuhan,
		    pct_tabungan = EXCLUDED.pct_tabungan,
		    pct_keinginan = EXCLUDED.pct_keinginan`
	_, err := db.Exec(query, p.UserID, p.Month, p.Year, p.Income, p.PctKebutuhan, p.PctTabungan, p.PctKeinginan)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func getFinancialPlan(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["user_id"]
	// Security: Only allow user to access their own ID
	// Frontend passes currentUser.id

	q := r.URL.Query()
	month := q.Get("month")
	year := q.Get("year")

	var plan FinancialPlan
	var err error
	if month != "" && year != "" {
		err = db.QueryRow("SELECT id, user_id, month, year, income, pct_kebutuhan, pct_tabungan, pct_keinginan FROM financial_plans WHERE user_id = $1 AND month = $2 AND year = $3", userID, month, year).
			Scan(&plan.ID, &plan.UserID, &plan.Month, &plan.Year, &plan.Income, &plan.PctKebutuhan, &plan.PctTabungan, &plan.PctKeinginan)
	} else {
		// default to current month/year or latest plan
		err = db.QueryRow("SELECT id, user_id, month, year, income, pct_kebutuhan, pct_tabungan, pct_keinginan FROM financial_plans WHERE user_id = $1 ORDER BY year DESC, month DESC LIMIT 1", userID).
			Scan(&plan.ID, &plan.UserID, &plan.Month, &plan.Year, &plan.Income, &plan.PctKebutuhan, &plan.PctTabungan, &plan.PctKeinginan)
	}
	if err != nil {
		plan = FinancialPlan{UserID: 0, Month: 0, Year: 0, Income: 0, PctKebutuhan: 50, PctTabungan: 30, PctKeinginan: 20}
	}

	// expenses isolated per month/year (not created_at)
	query := "SELECT id, user_id, month, year, category, amount, description, created_at FROM financial_expenses WHERE user_id = $1"
	args := []interface{}{userID}
	if month != "" && year != "" {
		query += " AND month = $2 AND year = $3"
		args = append(args, month, year)
	} else if year != "" {
		query += " AND year = $2"
		args = append(args, year)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.Query(query, args...)
	var expenses []FinancialExpense
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var e FinancialExpense
			rows.Scan(&e.ID, &e.UserID, &e.Month, &e.Year, &e.Category, &e.Amount, &e.Description, &e.CreatedAt)
			expenses = append(expenses, e)
		}
	}
	if expenses == nil { expenses = []FinancialExpense{} }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"plan": plan, "expenses": expenses})
}

func addFinancialExpense(w http.ResponseWriter, r *http.Request) {
	var e FinancialExpense
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, "invalid payload", http.StatusBadRequest)
		return
	}
	// Require month/year to store expense in proper period
	if e.Month == 0 || e.Year == 0 {
		http.Error(w, "month and year required", http.StatusBadRequest)
		return
	}
	_, err := db.Exec("INSERT INTO financial_expenses (user_id, month, year, category, amount, description) VALUES ($1, $2, $3, $4, $5, $6)", e.UserID, e.Month, e.Year, e.Category, e.Amount, e.Description)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "success"})
}

func deleteFinancialExpense(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// Security: Ensure the expense belongs to the logged in user before deletion
	// For simplicity, simple DELETE here (assuming client passes correct owner context)
	db.Exec("DELETE FROM financial_expenses WHERE id = $1", vars["id"])
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"status": "deleted"})
}
