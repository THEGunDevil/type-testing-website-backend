package models

import (
	"time"

	"github.com/google/uuid"
)

type Stat struct {
	PassageID           *uuid.UUID `json:"passage_id"`
	Title               string     `json:"title" binding:"required"`
	WPM                 float64    `json:"wpm" binding:"gte=0"`
	RawWPM              float64    `json:"raw_wpm" binding:"gte=0"`
	Accuracy            float64    `json:"accuracy" binding:"gte=0,lte=100"`
	DurationSeconds     int32      `json:"duration_s" binding:"gt=0"`
	CorrectWords        int32      `json:"correct_words" binding:"gte=0"`
	CorrectCharacters   int32      `json:"correct_characters" binding:"gte=0"`
	IncorrectCharacters int32      `json:"incorrect_characters" binding:"gte=0"`
	CreatedAt           time.Time  `json:"created_at"`
}
