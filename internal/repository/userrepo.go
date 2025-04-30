package repository

import (
	"errors"
	"mlue/internal/models"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepo interface {
	Get(id uint) (models.User, error)
	GetCountUserColors(id uint) int64
	FindByGoogleSub(sub string) (models.User, error)
	Create(u *models.User) error
	Update(u *models.User) error
}

type userRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) Get(id uint) (models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, ErrUserNotFound
	}
	return user, err
}

func (r *userRepo) GetCountUserColors(id uint) (count int64) {
	err := r.db.Model(&models.Color{}).Where("user_id = ?", id).Count(&count).Error
	if err != nil {
		count = 0
	}
	return
}

func (r *userRepo) FindByGoogleSub(sub string) (models.User, error) {
	var user models.User
	err := r.db.Where("google_sub = ?", sub).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, ErrUserNotFound
	}
	return user, err
}

func (r *userRepo) Create(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *userRepo) Update(u *models.User) error {
	return r.db.Save(u).Error
}
