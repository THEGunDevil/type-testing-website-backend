package handlers

import (
    "errors"
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5"
)

func (h *UserHandler) GetPassage(c *gin.Context) {
    language := c.Query("language")

    if language == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Language is required",
        })
        return
    }

    passage, err := h.DB.GetRandomPassage(
        c.Request.Context(),
        language,
    )

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "No passage found for this language",
            })
            return
        }

        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to fetch passage",
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "passage": passage,
    })
}