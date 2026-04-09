package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"gomonitor/backend/model"
	"gomonitor/backend/pkg/profiler"
	"gomonitor/backend/service"
)

type ProfileHandler struct {
	service *service.ProfileService
}

func NewProfileHandler(svc *service.ProfileService) *ProfileHandler {
	return &ProfileHandler{service: svc}
}

func (h *ProfileHandler) Capture(c *gin.Context) {
	var req model.ProfileCaptureRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	response, err := h.service.CaptureSummary(req)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unsupported profile type") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *ProfileHandler) Download(c *gin.Context) {
	profileType := strings.TrimSpace(strings.ToLower(c.Param("type")))
	seconds, _ := strconv.Atoi(c.DefaultQuery("seconds", "10"))

	result, err := h.service.CaptureRaw(profileType, seconds)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unsupported") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	filename := "profile-" + result.Type + ".pprof"
	if result.Type == "cpu" {
		filename = "profile-cpu-" + strconv.Itoa(result.DurationSeconds) + "s.pprof"
	}

	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", result.Data)
}

func (h *ProfileHandler) Types(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"types": profiler.SupportedTypes()})
}
