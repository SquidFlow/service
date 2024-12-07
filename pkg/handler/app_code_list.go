package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AppCodeResponse struct {
	Success  bool     `json:"success"`
	Message  string   `json:"message"`
	AppCodes []string `json:"appCodes"`
}

// AppCodeList handles the request to list app codes
func AppCodeList(c *gin.Context) {
	appCodes := []string{"esfs", "esfs-dev", "esfs-test"}
	c.JSON(http.StatusOK, AppCodeResponse{
		Success:  true,
		Message:  "App codes listed successfully",
		AppCodes: appCodes,
	})
	return
}
