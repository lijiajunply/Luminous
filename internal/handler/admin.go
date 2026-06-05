package handler

import (
	"net/http"
	"strconv"
	"strings"

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

func (h *AdminHandler) AdminListSchools(c *gin.Context) {
	schools, err := h.Repo.FindAll(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to list schools")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	total := len(schools)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	paged := schools[start:end]
	response.SuccessList(c, http.StatusOK, "success", total, paged)
}

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

	if err := h.Repo.Create(c.Request.Context(), school); err != nil {
		code := http.StatusInternalServerError
		if strings.Contains(err.Error(), "already exists") {
			code = http.StatusConflict
		}
		response.Error(c, code, err.Error())
		return
	}
	response.Success(c, http.StatusCreated, "school created", school)
}

func (h *AdminHandler) UpdateSchool(c *gin.Context) {
	code := c.Param("code")

	existing, err := h.Repo.FindByCode(c.Request.Context(), code)
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

	if err := h.Repo.Update(c.Request.Context(), existing); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		response.Error(c, status, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "school updated", existing)
}

func (h *AdminHandler) DeleteSchool(c *gin.Context) {
	code := c.Param("code")
	if err := h.Repo.Delete(c.Request.Context(), code); err != nil {
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "not found") {
			status = http.StatusNotFound
		}
		response.Error(c, status, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "school deleted", nil)
}
