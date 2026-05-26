package repositorty

import (
	"blog-backend/internal/database"
	"blog-backend/internal/models"

	"gorm.io/gorm"
)

type PostRepository struct{}

func (r *PostRepository) Create(post *models.Post) error {
	return database.DB.Create(post).Error
}

func (r *PostRepository) FindByID(id uint) (*models.Post, error) {
	var post models.Post
	err := database.DB.Preload("User").Preload("Comments.User").First(&post, id).Error
	return &post, err
}

func (r *PostRepository) FindBySlug(slug string) (*models.Post, error) {
	var post models.Post
	err := database.DB.Preload("User").Where("slug = ?", slug).First(&post).Error
	return &post, err
}

func (r *PostRepository) FindAll(page, pageSize int, status string, categoryID *uint, tagID *uint) ([]models.Post, int64, error) {
	var posts []models.Post
	var total int64

	query := database.DB.Model(&models.Post{}).Preload("User").Preload("Category").Preload("Tags")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if categoryID != nil && *categoryID > 0 {
		query = query.Where("category_id = ?", categoryID)
	}

	if tagID != nil && *tagID > 0 {
		query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").Where("post_tags.tag_id = ?", tagID)
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error

	return posts, total, err
}

func (r *PostRepository) Update(post *models.Post) error {
	return database.DB.Save(post).Error
}

func (r *PostRepository) Delete(id uint) error {
	return database.DB.Delete(&models.Post{}, id).Error
}

func (r *PostRepository) IncrementViews(id uint) error {
	return database.DB.Model(&models.Post{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
