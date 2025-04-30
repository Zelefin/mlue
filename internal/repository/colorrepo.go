package repository

import (
	"errors"
	"mlue/internal/models"

	"gorm.io/gorm"
)

var ErrColorNotFound = errors.New("color not found")

type ColorRepo interface {
	GetAll() ([]models.Color, error)
	GetByUser(userID uint) ([]models.Color, error)
	Get(id uint) (models.Color, error)
	Create(c *models.Color) error
	Update(c *models.Color) error
	Delete(id uint) error
}

type colorRepo struct{ db *gorm.DB }

func NewColorRepo(db *gorm.DB) ColorRepo {
	return &colorRepo{db: db}
}

func (r *colorRepo) GetAll() ([]models.Color, error) {
	var cs []models.Color
	return cs, r.db.Find(&cs).Error
}

func (r *colorRepo) GetByUser(userID uint) ([]models.Color, error) {
	var cs []models.Color
	return cs, r.db.Where("user_id = ?", userID).Find(&cs).Error
}

func (r *colorRepo) Get(id uint) (models.Color, error) {
	var c models.Color
	err := r.db.First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return c, ErrColorNotFound
	}
	return c, err
}

func (r *colorRepo) Create(c *models.Color) error {
	return r.db.Create(c).Error
}

func (r *colorRepo) Update(c *models.Color) error {
	return r.db.Save(c).Error
}

func (r *colorRepo) Delete(id uint) error {
	return r.db.Delete(&models.Color{}, id).Error
}
