package repository

import (
	"context"
	"fmt"
	"time"

	"luminous/internal/config"
	"luminous/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGSchoolRepository struct {
	pool *pgxpool.Pool
}

func NewPGSchoolRepository(ctx context.Context, dbConfig config.DatabaseConfig) (*PGSchoolRepository, error) {
	dsn := dbConfig.DSN
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			dbConfig.User, dbConfig.Password, dbConfig.Host, dbConfig.Port, dbConfig.DBName, dbConfig.SSLMode)
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}
	poolCfg.MaxConns = dbConfig.PoolMaxConns
	if poolCfg.MaxConns == 0 {
		poolCfg.MaxConns = 20
	}
	poolCfg.MinConns = dbConfig.PoolMinConns

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	repo := &PGSchoolRepository{pool: pool}
	if err := repo.autoMigrate(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("migrate database: %w", err)
	}

	return repo, nil
}

func (r *PGSchoolRepository) Close() {
	r.pool.Close()
}

func (r *PGSchoolRepository) autoMigrate(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schools (
			code       TEXT PRIMARY KEY,
			name       TEXT NOT NULL,
			website    TEXT NOT NULL DEFAULT '',
			features   TEXT[] NOT NULL DEFAULT '{}',
			enabled    BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return err
	}

	// Ensure columns exist for forward-compatible schema evolution.
	// New columns added in the future should use ALTER TABLE ... ADD COLUMN IF NOT EXISTS.
	columns := map[string]string{
		"website":    "TEXT NOT NULL DEFAULT ''",
		"features":   "TEXT[] NOT NULL DEFAULT '{}'",
		"enabled":    "BOOLEAN NOT NULL DEFAULT true",
		"created_at": "TIMESTAMPTZ NOT NULL DEFAULT NOW()",
		"updated_at": "TIMESTAMPTZ NOT NULL DEFAULT NOW()",
	}
	for col, def := range columns {
		_, err := r.pool.Exec(ctx, fmt.Sprintf(
			`DO $$ BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_name='schools' AND column_name='%s'
				) THEN
					ALTER TABLE schools ADD COLUMN %s %s;
				END IF;
			END $$;`, col, col, def))
		if err != nil {
			return fmt.Errorf("migrate column %s: %w", col, err)
		}
	}

	return nil
}

func (r *PGSchoolRepository) FindAll(ctx context.Context) ([]*model.School, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT code, name, website, features, enabled, created_at, updated_at
		 FROM schools ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("find all schools: %w", err)
	}
	defer rows.Close()

	var schools []*model.School
	for rows.Next() {
		s, err := scanSchool(rows)
		if err != nil {
			return nil, err
		}
		schools = append(schools, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate schools: %w", err)
	}
	return schools, nil
}

func (r *PGSchoolRepository) FindEnabled(ctx context.Context) ([]*model.School, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT code, name, website, features, enabled, created_at, updated_at
		 FROM schools WHERE enabled = true ORDER BY code`)
	if err != nil {
		return nil, fmt.Errorf("find enabled schools: %w", err)
	}
	defer rows.Close()

	var schools []*model.School
	for rows.Next() {
		s, err := scanSchool(rows)
		if err != nil {
			return nil, err
		}
		schools = append(schools, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate enabled schools: %w", err)
	}
	return schools, nil
}

func (r *PGSchoolRepository) FindByCode(ctx context.Context, code string) (*model.School, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT code, name, website, features, enabled, created_at, updated_at
		 FROM schools WHERE code = $1`, code)
	return scanSchool(row)
}

func (r *PGSchoolRepository) Create(ctx context.Context, school *model.School) error {
	now := time.Now()
	school.CreatedAt = now
	school.UpdatedAt = now
	if school.Features == nil {
		school.Features = []model.Feature{}
	}

	tag, err := r.pool.Exec(ctx,
		`INSERT INTO schools (code, name, website, features, enabled, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (code) DO NOTHING`,
		school.Code, school.Name, school.Website,
		featuresToStrings(school.Features), school.Enabled,
		school.CreatedAt, school.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create school: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("school already exists: %s", school.Code)
	}
	return nil
}

func (r *PGSchoolRepository) Update(ctx context.Context, school *model.School) error {
	school.UpdatedAt = time.Now()

	tag, err := r.pool.Exec(ctx,
		`UPDATE schools
		 SET name=$1, website=$2, features=$3, enabled=$4, updated_at=$5
		 WHERE code=$6`,
		school.Name, school.Website, featuresToStrings(school.Features),
		school.Enabled, school.UpdatedAt, school.Code)
	if err != nil {
		return fmt.Errorf("update school: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("school not found: %s", school.Code)
	}
	return nil
}

func (r *PGSchoolRepository) Delete(ctx context.Context, code string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM schools WHERE code = $1`, code)
	if err != nil {
		return fmt.Errorf("delete school: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("school not found: %s", code)
	}
	return nil
}

func scanSchool(row pgx.Row) (*model.School, error) {
	var s model.School
	var features []string
	err := row.Scan(&s.Code, &s.Name, &s.Website, &features,
		&s.Enabled, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("school not found")
		}
		return nil, fmt.Errorf("scan school: %w", err)
	}
	s.Features = stringsToFeatures(features)
	return &s, nil
}

func featuresToStrings(features []model.Feature) []string {
	result := make([]string, len(features))
	for i, f := range features {
		result[i] = string(f)
	}
	return result
}

func stringsToFeatures(strs []string) []model.Feature {
	result := make([]model.Feature, len(strs))
	for i, s := range strs {
		result[i] = model.Feature(s)
	}
	return result
}
