package handler

import (
	"net/http"

	"github.com/GlazedCurd/PlataTest/internal/db"
	"github.com/GlazedCurd/PlataTest/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	db        db.DB
	zapLogger *zap.Logger
}

func NewHandler(db db.DB, zapLogger *zap.Logger) *Handler {
	return &Handler{db: db, zapLogger: zapLogger}
}

func (h *Handler) GetLatest(c *gin.Context) {
	h.zapLogger.Info("Last update requested", zap.String("pair", c.Param("PAIR")))
	pair := c.Param("PAIR")
	lastUpdate, err := h.db.GetLastSuccessfulUpdate(model.Code(pair))
	if err != nil {
		h.zapLogger.Error("Failed to get last successful update", zap.String("pair", pair), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get last successful update"})
		return
	}
	c.JSON(http.StatusOK, lastUpdate)
}

func (h *Handler) RequestUpdate(c *gin.Context) {
	var update model.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	update.Code = model.Code(c.Param("PAIR"))
	h.zapLogger.Info("New update requested", zap.String("pair", c.Param("PAIR")), zap.String("idempotency_key", update.IdempotencyKey))
	insertedUpdate, err := h.db.InsertUpdate(&update)
	if err != nil {
		h.zapLogger.Error("Failed to insert update", zap.String("pair", c.Param("PAIR")), zap.String("idempotency_key", update.IdempotencyKey), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert update"})
		return
	}
	c.JSON(http.StatusOK, insertedUpdate)
}

func (h *Handler) GetUpdate(c *gin.Context) {
	updateId := c.GetUint64("UPDATE_ID")
	h.zapLogger.Info("Last update requested", zap.String("pair", c.Param("PAIR")), zap.Int("update_id", int(updateId)))
	update, err := h.db.GetUpdate(model.UpdateId(updateId))
	if err != nil {
		h.zapLogger.Error("Failed to get update", zap.String("pair", c.Param("PAIR")), zap.Int("update_id", int(updateId)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get update"})
		return
	}
	c.JSON(http.StatusOK, update)
}
