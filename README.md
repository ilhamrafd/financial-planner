# E-Wallet Simple App

Aplikasi E-Wallet sederhana dengan:
- **Backend**: Golang (`net/mux`, `lib/pq`)
- **Frontend**: Vue 3 + Vite + Axios
- **Database**: PostgreSQL
- **Telegram Bot**: Integrasi CRUD (Cek Saldo, Deposit, Withdraw, Registrasi via Bot)

## Struktur Folder
- `backend/` : Berisi source code Go (`main.go` untuk REST API, `bot.go` untuk Telegram Bot, `schema.sql` untuk DB).
- `frontend/` : Berisi source code Vue 3 (`App.vue`, `package.json`).

## Cara Menjalankan

### 1. Database PostgreSQL
Buat database baru bernama `ewallet_db` lalu jalankan skema dari `backend/schema.sql`:
```sql
CREATE DATABASE ewallet_db;
```
Jalankan file `schema.sql` di database tersebut.

### 2. Backend Golang & Telegram Bot
Masuk ke folder `backend`, set environment variable, lalu jalankan:
```bash
export DATABASE_URL="postgres://username:password@localhost:5432/ewallet_db?sslmode=disable"
export TELEGRAM_BOT_TOKEN="YOUR_BOT_TOKEN_HERE"

# Menjalankan REST API Server
go run main.go

# Menjalankan Telegram Bot (di terminal terpisah)
go run bot.go
```

### 3. Frontend Vue
Masuk ke folder `frontend`, install dependensi lalu jalankan development server:
```bash
npm install
npm run dev
```
