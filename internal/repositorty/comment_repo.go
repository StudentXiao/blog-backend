package repositorty

import (
	"blog-backend/internal/database"
	"blog-backend/internal/models"
)

type CommentRepository struct{}

func (r *CommentRepository) Create(comment *models.Comment) error {
	return database.DB.Create(comment).Error
}

func (r *CommentRepository) FindByID(id uint) (*models.Comment, error) {
	var comment models.Comment
	err := database.DB.Preload("User").Preload("Replies.User").First(&comment, id).Error
	return &comment, err
}

func (r *CommentRepository) FindByPostID(postID uint, page, pageSize int, status string) ([]models.Comment, int64, error) {
	var comments []models.Comment
	var total int64

	query := database.DB.Model(&models.Comment{}).Where("post_id = ? AND parents IS NULL", postID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("User").Preload("Replies.User").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&comments).Error

	return comments, total, err
}

func (r *CommentRepository) Update(comment *models.Comment) error {
	return database.DB.Save(comment).Error
}

func (r *CommentRepository) Delete(id uint) error {
	return database.DB.Delete(&models.Comment{}, id).Error
}

func (r *CommentRepository) GetPendingCount() (int64, error) {
	var count int64
	err := database.DB.Model(&models.Comment{}).Where("status = ?", "pending").Count(&count).Error
	return count, err
}
