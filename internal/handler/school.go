package handler

import (
	"net/http"

	"luminous/internal/repository"
	"luminous/internal/response"

	"github.com/gin-gonic/gin"
)

type SchoolHandler struct {
	Repo repository.SchoolRepository
}

func NewSchoolHandler(repo repository.SchoolRepository) *SchoolHandler {
	return &SchoolHandler{Repo: repo}
}

func (h *SchoolHandler) ListSchools(c *gin.Context) {
	schools, err := h.Repo.FindEnabled(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list schools")
		return
	}
	response.SuccessList(c, http.StatusOK, "success", len(schools), schools)
}

func (h *SchoolHandler) GetSchool(c *gin.Context) {
	code := c.Param("code")
	school, err := h.Repo.FindByCode(c.Request.Context(), code)
	if err != nil {
		response.Error(c, http.StatusNotFound, "school not found")
		return
	}
	response.Success(c, http.StatusOK, "success", school)
}
