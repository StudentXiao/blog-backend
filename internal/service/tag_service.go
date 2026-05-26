package service

import (
	"blog-backend/internal/models"
	"blog-backend/internal/repositorty"
	"strings"
)

type TagService struct {
	repo *repositorty.TagRepository
}

func NewTagService() *TagService {
	return &TagService{
		repo: &repositorty.TagRepository{},
	}
}

func (s *TagService) Create(tag *models.Tag) error {
	if tag.Slug == "" {
		tag.Slug = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(tag.Name, " ", "-"), "?", ""))
	}
	return s.repo.Cteate(tag)
}

func (s *TagService) GetByID(id uint) (*models.Tag, error) {
	return s.repo.FindByID(id)
}

func (s *TagService) GetBySlug(slug string) (*models.Tag, error) {
	return s.repo.FindBySlug(slug)
}

func (s *TagService) GetAll() ([]models.Tag, error) {
	return s.repo.FindAll()
}

func (s *TagService) Update(tag *models.Tag) error {
	return s.repo.Update(tag)
}

func (s *TagService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *TagService) GetORCreate(name, slug string) (*models.Tag, error) {
	return s.repo.GetORCreate(name, slug)
}

func (s *TagService) ProcessTags(tagNames []string) ([]models.Tag, error) {
	var tags []models.Tag
	for _, tagName := range tagNames {
		tagName = strings.TrimSpace(tagName)
		if tagName == "" {
			continue
		}
		slug := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(tagName, " ", "-"), "?", ""))
		tag, err := s.repo.GetORCreate(tagName, slug)
		if err != nil {
			return nil, err
		}
		tags = append(tags, *tag)
	}
	return tags, nil
}
