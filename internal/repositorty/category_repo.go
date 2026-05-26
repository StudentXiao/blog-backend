package repositorty

import (
	"blog-backend/internal/database"
	"blog-backend/internal/models"
)

type CategoryRepository struct{}

func (r *CategoryRepository) Create(category *models.Category) error {
	return database.DB.Create(category).Error
}

func (r *CategoryRepository) FindByID(id uint) (*models.Category, error) {
	var category models.Category
	err := database.DB.Preload("Posts").First(&category, id).Error
	return &category, err
}

func (r *CategoryRepository) FindBySlug(slug string) (*models.Category, error) {
	var category models.Category
	err := database.DB.Where("slug = ?", slug).First(&category, slug).Error
	return &category, err
}

func (r *CategoryRepository) FindAll() ([]models.Category, error) {
	var categories []models.Category
	err := database.DB.Order("sort_order ASC").Find(&categories).Error
	return categories, err
}

func (r *CategoryRepository) Update(category *models.Category) error {
	return database.DB.Save(category).Error
}

func (r *CategoryRepository) Delete(id uint) error {
	return database.DB.Delete(&models.Category{}, id).Error
}
