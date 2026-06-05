package handler

import (
	"errors"
	"log/slog"
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}

	schools, err := h.Repo.FindAll(c.Request.Context(), (page-1)*pageSize, pageSize)
	if err != nil {
		slog.Error("failed to list schools", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to list schools")
		return
	}

	total, err := h.Repo.Count(c.Request.Context())
	if err != nil {
		slog.Error("failed to count schools", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to count schools")
		return
	}

	response.SuccessList(c, http.StatusOK, "success", total, schools)
}

func (h *AdminHandler) CreateSchool(c *gin.Context) {
	var req model.CreateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if !model.IsValidSchoolCode(req.Code) {
		response.Error(c, http.StatusBadRequest, "invalid school code: must be 1-20 chars, uppercase alphanumeric, hyphens or underscores")
		return
	}
	if !model.IsValidURL(req.Website) {
		response.Error(c, http.StatusBadRequest, "invalid website URL")
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
		if strings.Contains(err.Error(), "already exists") {
			response.Error(c, http.StatusConflict, "school already exists")
			return
		}
		slog.Error("failed to create school", "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to create school")
		return
	}
	response.Success(c, http.StatusCreated, "school created", school)
}

func (h *AdminHandler) UpdateSchool(c *gin.Context) {
	code := c.Param("code")
	if !model.IsValidSchoolCode(code) {
		response.Error(c, http.StatusBadRequest, "invalid school code")
		return
	}

	existing, err := h.Repo.FindByCode(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.Error(c, http.StatusNotFound, "school not found")
		} else {
			slog.Error("failed to get school for update", "code", code, "error", err)
			response.Error(c, http.StatusInternalServerError, "failed to get school")
		}
		return
	}

	var req model.UpdateSchoolRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.Website != nil {
		if !model.IsValidURL(*req.Website) {
			response.Error(c, http.StatusBadRequest, "invalid website URL")
			return
		}
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
		slog.Error("failed to update school", "code", code, "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to update school")
		return
	}
	response.Success(c, http.StatusOK, "school updated", existing)
}

func (h *AdminHandler) DeleteSchool(c *gin.Context) {
	code := c.Param("code")
	if !model.IsValidSchoolCode(code) {
		response.Error(c, http.StatusBadRequest, "invalid school code")
		return
	}
	if err := h.Repo.Delete(c.Request.Context(), code); err != nil {
		slog.Error("failed to delete school", "code", code, "error", err)
		response.Error(c, http.StatusInternalServerError, "failed to delete school")
		return
	}
	response.Success(c, http.StatusOK, "school deleted", nil)
}
