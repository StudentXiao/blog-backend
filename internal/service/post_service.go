package service

import (
	"blog-backend/internal/database"
	"blog-backend/internal/models"
	"blog-backend/internal/repositorty"
	"errors"
	"strings"
	"time"
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

func (s *PostService) Create(post *models.Post, tagNames []string) error {
	// 1. 如果提供了 CategoryID，验证其是否存在
	if post.CategoryID != nil && *post.CategoryID > 0 {
		var count int64
		database.DB.Model(&models.Category{}).Where("id = ?", *post.CategoryID).Count(&count)
		if count == 0 {
			return errors.New("category id not found")
		}
	}

	// Generate slug from title if not provided
	if post.Slug == "" {
		post.Slug = generateSlug(post.Title) // strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(post.Title, " ", "-"), "?", ""))
	}

	// Set published time if status is published
	if post.Status == "published" && post.PublicshedAt == nil {
		now := time.Now()
		post.PublicshedAt = &now
	}

	// Process tags
	if len(tagNames) > 0 {
		tags, err := s.tagService.ProcessTags(tagNames)
		if err != nil {
			return err
		}
		post.Tags = tags
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

func (s *PostService) List(page, pageSize int, status string, categoryID *uint, tagID *uint) ([]models.Post, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.postRepo.FindAll(page, pageSize, status, categoryID, tagID)
}

func (s *PostService) Update(post *models.Post, tagNames []string) error {
	// Update published time if status changed to published
	if post.Status == "published" && post.PublicshedAt == nil {
		now := time.Now()
		post.PublicshedAt = &now
	}

	// Process tags
	if len(tagNames) > 0 {
		tags, err := s.tagService.ProcessTags(tagNames)
		if err != nil {
			return err
		}
		// Clear existion tags and add ones
		if err := database.DB.Model(post).Association("Tags").Clear(); err != nil {
			return err
		}
		post.Tags = tags
	}

	return s.postRepo.Update(post)
}

func (s *PostService) Delete(id uint) error {
	return s.postRepo.Delete(id)
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "?", "")
	slug = strings.ReplaceAll(slug, "/", "-")
	slug = strings.ReplaceAll(slug, "\\", "-")
	return slug
}
