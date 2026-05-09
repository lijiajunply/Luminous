package handler

import (
	"net/http"

	"luminous/internal/model"
	"luminous/internal/repository"
	"luminous/internal/response"

	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	Repo repository.SchoolRepository
}

func NewAdminHandler(repo repository.SchoolRepository) *AdminHandler {
	return &AdminHandler{Repo: repo}
}

// AdminListSchools 管理员列出所有学校（含未启用）
func (h *AdminHandler) AdminListSchools(c *gin.Context) {
	schools, err := h.Repo.FindAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list schools")
		return
	}
	response.SuccessList(c, http.StatusOK, "success", len(schools), schools)
}

// CreateSchool 管理员新增学校
func (h *AdminHandler) CreateSchool(c *gin.Context) {
	var req model.CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	for _, f := range req.Features {
		if !model.IsValidFeature(f) {
			response.Error(c, http.StatusBadRequest, "invalid feature: "+string(f))
			return
		}
	}

	school := &model.School{
		Code:     req.Code,
		Name:     req.Name,
		Website:  req.Website,
		Features: req.Features,
		Enabled:  true,
	}

	if err := h.Repo.Create(school); err != nil {
		response.Error(c, http.StatusConflict, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "school created", school)
}

// UpdateSchool 管理员更新学校（部分更新）
func (h *AdminHandler) UpdateSchool(c *gin.Context) {
	code := c.Param("code")

	existing, err := h.Repo.FindByCode(code)
	if err != nil {
		response.Error(c, http.StatusNotFound, "school not found")
		return
	}

	var req model.UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Website != nil {
		existing.Website = *req.Website
	}
	if req.Features != nil {
		for _, f := range *req.Features {
			if !model.IsValidFeature(f) {
				response.Error(c, http.StatusBadRequest, "invalid feature: "+string(f))
				return
			}
		}
		existing.Features = *req.Features
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := h.Repo.Update(existing); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update school")
		return
	}
	response.Success(c, http.StatusOK, "school updated", existing)
}

// DeleteSchool 管理员删除学校
func (h *AdminHandler) DeleteSchool(c *gin.Context) {
	code := c.Param("code")
	if err := h.Repo.Delete(code); err != nil {
		response.Error(c, http.StatusNotFound, "school not found")
		return
	}
	response.Success(c, http.StatusOK, "school deleted", nil)
}
