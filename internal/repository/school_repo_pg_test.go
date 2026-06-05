//go:build integration

package repository

import (
	"context"
	"os"
	"testing"

	"luminous/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

func setupPGTest(t *testing.T) *PGSchoolRepository {
	t.Helper()

	dsn := os.Getenv("LUMINOUS_TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://luminous:luminous@localhost:5432/luminous?sslmode=disable"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("skipping PG integration test (no database): %v", err)
		return nil
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("skipping PG integration test (ping failed): %v", err)
		return nil
	}

	pool.Exec(ctx, "DROP TABLE IF EXISTS schools CASCADE")

	repo := &PGSchoolRepository{pool: pool}
	if err := repo.autoMigrate(ctx); err != nil {
		pool.Close()
		t.Fatalf("migrate: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, "DROP TABLE IF EXISTS schools CASCADE")
		pool.Close()
	})

	return repo
}

func TestPGCreateAndFindAll(t *testing.T) {
	repo := setupPGTest(t)

	school := &model.School{
		Code:     "TEST",
		Name:     "Test University",
		Website:  "https://test.edu",
		Features: []model.Feature{model.FeatureTimetable},
		Enabled:  true,
	}
	if err := repo.Create(testCtx, school); err != nil {
		t.Fatalf("Create: %v", err)
	}

	all, err := repo.FindAll(testCtx)
	if err != nil {
		t.Fatalf("FindAll: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 school, got %d", len(all))
	}
	if all[0].Code != "TEST" {
		t.Fatalf("expected TEST, got %s", all[0].Code)
	}
	if len(all[0].Features) != 1 || all[0].Features[0] != model.FeatureTimetable {
		t.Fatalf("expected timetable feature, got %v", all[0].Features)
	}
}

func TestPGFindByCode(t *testing.T) {
	repo := setupPGTest(t)

	repo.Create(testCtx, &model.School{
		Code: "FIND", Name: "Find Me", Website: "https://find.edu",
		Features: []model.Feature{}, Enabled: true,
	})

	found, err := repo.FindByCode(testCtx, "FIND")
	if err != nil {
		t.Fatalf("FindByCode: %v", err)
	}
	if found.Name != "Find Me" {
		t.Fatalf("unexpected name: %s", found.Name)
	}

	_, err = repo.FindByCode(testCtx, "MISSING")
	if err == nil {
		t.Fatal("expected error for missing school")
	}
}

func TestPGFindEnabled(t *testing.T) {
	repo := setupPGTest(t)

	repo.Create(testCtx, &model.School{Code: "ON", Name: "On", Website: "https://a.edu", Features: nil, Enabled: true})
	repo.Create(testCtx, &model.School{Code: "OFF", Name: "Off", Website: "https://b.edu", Features: nil, Enabled: false})

	enabled, err := repo.FindEnabled(testCtx)
	if err != nil {
		t.Fatalf("FindEnabled: %v", err)
	}
	if len(enabled) != 1 {
		t.Fatalf("expected 1 enabled, got %d", len(enabled))
	}
	if enabled[0].Code != "ON" {
		t.Fatalf("expected ON, got %s", enabled[0].Code)
	}
}

func TestPGUpdate(t *testing.T) {
	repo := setupPGTest(t)

	repo.Create(testCtx, &model.School{Code: "UPD", Name: "Old", Website: "https://old.edu", Features: nil, Enabled: true})

	updated := &model.School{Code: "UPD", Name: "New", Website: "https://new.edu", Features: nil, Enabled: false}
	if err := repo.Update(testCtx, updated); err != nil {
		t.Fatalf("Update: %v", err)
	}

	found, _ := repo.FindByCode(testCtx, "UPD")
	if found.Name != "New" {
		t.Fatalf("expected New, got %s", found.Name)
	}
	if found.Website != "https://new.edu" {
		t.Fatalf("expected new url, got %s", found.Website)
	}
}

func TestPGDelete(t *testing.T) {
	repo := setupPGTest(t)

	repo.Create(testCtx, &model.School{Code: "DEL", Name: "Delete Me", Website: "https://del.edu", Features: nil, Enabled: true})

	if err := repo.Delete(testCtx, "DEL"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := repo.FindByCode(testCtx, "DEL")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestPGCreateDuplicate(t *testing.T) {
	repo := setupPGTest(t)

	school := &model.School{Code: "DUP", Name: "Dup", Website: "https://dup.edu", Features: nil, Enabled: true}
	if err := repo.Create(testCtx, school); err != nil {
		t.Fatal(err)
	}

	if err := repo.Create(testCtx, school); err == nil {
		t.Fatal("expected error for duplicate code")
	}
}
