package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"luminous/internal/model"
	"luminous/internal/repository"
	"luminous/internal/response"

	"github.com/gin-gonic/gin"
)

func setupTest(t *testing.T) (*gin.Engine, *SchoolHandler) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	repo, err := repository.NewJSONSchoolRepository(t.TempDir() + "/schools.json")
	if err != nil {
		t.Fatal(err)
	}
	return gin.New(), NewSchoolHandler(repo)
}

func TestListSchoolsEmpty(t *testing.T) {
	r, h := setupTest(t)
	r.GET("/api/v1/schools", h.ListSchools)

	req := httptest.NewRequest("GET", "/api/v1/schools", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp response.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	listData := resp.Data.(map[string]interface{})
	if int(listData["total"].(float64)) != 0 {
		t.Fatalf("expected 0 schools, got %v", listData["total"])
	}
}

func TestGetSchoolNotFound(t *testing.T) {
	r, h := setupTest(t)
	r.GET("/api/v1/schools/:code", h.GetSchool)

	req := httptest.NewRequest("GET", "/api/v1/schools/NOEXIST", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestFeatureValidation(t *testing.T) {
	if model.IsValidFeature("invalid_feature") {
		t.Fatal("expected invalid feature to be rejected")
	}
	if !model.IsValidFeature(model.FeatureTimetable) {
		t.Fatal("expected valid feature to be accepted")
	}
}
