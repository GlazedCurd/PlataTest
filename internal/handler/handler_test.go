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
	getConflictedUpdate         func(ctx context.Context, idempotencyKey string, code model.Code) (*model.Update, error)
	insertUpdate                func(ctx context.Context, update *model.Update) (*model.Update, error)
	getUpdate                   func(ctx context.Context, code model.Code, updateId model.UpdateId) (*model.Update, error)
	updateUpdate                func(ctx context.Context, update *model.Update) (*model.Update, error)
	getLastSuccessfulUpdate     func(ctx context.Context, code model.Code) (*model.Update, error)
	getRecentlyUpdatesToProcess func(ctx context.Context) ([]model.Update, error)
}

func NewDbMock() *dbMock {
	return &dbMock{
		getConflictedUpdate: func(ctx context.Context, idempotencyKey string, code model.Code) (*model.Update, error) {
			return nil, nil
		},
		insertUpdate: func(ctx context.Context, update *model.Update) (*model.Update, error) {
			return nil, nil
		},
		getUpdate: func(ctx context.Context, code model.Code, updateId model.UpdateId) (*model.Update, error) {
			return nil, nil
		},
		updateUpdate: func(ctx context.Context, update *model.Update) (*model.Update, error) {
			return nil, nil
		},
		getLastSuccessfulUpdate: func(ctx context.Context, code model.Code) (*model.Update, error) {
			return nil, nil
		},
		getRecentlyUpdatesToProcess: func(ctx context.Context) ([]model.Update, error) {
			return nil, nil
		},
	}
}

func (d *dbMock) Close() error {
	return nil
}

func (d *dbMock) GetConflictedUpdate(ctx context.Context, idempotencyKey string, code model.Code) (*model.Update, error) {
	return d.getConflictedUpdate(ctx, idempotencyKey, code)
}

func (d *dbMock) InsertUpdate(ctx context.Context, update *model.Update) (*model.Update, error) {
	return d.insertUpdate(ctx, update)
}

func (d *dbMock) GetUpdate(ctx context.Context, code model.Code, updateId model.UpdateId) (*model.Update, error) {
	return d.getUpdate(ctx, code, updateId)
}

func (d *dbMock) UpdateUpdate(ctx context.Context, update *model.Update) (*model.Update, error) {
	return d.updateUpdate(ctx, update)
}

func (d *dbMock) GetLastSuccessfulUpdate(ctx context.Context, code model.Code) (*model.Update, error) {
	return d.getLastSuccessfulUpdate(ctx, code)
}

func (d *dbMock) GetRecentlyUpdatesToProcess(ctx context.Context) ([]model.Update, error) {
	return d.getRecentlyUpdatesToProcess(ctx)
}

func TestInsert(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	idempotencyKey := "abcd"
	price := float64(0.0)
	updateExpected := &model.Update{
		ID:             1,
		IdempotencyKey: idempotencyKey,
		Code:           "EUR_USD",
		Price:          &price,
		Status:         model.STATUS_SUCCESS,
	}

	dbmock.insertUpdate = func(ctx context.Context, update *model.Update) (*model.Update, error) {
		assert.Equal(t, update.IdempotencyKey, idempotencyKey)
		assert.Equal(t, update.Code, "EUR_USD")
		assert.Equal(t, update.ID, model.UpdateId(0))
		return updateExpected, nil
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
	assert.Equal(t, w.Code, 200)
	var response model.Update
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal("failed to unmarshal responce")
	}
	assert.Equal(t, response, updateExpected)
}

func TestGetLast(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	idempotencyKey := "abcd"
	price := float64(0.0)
	updateExpected := &model.Update{
		ID:             1,
		IdempotencyKey: idempotencyKey,
		Code:           "EUR_USD",
		Price:          &price,
		Status:         model.STATUS_SUCCESS,
	}

	dbmock.getLastSuccessfulUpdate = func(ctx context.Context, code model.Code) (*model.Update, error) {
		assert.Equal(t, code, "EUR_USD")
		return updateExpected, nil
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatal("failed to create logger")
	}
	SetupHandlers(r, dbmock, logger)

	w := httptest.NewRecorder()

	pair := "EUR_USD"
	req, _ := http.NewRequest("GET", fmt.Sprintf("/quotes/%s", pair), nil)
	r.ServeHTTP(w, req)
	assert.Equal(t, w.Code, 200)
	var response model.Update
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal("failed to unmarshal responce")
	}
	assert.Equal(t, response, *updateExpected)
}

func TestGetSpec(t *testing.T) {
	r := gin.Default()
	dbmock := NewDbMock()

	idempotencyKey := "abcd"
	price := float64(0.0)
	updateId := model.UpdateId(1)
	updateExpected := &model.Update{
		ID:             updateId,
		IdempotencyKey: idempotencyKey,
		Code:           "EUR_USD",
		Price:          &price,
		Status:         model.STATUS_SUCCESS,
	}

	dbmock.getUpdate = func(ctx context.Context, code model.Code, update model.UpdateId) (*model.Update, error) {
		assert.Equal(t, code, "EUR_USD")
		assert.Equal(t, update, updateId)
		return updateExpected, nil
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
	assert.Equal(t, w.Code, 200)
	var response model.Update
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatal("failed to unmarshal responce")
	}
	assert.Equal(t, response, *updateExpected)
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
