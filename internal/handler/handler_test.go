package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GlazedCurd/PlataTest/internal/db"
	"github.com/GlazedCurd/PlataTest/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
	"go.uber.org/zap"
)

type dbMock struct {
	getConflictedTask         func(ctx context.Context, idempotencyKey string, code model.Code) (*model.Task, error)
	insertTask                func(ctx context.Context, task *model.Task) (*model.Task, error)
	getTask                   func(ctx context.Context, code model.Code, taskId model.TaskId) (*model.Task, error)
	taskTask                  func(ctx context.Context, task *model.Task) (*model.Task, error)
	getLastSuccessfulTask     func(ctx context.Context, code model.Code) (*model.Task, error)
	getRecentlyTasksToProcess func(ctx context.Context) ([]model.Task, error)
}

func NewDbMock() *dbMock {
	return &dbMock{
		getConflictedTask: func(ctx context.Context, idempotencyKey string, code model.Code) (*model.Task, error) {
			return nil, nil
		},
		insertTask: func(ctx context.Context, task *model.Task) (*model.Task, error) {
			return nil, nil
		},
		getTask: func(ctx context.Context, code model.Code, taskId model.TaskId) (*model.Task, error) {
			return nil, nil
		},
		taskTask: func(ctx context.Context, task *model.Task) (*model.Task, error) {
			return nil, nil
		},
		getLastSuccessfulTask: func(ctx context.Context, code model.Code) (*model.Task, error) {
			return nil, nil
		},
		getRecentlyTasksToProcess: func(ctx context.Context) ([]model.Task, error) {
			return nil, nil
		},
	}
}

func (d *dbMock) Close() error {
	return nil
}

func (d *dbMock) GetConflictedTask(ctx context.Context, idempotencyKey string, code model.Code) (*model.Task, error) {
	return d.getConflictedTask(ctx, idempotencyKey, code)
}

func (d *dbMock) InsertTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	return d.insertTask(ctx, task)
}

func (d *dbMock) GetTask(ctx context.Context, code model.Code, taskId model.TaskId) (*model.Task, error) {
	return d.getTask(ctx, code, taskId)
}

func (d *dbMock) UpdateTask(ctx context.Context, task *model.Task) (*model.Task, error) {
	return d.taskTask(ctx, task)
}

func (d *dbMock) GetLastSuccessfulTask(ctx context.Context, code model.Code) (*model.Task, error) {
	return d.getLastSuccessfulTask(ctx, code)
}

func (d *dbMock) GetRecentlyTasksToProcess(ctx context.Context) ([]model.Task, error) {
	return d.getRecentlyTasksToProcess(ctx)
}

func TestInsert(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	idempotencyKey := "abcd"
	price := float64(0.0)
	taskExpected := &model.Task{
		ID:             1,
		IdempotencyKey: idempotencyKey,
		Code:           "EUR_USD",
		Price:          &price,
		Status:         model.STATUS_SUCCESS,
	}

	dbmock.insertTask = func(ctx context.Context, task *model.Task) (*model.Task, error) {
		assert.Equal(t, task.IdempotencyKey, idempotencyKey)
		assert.Equal(t, task.Code, "EUR_USD")
		assert.Equal(t, task.ID, model.TaskId(0))
		return taskExpected, nil
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("create logger")
	}
	defer logger.Sync()
	SetupHandlers(r, dbmock, logger)

	w := httptest.NewRecorder()

	// Create an example user for testing
	request := struct {
		IdempotencyKey string `json:"idempotency_key,omitempty"`
	}{
		IdempotencyKey: "abcd",
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request %s", err)
	}
	pair := "EUR_USD"
	req, _ := http.NewRequest("POST", fmt.Sprintf("/quotes/%s/task", pair), strings.NewReader(string(requestJson)))
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)
	var response model.Task
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("unmarshal responce %s", err)
	}
	assert.Equal(t, response, taskExpected)
}

func TestGetLast(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	idempotencyKey := "abcd"
	price := float64(0.0)
	taskExpected := &model.Task{
		ID:             1,
		IdempotencyKey: idempotencyKey,
		Code:           "EUR_USD",
		Price:          &price,
		Status:         model.STATUS_SUCCESS,
	}

	dbmock.getLastSuccessfulTask = func(ctx context.Context, code model.Code) (*model.Task, error) {
		assert.Equal(t, code, "EUR_USD")
		return taskExpected, nil
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Creating logger %s", err)
	}
	SetupHandlers(r, dbmock, logger)

	w := httptest.NewRecorder()

	pair := "EUR_USD"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/quotes/%s", pair), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)
	var response model.Task
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Unmarshal response %s", err)
	}
	assert.Equal(t, response, *taskExpected)
}

func TestGetSpec(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	idempotencyKey := "abcd"
	price := float64(0.0)
	taskId := model.TaskId(1)
	taskExpected := &model.Task{
		ID:             taskId,
		IdempotencyKey: idempotencyKey,
		Code:           "EUR_USD",
		Price:          &price,
		Status:         model.STATUS_SUCCESS,
	}

	dbmock.getTask = func(ctx context.Context, code model.Code, task model.TaskId) (*model.Task, error) {
		assert.Equal(t, code, "EUR_USD")
		assert.Equal(t, task, taskId)
		return taskExpected, nil
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Creating logger %s", err)
	}
	SetupHandlers(r, dbmock, logger)

	w := httptest.NewRecorder()

	pair := "EUR_USD"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/quotes/%s/task/%d", pair, taskId), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)
	var response model.Task
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Unmarshal response %s", err)
	}
	assert.Equal(t, response, *taskExpected)
}

func TestInsertConflict(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	idempotencyKey := "abcd"

	dbmock.insertUpdate = func(ctx context.Context, update *model.Update) (*model.Update, error) {
		assert.Equal(t, update.IdempotencyKey, idempotencyKey)
		assert.Equal(t, update.Code, "EUR_USD")
		assert.Equal(t, update.ID, model.UpdateId(0))
		return nil, fmt.Errorf("conflict with different body: %w", db.ErrorConflictWithDifferentBody)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("failed to create logger")
	}
	SetupHandlers(r, dbmock, logger)

	w := httptest.NewRecorder()

	// Create an example user for testing
	request := struct {
		IdempotencyKey string `json:"idempotency_key,omitempty"`
	}{
		IdempotencyKey: "abcd",
	}
	requestJson, err := json.Marshal(request)
	if err != nil {
		t.Fatal("failed to marshal request")
	}
	pair := "EUR_USD"
	req, _ := http.NewRequest("POST", fmt.Sprintf("/quotes/%s/update", pair), strings.NewReader(string(requestJson)))
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 409)
}

func TestGetSpecNotFound(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	updateId := model.UpdateId(1)

	dbmock.getUpdate = func(ctx context.Context, code model.Code, update model.UpdateId) (*model.Update, error) {
		assert.Equal(t, code, "EUR_USD")
		assert.Equal(t, update, updateId)
		return nil, fmt.Errorf("not found: %w", db.ErrorNotFound)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("failed to create logger")
	}
	SetupHandlers(r, dbmock, logger)

	w := httptest.NewRecorder()

	pair := "EUR_USD"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/quotes/%s/update/%d", pair, updateId), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 404)
}
