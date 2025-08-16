package handler

import (
	"database/sql"
	"net/http"

	"github.com/GlazedCurd/PlataTest/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	db        *sql.DB
	zapLogger *zap.Logger
}

func NewHandler(db *sql.DB, zapLogger *zap.Logger) *Handler {
	return &Handler{db: db, zapLogger: zapLogger}
}

func (h *Handler) GetLatest(c *gin.Context) {
	var update model.Update
	h.zapLogger.Info("Last update requested", zap.String("pair", c.Param("PAIR")))
	c.JSON(http.StatusOK, update)
}

func (h *Handler) RequestUpdate(c *gin.Context) {
	var update model.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.zapLogger.Info("New update requested", zap.String("pair", c.Param("PAIR")), zap.String("idempotency_key", update.IdempotencyKey))
	c.JSON(http.StatusOK, update)
}

func (h *Handler) GetUpdate(c *gin.Context) {
	var update model.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.zapLogger.Info("Last update requested", zap.String("pair", c.Param("PAIR")), zap.Int("update_id", update.ID))
	c.JSON(http.StatusOK, update)
}
