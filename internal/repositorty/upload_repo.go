package repositorty

import (
	"blog-backend/internal/database"
	"blog-backend/internal/models"
)

type UploadRepository struct{}

func (r *UploadRepository) Create(upload *models.Upload) error {
	return database.DB.Create(upload).Error
}

func (r *UploadRepository) FindByID(id uint) (*models.Upload, error) {
	var upload models.Upload
	err := database.DB.First(&upload, id).Error
	return &upload, err
}

func (r *UploadRepository) FindByUserID(userID uint, page, pageSize int) ([]models.Upload, int64, error) {
	var uploads []models.Upload
	var total int64

	query := database.DB.Model(&models.Upload{}).Where("user_id = ", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&uploads).Error

	return uploads, total, err
}

func (r *UploadRepository) Delete(id uint) error {
	return database.DB.Delete(&models.Upload{}, id).Error
}
