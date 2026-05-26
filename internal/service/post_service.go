package service

import (
	"blog-backend/internal/models"
	"blog-backend/internal/repositorty"
	"strings"
)

type PostService struct {
	postRepo   *repositorty.PostRepository
	tagService *TagService
}

func NewPostService() *PostService {
	return &PostService{
		postRepo:   &repositorty.PostRepository{},
		tagService: NewTagService(),
	}
}

func (s *PostService) Create(post *models.Post) error {
	// Generate slug from title if not provided
	if post.Slug == "" {
		post.Slug = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(post.Title, " ", "-"), "?", ""))
	}
	return s.postRepo.Create(post)
}

func (s *PostService) GetByID(id uint) (*models.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *PostService) GetBySlug(slug string) (*models.Post, error) {
	post, err := s.postRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	// Increment views
	go s.postRepo.IncrementViews(post.ID)
	return post, nil
}

func (s *PostService) List(page, pageSize int, status string) ([]models.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.postRepo.FindAll(page, pageSize, status)
}

func (s *PostService) Update(post *models.Post) error {
	return s.postRepo.Update(post)
}

func (s *PostService) Delete(id uint) error {
	return s.postRepo.Delete(id)
}
