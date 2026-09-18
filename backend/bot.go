//go:build ignore

package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	botToken := "8801688586:AAFJ6T-UPAHLIR2PuDM3JclFpnJnRbzrlN0" 
	
	if envToken := os.Getenv("TELEGRAM_BOT_TOKEN"); envToken != "" {
		botToken = envToken
	}

	if botToken == "" || botToken == "7823561234:AAHxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		log.Fatal("ERROR: Harap masukkan TELEGRAM_BOT_TOKEN yang valid di dalam file bot.go!")
	}

	var err error
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatalf("Failed to initialize Telegram Bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Authorized on Telegram account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		msg := update.Message
		text := msg.Text
		chatID := msg.Chat.ID
		telegramID := msg.From.ID
		username := msg.From.UserName
		name := msg.From.FirstName + " " + msg.From.LastName

		args := strings.Fields(text)
		if len(args) == 0 {
			continue
		}

		command := args[0]
		response := "Perintah tidak dikenali. Ketik /help untuk daftar perintah."

		switch command {
		case "/start":
			response = fmt.Sprintf("Halo %s! Selamat datang di E-Wallet Bot, King.\n\nDaftar Perintah:\n/register - Membuat akun & wallet\n/balance - Cek saldo e-wallet\n/deposit <jumlah> - Isi saldo\n/withdraw <jumlah> - Tarik saldo", name)
		
		case "/help":
			response = "Daftar Perintah E-Wallet Bot:\n- /register : Pendaftaran akun baru\n- /balance : Cek saldo terkini\n- /deposit <nominal> : Top up saldo\n- /withdraw <nominal> : Tarik saldo"

		case "/register":
			var exists bool
			_ = db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE telegram_id = $1)", telegramID).Scan(&exists)
			if exists {
				response = "Akun Anda sudah terdaftar sebelumnya, King!"
				break
			}

			var userID int
			err = db.QueryRow("INSERT INTO users (telegram_id, username, name) VALUES ($1, $2, $3) RETURNING id",
				telegramID, username, name).Scan(&userID)
			if err != nil {
				response = "Gagal mendaftarkan akun."
				break
			}

			_, err = db.Exec("INSERT INTO wallets (user_id, balance) VALUES ($1, 0.00)", userID)
			if err != nil {
				response = "Gagal membuat wallet."
				break
			}
			response = "Registrasi berhasil, King! Wallet Anda telah dibuat dengan saldo awal Rp 0."

		case "/balance":
			var balance float64
			err := db.QueryRow(`
				SELECT w.balance FROM wallets w 
				JOIN users u ON w.user_id = u.id 
				WHERE u.telegram_id = $1`, telegramID).Scan(&balance)
			if err != nil {
				response = "Anda belum terdaftar, King. Silakan ketik /register terlebih dahulu."
				break
			}
			response = fmt.Sprintf("Saldo e-wallet Anda saat ini, King: Rp %.2f", balance)

		case "/deposit":
			if len(args) < 2 {
				response = "Format salah, King. Gunakan: /deposit <jumlah>\nContoh: /deposit 50000"
				break
			}
			amount, err := strconv.ParseFloat(args[1], 64)
			if err != nil || amount <= 0 {
				response = "Jumlah deposit tidak valid, King."
				break
			}

			var walletID int
			var currentBalance float64
			err = db.QueryRow(`
				SELECT w.id, w.balance FROM wallets w 
				JOIN users u ON w.user_id = u.id 
				WHERE u.telegram_id = $1`, telegramID).Scan(&walletID, &currentBalance)
			if err != nil {
				response = "Anda belum terdaftar, King. Silakan ketik /register terlebih dahulu."
				break
			}

			newBalance := currentBalance + amount
			_, err = db.Exec("UPDATE wallets SET balance = $1 WHERE id = $2", newBalance, walletID)
			if err != nil {
				response = "Gagal memproses deposit ke database."
				break
			}

			_, _ = db.Exec("INSERT INTO transactions (wallet_id, type, amount, description) VALUES ($1, 'DEPOSIT', $2, 'Telegram Bot Deposit')",
				walletID, amount)

			response = fmt.Sprintf("✅ Deposit berhasil, King!\nNominal: Rp %.2f\nSaldo Baru: Rp %.2f", amount, newBalance)

		case "/withdraw":
			if len(args) < 2 {
				response = "Format salah, King. Gunakan: /withdraw <jumlah>\nContoh: /withdraw 25000"
				break
			}
			amount, err := strconv.ParseFloat(args[1], 64)
			if err != nil || amount <= 0 {
				response = "Jumlah withdraw tidak valid, King."
				break
			}

			var walletID int
			var currentBalance float64
			err = db.QueryRow(`
				SELECT w.id, w.balance FROM wallets w 
				JOIN users u ON w.user_id = u.id 
				WHERE u.telegram_id = $1`, telegramID).Scan(&walletID, &currentBalance)
			if err != nil {
				response = "Anda belum terdaftar, King. Silakan ketik /register terlebih dahulu."
				break
			}

			if currentBalance < amount {
				response = fmt.Sprintf("❌ Saldo tidak mencukupi, King!\nSaldo saat ini: Rp %.2f", currentBalance)
				break
			}

			newBalance := currentBalance - amount
			_, err = db.Exec("UPDATE wallets SET balance = $1 WHERE id = $2", newBalance, walletID)
			if err != nil {
				response = "Gagal memproses withdraw."
				break
			}

			_, _ = db.Exec("INSERT INTO transactions (wallet_id, type, amount, description) VALUES ($1, 'WITHDRAW', $2, 'Telegram Bot Withdraw')",
				walletID, amount)

			response = fmt.Sprintf("✅ Withdraw berhasil, King!\nNominal: Rp %.2f\nSaldo Baru: Rp %.2f", amount, newBalance)
		}

		msgOut := tgbotapi.NewMessage(chatID, response)
		bot.Send(msgOut)
	}
}
