package db

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/GlazedCurd/PlataTest/internal/model"
	"github.com/lib/pq"
)

type DB interface {
	Close() error
	InsertUpdate(update *model.Update) (*model.Update, error)
	UpdateUpdate(update *model.Update) (*model.Update, error)
	GetUpdate(updateId model.UpdateId) (*model.Update, error)
	GetLastSuccessfulUpdate(code model.Code) (*model.Update, error)
}

type dbImpl struct {
	database *sql.DB
}

func ConnectDB(host, port, user, password, dbname string) (DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("database initialization failed %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database ping failed %w", err)
	}

	log.Println("Successfully connected to the database")
	return &dbImpl{database: db}, nil
}

func (d *dbImpl) Close() error {
	return d.database.Close()
}

func (d *dbImpl) GetConflictedUpdate(idempotencyKey string, code model.Code) (*model.Update, error) {
	var update model.Update
	err := d.database.QueryRow(`
        SELECT id, code, idempotency_key, quote, status, created_at, updated_at
        FROM quotes
        WHERE idempotency_key = $1 AND code = $2
        LIMIT 1
    `, idempotencyKey, code).Scan(
		&update.ID,
		&update.Code,
		&update.IdempotencyKey,
		&update.Price,
		&update.Status,
		&update.CreatedAt,
		&update.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No conflict found
		}
		return nil, fmt.Errorf("failed to check for conflicted update: %w", err)
	}

	return &update, nil
}

func (d *dbImpl) InsertUpdate(update *model.Update) (*model.Update, error) {
	var updateRes model.Update
	err := d.database.QueryRow(`
        INSERT INTO quotes (code, idempotency_key) 
        VALUES ($1, $2) 
        RETURNING id, code, idempotency_key, quote, status, created_at, updated_at
    `, update.Code, update.IdempotencyKey).Scan(
		&updateRes.ID,
		&updateRes.Code,
		&updateRes.IdempotencyKey,
		&updateRes.Price,
		&updateRes.Status,
		&updateRes.CreatedAt,
		&updateRes.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				updateResP, err := d.GetConflictedUpdate(update.IdempotencyKey, update.Code)
				if err != nil {
					return nil, fmt.Errorf("failed to get conflicted update: %w", err)
				}
				return updateResP, nil
			default:
				return nil, fmt.Errorf("database error [%s]: %s", pqErr.Code, pqErr.Message)
			}
		}
		return nil, fmt.Errorf("failed to insert and scan update: %w", err)
	}

	return &updateRes, nil
}

func (d *dbImpl) GetUpdate(updateId model.UpdateId) (*model.Update, error) {
	var update model.Update
	err := d.database.QueryRow(`
        SELECT id, code, idempotency_key, quote, status, created_at, updated_at
        FROM quotes
        WHERE id = $1
    `, updateId).Scan(
		&update.ID,
		&update.Code,
		&update.IdempotencyKey,
		&update.Price,
		&update.Status,
		&update.CreatedAt,
		&update.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get update: %w", err)
	}

	return &update, nil
}

func (d *dbImpl) UpdateUpdate(update *model.Update) (*model.Update, error) {
	var updatedRes model.Update
	// TODO: use context
	err := d.database.QueryRow(`
        UPDATE quotes 
        SET status = $1,
            quote = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
        RETURNING id, code, idempotency_key, quote, status, created_at, updated_at
    `, update.Status, update.Price, update.ID).Scan(
		&updatedRes.ID,
		&updatedRes.Code,
		&updatedRes.IdempotencyKey,
		&updatedRes.Price,
		&updatedRes.Status,
		&updatedRes.CreatedAt,
		&updatedRes.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("update not found: %w", err)
		}
		return nil, fmt.Errorf("failed to update quote: %w", err)
	}

	return &updatedRes, nil
}

func (d *dbImpl) GetLastSuccessfulUpdate(code model.Code) (*model.Update, error) {
	var update model.Update
	err := d.database.QueryRow(`
        SELECT id, code, idempotency_key, quote, status, created_at, updated_at
        FROM quotes
        WHERE code = $1 AND status = 'completed'
        ORDER BY created_at DESC
        LIMIT 1
    `, code).Scan(
		&update.ID,
		&update.Code,
		&update.IdempotencyKey,
		&update.Price,
		&update.Status,
		&update.CreatedAt,
		&update.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get last successful update: %w", err)
	}

	return &update, nil
}
