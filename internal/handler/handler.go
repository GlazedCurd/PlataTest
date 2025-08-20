package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/GlazedCurd/PlataTest/internal/db"
	"github.com/GlazedCurd/PlataTest/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	db        db.DB
	zapLogger *zap.Logger
}

func SetupHandlers(r *gin.Engine, db db.DB, zapLogger *zap.Logger) {
	h := &Handler{db: db, zapLogger: zapLogger}
	// Set up routes
	r.GET("/quotes/:PAIR", h.GetLatest)
	r.POST("/quotes/:PAIR/task", h.RequestTask)
	r.GET("/quotes/:PAIR/task/:TASK_ID", h.GetTask)
}

func (h *Handler) GetLatest(c *gin.Context) {
	h.zapLogger.Info("Last task requested", zap.String("pair", c.Param("PAIR")))
	pair := c.Param("PAIR")
	lastTask, err := h.db.GetLastSuccessfulTask(c.Request.Context(), model.Code(pair))
	if err != nil {
		if errors.Is(err, db.ErrorNotFound) {
			h.zapLogger.Error("Task not found", zap.String("pair", c.Param("PAIR")))
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		h.zapLogger.Error("get last successful task", zap.String("pair", pair), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get last successful task"})
		return
	}
	c.JSON(http.StatusOK, lastTask)
}

func (h *Handler) RequestTask(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	task.Code = model.Code(c.Param("PAIR"))
	h.zapLogger.Info("New task requested", zap.String("pair", c.Param("PAIR")), zap.String("idempotency_key", task.IdempotencyKey))
	insertedTask, err := h.db.InsertTask(c.Request.Context(), &task)
	if err != nil {
		if errors.Is(err, db.ErrorConflictWithDifferentBody) {
			h.zapLogger.Error("Conflict with different body", zap.String("pair", c.Param("PAIR")), zap.String("idempotency_key", task.IdempotencyKey))
			c.JSON(http.StatusConflict, gin.H{"error": "Conflict with different body"})
			return
		}
		h.zapLogger.Error("insert task", zap.String("pair", c.Param("PAIR")), zap.String("idempotency_key", task.IdempotencyKey), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert task"})
		return
	}
	c.JSON(http.StatusOK, insertedTask)
}

func (h *Handler) GetTask(c *gin.Context) {
	taskId, err := strconv.Atoi(c.Param("TASK_ID"))
	if err != nil {
		h.zapLogger.Error("Invalid task ID", zap.String("pair", c.Param("PAIR")), zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}
	h.zapLogger.Info("Task requested", zap.String("pair", c.Param("PAIR")), zap.Int("task_id", int(taskId)))
	task, err := h.db.GetTask(c.Request.Context(), model.Code(c.Param("PAIR")), model.TaskId(taskId))
	if err != nil {
		if errors.Is(err, db.ErrorNotFound) {
			h.zapLogger.Error("Task not found", zap.String("pair", c.Param("PAIR")), zap.Int("task_id", int(taskId)))
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
			return
		}
		h.zapLogger.Error("get task", zap.String("pair", c.Param("PAIR")), zap.Int("task_id", int(taskId)), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get task"})
		return
	}
	c.JSON(http.StatusOK, task)
}
