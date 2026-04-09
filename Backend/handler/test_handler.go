package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"gomonitor/backend/model"
	"gomonitor/backend/service"
)

type TestHandler struct {
	service *service.TestService
}

func NewTestHandler(svc *service.TestService) *TestHandler {
	return &TestHandler{service: svc}
}

func (h *TestHandler) RunTest(c *gin.Context) {
	var req model.RunTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	urls := normalizeAndFilterURLs(req.URLs)
	if len(urls) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one valid URL is required"})
		return
	}

	headers := normalizeHeaders(req.Headers)

	response := h.service.RunTest(c.Request.Context(), urls, headers)
	c.JSON(http.StatusOK, response)
}

func normalizeAndFilterURLs(urls []string) []string {
	cleaned := make([]string, 0, len(urls))
	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		if u == "" {
			continue
		}
		cleaned = append(cleaned, u)
	}
	return cleaned
}

func normalizeHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	cleaned := make(map[string]string, len(headers))
	for k, v := range headers {
		key := strings.TrimSpace(k)
		value := strings.TrimSpace(v)
		if key == "" || value == "" {
			continue
		}
		cleaned[key] = value
	}

	if len(cleaned) == 0 {
		return nil
	}

	return cleaned
}
