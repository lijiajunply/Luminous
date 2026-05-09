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

// ListSchools 需求1：返回所有启用的学校列表及总数
func (h *SchoolHandler) ListSchools(c *gin.Context) {
	schools, err := h.Repo.FindEnabled()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list schools")
		return
	}
	response.SuccessList(c, http.StatusOK, "success", len(schools), schools)
}

// GetSchool 需求2：根据代号返回学校详情及支持功能
func (h *SchoolHandler) GetSchool(c *gin.Context) {
	code := c.Param("code")
	school, err := h.Repo.FindByCode(code)
	if err != nil {
		response.Error(c, http.StatusNotFound, "school not found")
		return
	}
	response.Success(c, http.StatusOK, "success", school)
}
