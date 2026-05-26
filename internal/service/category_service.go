package service

import (
	"blog-backend/internal/models"
	"blog-backend/internal/repositorty"
	"strings"
)

type CategoryService struct {
	repo *repositorty.CategoryRepository
}

func NewCategoryService() *CategoryService {
	return &CategoryService{
		repo: &repositorty.CategoryRepository{},
	}
}

func (s *CategoryService) Create(category *models.Category) error {
	if category.Slug == "" {
		category.Slug = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(category.Name, " ", "-"), "?", ""))
	}
	return s.repo.Create(category)
}

func (s *CategoryService) GetByID(id uint) (*models.Category, error) {
	return s.repo.FindByID(id)
}

func (s *CategoryService) GetBySlug(slug string) (*models.Category, error) {
	return s.repo.FindBySlug(slug)
}

func (s *CategoryService) GetAll() ([]models.Category, error) {
	return s.repo.FindAll()
}

func (s *CategoryService) Update(category *models.Category) error {
	return s.repo.Update(category)
}

func (s *CategoryService) Delete(id uint) error {
	return s.repo.Delete(id)
}
