package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"luminous/internal/model"
)

type SchoolRepository interface {
	FindAll(ctx context.Context) ([]*model.School, error)
	FindEnabled(ctx context.Context) ([]*model.School, error)
	FindByCode(ctx context.Context, code string) (*model.School, error)
	Create(ctx context.Context, school *model.School) error
	Update(ctx context.Context, school *model.School) error
	Delete(ctx context.Context, code string) error
}

type JSONSchoolRepository struct {
	mu      sync.RWMutex
	schools map[string]*model.School
	path    string
}

func NewJSONSchoolRepository(path string) (*JSONSchoolRepository, error) {
	repo := &JSONSchoolRepository{
		schools: make(map[string]*model.School),
		path:    path,
	}
	if err := repo.load(); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *JSONSchoolRepository) load() error {
	data, err := os.ReadFile(r.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read schools file: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &r.schools); err != nil {
		return fmt.Errorf("parse schools file: %w", err)
	}
	return nil
}

func (r *JSONSchoolRepository) save() error {
	data, err := json.MarshalIndent(r.schools, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal schools: %w", err)
	}
	if err := os.WriteFile(r.path, data, 0644); err != nil {
		return fmt.Errorf("write schools file: %w", err)
	}
	return nil
}

func (r *JSONSchoolRepository) FindAll(_ context.Context) ([]*model.School, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.School, 0, len(r.schools))
	for _, s := range r.schools {
		result = append(result, s)
	}
	return result, nil
}

func (r *JSONSchoolRepository) FindEnabled(_ context.Context) ([]*model.School, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*model.School, 0)
	for _, s := range r.schools {
		if s.Enabled {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *JSONSchoolRepository) FindByCode(_ context.Context, code string) (*model.School, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.schools[code]
	if !ok {
		return nil, fmt.Errorf("school not found: %s", code)
	}
	return s, nil
}

func (r *JSONSchoolRepository) Create(_ context.Context, school *model.School) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.schools[school.Code]; exists {
		return fmt.Errorf("school already exists: %s", school.Code)
	}

	now := time.Now()
	school.CreatedAt = now
	school.UpdatedAt = now
	if school.Features == nil {
		school.Features = []model.Feature{}
	}
	r.schools[school.Code] = school
	return r.save()
}

func (r *JSONSchoolRepository) Update(_ context.Context, school *model.School) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.schools[school.Code]
	if !ok {
		return fmt.Errorf("school not found: %s", school.Code)
	}

	school.CreatedAt = existing.CreatedAt
	school.UpdatedAt = time.Now()
	r.schools[school.Code] = school
	return r.save()
}

func (r *JSONSchoolRepository) Delete(_ context.Context, code string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.schools[code]; !ok {
		return fmt.Errorf("school not found: %s", code)
	}
	delete(r.schools, code)
	return r.save()
}
