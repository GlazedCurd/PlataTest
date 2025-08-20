package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/GlazedCurd/PlataTest/internal/model"
	"github.com/lib/pq"
)

const (
	pgConflictCode = "23505" // unique_violation
)

var (
	ErrorConflictWithDifferentBody = errors.New("conflict with different body")
	ErrorNotFound                  = errors.New("not found")
)

type DB interface {
	Close() error
	InsertTask(ctx context.Context, task *model.Task) (*model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) (*model.Task, error)
	GetTask(ctx context.Context, code model.Code, taskId model.TaskId) (*model.Task, error)
	GetLastSuccessfulTask(ctx context.Context, code model.Code) (*model.Task, error)
	GetRecentlyTasksToProcess(ctx context.Context) ([]model.Task, error)
}

type dbImpl struct {
	database *sql.DB
}

func ConnectDB(host, port, user, password, dbname string) (DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("database initialization %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("database ping %w", err)
	}

	log.Println("Successfully connected to the database")
	return &dbImpl{database: db}, nil
}

func (d *dbImpl) Close() error {
	return d.database.Close()
}

func (d *dbImpl) GetConflictedTask(ctx context.Context, idempotencyKey string, code model.Code) (*model.Task, error) {
	var task model.Task
	err := d.database.QueryRowContext(ctx, `
        SELECT id, code, idempotency_key, quote, status, created_at, updated_at
        FROM quotes
        WHERE idempotency_key = $1 AND code = $2
        LIMIT 1
    `, idempotencyKey, code).Scan(
		&task.ID,
		&task.Code,
		&task.IdempotencyKey,
		&task.Price,
		&task.Status,
		&task.CreatedAt,
		&task.TaskdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorConflictWithDifferentBody
		}
		return nil, fmt.Errorf("getting conflicted task: %w", err)
	}

	return &task, nil
}

func (d *dbImpl) InsertTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	var taskRes model.Task
	err := d.database.QueryRowContext(ctx, `
        INSERT INTO quotes (code, idempotency_key) 
        VALUES ($1, $2) 
        RETURNING id, code, idempotency_key, quote, status, created_at, updated_at
    `, task.Code, task.IdempotencyKey).Scan(
		&taskRes.ID,
		&taskRes.Code,
		&taskRes.IdempotencyKey,
		&taskRes.Price,
		&taskRes.Status,
		&taskRes.CreatedAt,
		&taskRes.TaskdAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case pgConflictCode:
				taskResP, err := d.GetConflictedTask(ctx, task.IdempotencyKey, task.Code)
				if err != nil {
					return nil, fmt.Errorf("get conflicted task: %w", err)
				}
				return taskResP, nil
			default:
				return nil, fmt.Errorf("database request [%s]: %s", pqErr.Code, pqErr.Message)
			}
		}
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, fmt.Errorf("insert and scan task: %w", err)
	}

	return &taskRes, nil
}

func (d *dbImpl) GetTask(ctx context.Context, code model.Code, taskId model.TaskId) (*model.Task, error) {
	var task model.Task
	err := d.database.QueryRowContext(ctx, `
        SELECT id, code, idempotency_key, quote, status, created_at, updated_at
        FROM quotes
        WHERE id = $1 AND code = $2
    `, taskId, code).Scan(
		&task.ID,
		&task.Code,
		&task.IdempotencyKey,
		&task.Price,
		&task.Status,
		&task.CreatedAt,
		&task.TaskdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &task, nil
}

func (d *dbImpl) UpdateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	var updatedRes model.Task
	err := d.database.QueryRowContext(ctx, `
        UPDATE quotes 
        SET status = $1,
            quote = $2,
            updated_at = CURRENT_TIMESTAMP
        WHERE id = $3
        RETURNING id, code, idempotency_key, quote, status, created_at, updated_at
    `, task.Status, task.Price, task.ID).Scan(
		&updatedRes.ID,
		&updatedRes.Code,
		&updatedRes.IdempotencyKey,
		&updatedRes.Price,
		&updatedRes.Status,
		&updatedRes.CreatedAt,
		&updatedRes.TaskdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, fmt.Errorf("task quote: %w", err)
	}

	return &updatedRes, nil
}

func (d *dbImpl) GetLastSuccessfulTask(ctx context.Context, code model.Code) (*model.Task, error) {
	var task model.Task
	err := d.database.QueryRowContext(ctx, `
        SELECT id, code, idempotency_key, quote, status, created_at, updated_at
        FROM quotes
        WHERE code = $1 AND status = 'success'
        ORDER BY updated_at DESC
        LIMIT 1
    `, code).Scan(
		&task.ID,
		&task.Code,
		&task.IdempotencyKey,
		&task.Price,
		&task.Status,
		&task.CreatedAt,
		&task.TaskdAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrorNotFound
		}
		return nil, fmt.Errorf("get last successful task: %w", err)
	}

	return &task, nil
}

func (d *dbImpl) GetRecentlyTasksToProcess(ctx context.Context) ([]model.Task, error) {
	var tasks []model.Task
	rows, err := d.database.QueryContext(ctx, `
        SELECT id, code, idempotency_key, quote, status, created_at, updated_at	
        FROM quotes
        WHERE status = 'pending'
        ORDER BY created_at DESC
    `)
	if err != nil {
		return nil, fmt.Errorf("get recently requested tasks: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var task model.Task
		if err := rows.Scan(
			&task.ID,
			&task.Code,
			&task.IdempotencyKey,
			&task.Price,
			&task.Status,
			&task.CreatedAt,
			&task.TaskdAt,
		); err != nil {
			return nil, fmt.Errorf("scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
