package repositorty

import (
	"blog-backend/internal/database"
	"blog-backend/internal/models"
)

type TagRepository struct{}

func (r *TagRepository) Cteate(tag *models.Tag) error {
	return database.DB.Create(tag).Error
}

func (r *TagRepository) FindByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	err := database.DB.First(&tag, id).Error
	return &tag, err
}

func (r *TagRepository) FindBySlug(slug string) (*models.Tag, error) {
	var tag models.Tag
	err := database.DB.Where("slug = ?", slug).First(&tag).Error
	return &tag, err
}

func (r *TagRepository) FindAll() ([]models.Tag, error) {
	var tags []models.Tag
	err := database.DB.Find(&tags).Error
	return tags, err
}

func (r *TagRepository) Update(tag *models.Tag) error {
	return database.DB.Save(tag).Error
}

func (r *TagRepository) Delete(id uint) error {
	return database.DB.Delete(&models.Tag{}, id).Error
}

func (r *TagRepository) GetORCreate(name, slug string) (*models.Tag, error) {
	var tag models.Tag
	err := database.DB.Where("slug = ?", slug).First(&tag).Error
	if err == nil {
		return &tag, err
	}

	tag = models.Tag{
		Name: name,
		Slug: slug,
	}
	err = database.DB.Create(&tag).Error
	return &tag, err
}
