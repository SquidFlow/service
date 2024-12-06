package handler

import (
	"github.com/google/uuid"
)

// getNewId returns a new id for the resource
func getNewId() string {
	return uuid.New().String()
}
