package handlers

import (
	"log"
	"net/http"
	"typetype/internal/database"
	"typetype/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func (h *UserHandler) CreateStats(c *gin.Context) {
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}

	var req models.Stat
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Stats Failed] Validation error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var passageID pgtype.UUID
	if req.PassageID != nil {
		passageID = pgtype.UUID{
			Bytes: *req.PassageID,
			Valid: true,
		}
	}

	stat, err := h.DB.CreateTestResult(
		c.Request.Context(),
		database.CreateTestResultParams{
			UserID:              userID,
			PassageID:           passageID,
			Title:               req.Title,
			Wpm:                 float32(req.WPM),
			RawWpm:              float32(req.RawWPM),
			Accuracy:            float32(req.Accuracy),
			DurationS:           req.DurationSeconds,
			CorrectWords:        req.CorrectWords,
			CorrectCharacters:   req.CorrectCharacters,
			IncorrectCharacters: req.IncorrectCharacters,
		},
	)
	if err != nil {
		log.Printf("[DB Error] Failed saving stats: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save stats",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Stats saved successfully",
		"stats":   stat,
	})
}

func (h *UserHandler) GetStats(c *gin.Context) {
	userIDStr := c.GetString("userID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
		return
	}
	// Fetch stats from database
	stats, err := h.DB.GetUserStats(
		c,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch stats",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}