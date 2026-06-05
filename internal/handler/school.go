package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"luminous/internal/model"
	"luminous/internal/repository"
	"luminous/internal/response"

	"github.com/gin-gonic/gin"
)

type SchoolHandler struct {
	repo repository.SchoolRepository
}

func NewSchoolHandler(repo repository.SchoolRepository) *SchoolHandler {
	return &SchoolHandler{repo: repo}
}

func (h *SchoolHandler) ListSchools(c *gin.Context) {
	schools, err := h.repo.FindEnabled(c.Request.Context())
	if err != nil {
		slog.Error("failed to list schools", "request_id", c.GetString("request_id"), "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to list schools")
		return
	}
	response.SuccessList(c, http.StatusOK, "success", len(schools), schools)
}

func (h *SchoolHandler) GetSchool(c *gin.Context) {
	code := c.Param("code")
	if !model.IsValidSchoolCode(code) {
		response.Error(c, http.StatusBadRequest, "invalid school code")
		return
	}
	school, err := h.repo.FindByCode(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "school not found")
		} else {
			slog.Error("failed to get school", "request_id", c.GetString("request_id"), "code", code, "error", err)
			response.Error(c, http.StatusInternalServerError, "failed to get school")
		}
		return
	}
	if !school.Enabled {
		response.Error(c, http.StatusNotFound, "school not found or disabled")
		return
	}
	response.Success(c, http.StatusOK, "success", school)
}
