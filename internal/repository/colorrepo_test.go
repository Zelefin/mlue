package repository_test

import (
	"testing"

	"mlue/internal/models"
	"mlue/internal/repository"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (repository.ColorRepo, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory sqlite: %v", err)
	}

	// Migrate the schema
	if err := db.AutoMigrate(&models.Color{}); err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}

	repo := repository.NewColorRepo(db)
	return repo, db
}

func TestCreateAndGet(t *testing.T) {
	repo, _ := setupTestDB(t)

	col := &models.Color{
		UserID:        42,
		Hex:           "#abcdef",
		UserColorName: "MyColor",
		RealColorName: "Cobalt",
		Match:         true,
		Palette:       "#123456,#789abc",
	}

	// Create
	err := repo.Create(col)
	assert.NoError(t, err)
	assert.NotZero(t, col.ID, "expected ID to be set after Create")

	// Get by ID
	fetched, err := repo.Get(col.ID)
	assert.NoError(t, err)
	assert.Equal(t, col.Hex, fetched.Hex)
	assert.Equal(t, col.UserID, fetched.UserID)
}

func TestGetAllAndGetByUser(t *testing.T) {
	repo, _ := setupTestDB(t)

	// seed multiple
	colors := []models.Color{
		{UserID: 1, Hex: "#111111"},
		{UserID: 2, Hex: "#222222"},
		{UserID: 1, Hex: "#333333"},
	}
	for i := range colors {
		err := repo.Create(&colors[i])
		assert.NoError(t, err)
	}

	all, err := repo.GetAll()
	assert.NoError(t, err)
	assert.Len(t, all, 3)

	user1, err := repo.GetByUser(1)
	assert.NoError(t, err)
	assert.Len(t, user1, 2)
	// ensure correct hex values
	expected := map[string]bool{"#111111": true, "#333333": true}
	for _, c := range user1 {
		assert.True(t, expected[c.Hex])
	}
}

func TestUpdate(t *testing.T) {
	repo, _ := setupTestDB(t)

	col := &models.Color{UserID: 5, Hex: "#aaaaaa"}
	err := repo.Create(col)
	assert.NoError(t, err)

	// modify and update
	col.Hex = "#bbbbbb"
	col.UserColorName = "Updated"
	err = repo.Update(col)
	assert.NoError(t, err)

	fetched, err := repo.Get(col.ID)
	assert.NoError(t, err)
	assert.Equal(t, "#bbbbbb", fetched.Hex)
	assert.Equal(t, "Updated", fetched.UserColorName)
}

func TestDelete(t *testing.T) {
	repo, _ := setupTestDB(t)

	col := &models.Color{UserID: 7, Hex: "#deadbe"}
	err := repo.Create(col)
	assert.NoError(t, err)

	// delete
	err = repo.Delete(col.ID)
	assert.NoError(t, err)

	_, err = repo.Get(col.ID)
	assert.Error(t, err, "expected error fetching deleted record")
}
