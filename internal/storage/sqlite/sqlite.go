package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"infotecs-tz/internal/storage"
	"infotecs-tz/internal/utils"
	_ "modernc.org/sqlite"
	"time"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const fn = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	schema := `
	CREATE TABLE IF NOT EXISTS addresses(
	   id INTEGER PRIMARY KEY,
	   address TEXT UNIQUE NOT NULL CHECK(length(address) = 42 AND address LIKE '0x%'),
	   balance REAL NOT NULL CHECK(balance >= 0.0)
	   );
	
	CREATE INDEX IF NOT EXISTS idx_address ON addresses(address);
	
	CREATE TABLE IF NOT EXISTS transactions_history(
	   id INTEGER PRIMARY KEY,
	   timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
	   transaction_hash TEXT NOT NULL,
	   sender TEXT NOT NULL,
	   recipient TEXT NOT NULL,
	   amount REAL NOT NULL
		);
	`

	_, err = db.Exec(schema)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &Storage{db: db}, nil

}

func (s *Storage) IsEmpty() (bool, error) {
	const fn = "storage.sqlite.IsEmpty"

	stmt, err := s.db.Prepare("SELECT 1 FROM addresses LIMIT 1")
	if err != nil {
		return true, fmt.Errorf("%s: %w", fn, err)
	}

	var exists int

	err = stmt.QueryRow().Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return true, nil
		}
		return true, fmt.Errorf("%s: %w", fn, err)
	}

	return false, nil

}

func (s *Storage) IsExists(address string) (bool, error) {
	const fn = "storage.sqlite.IsExists"

	stmt, err := s.db.Prepare("SELECT 1 FROM addresses WHERE address = ?")
	if err != nil {
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	var exists int

	err = stmt.QueryRow(address).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("%s: %w", fn, err)
	}

	return true, nil
}

func (s *Storage) AddAddress(address string, balance float64) error {
	const fn = "storage.sqlite.AddAddress"

	stmt, err := s.db.Prepare("INSERT INTO addresses(address, balance) VALUES (?, ?)")
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	_, err = stmt.Exec(address, balance)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}
	return nil
}

func (s *Storage) GetBalance(address string) (float64, error) {
	const fn = "storage.sqlite.get_balance"

	stmt, err := s.db.Prepare("SELECT balance FROM addresses WHERE address = ?")
	if err != nil {
		return -1, fmt.Errorf("%s: %w", fn, err)
	}

	var balance float64

	err = stmt.QueryRow(address).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, storage.ErrAddressNotFound
		}

		return -1, fmt.Errorf("%s: %w", fn, err)
	}

	return balance, nil
}

func (s *Storage) GetLast(limit int) ([]storage.Transaction, error) {
	const fn = "storage.sqlite.GetLast"

	stmt, err := s.db.Prepare("SELECT * FROM transactions_history ORDER BY timestamp DESC LIMIT ?")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	rows, err := stmt.Query(limit)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	var transactions []storage.Transaction

	for rows.Next() {
		var tx storage.Transaction
		err := rows.Scan(&tx.Id, &tx.Timestamp, &tx.TransactionHash, &tx.Sender, &tx.Recipient, &tx.Amount)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", fn, err)
		}

		transactions = append(transactions, tx)
	}

	return transactions, nil
}

func (s *Storage) Send(sender, recipient string, amount float64) (string, error) {
	const fn = "storage.sqlite.send"

	senderBalance, err := s.GetBalance(sender)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}
	if amount > senderBalance {
		return "", storage.NotEnoughBalance
	}

	recipientBalance, err := s.GetBalance(recipient)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	_, err = tx.Exec(`UPDATE addresses SET balance = ? WHERE address = ?`, senderBalance-amount, sender)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	_, err = tx.Exec(`UPDATE addresses SET balance = ? WHERE address = ?`, recipientBalance+amount, recipient)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	txHash := utils.GenerateTransactionHash(sender, recipient, amount, time.Now().Unix())

	_, err = tx.Exec(`
	INSERT INTO transactions_history(timestamp, transaction_hash, sender, recipient, amount)
	VALUES (?, ?, ?, ?, ?)`,
		time.Now().Unix(), txHash, sender, recipient, amount,
	)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	return txHash, nil
}
