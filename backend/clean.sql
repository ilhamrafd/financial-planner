-- Reset data tabel users, wallets, dan transactions
TRUNCATE TABLE transactions, wallets, users RESTART IDENTITY CASCADE;
