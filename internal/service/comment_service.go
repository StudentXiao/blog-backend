package service

import (
	"blog-backend/internal/models"
	"blog-backend/internal/repositorty"
)

type CommentService struct {
	repo *repositorty.CommentRepository
}

func NewCommentService() *CommentService {
	return &CommentService{
		repo: &repositorty.CommentRepository{},
	}
}

func (s *CommentService) Create(comment *models.Comment) error {
	if comment.Status == "" {
		comment.Status = "pending"
	}
	return s.repo.Create(comment)
}

func (s *CommentService) GetByID(id uint) (*models.Comment, error) {
	return s.repo.FindByID(id)
}

func (s *CommentService) GetByPostID(postID uint, page, pageSize int) ([]models.Comment, int64, error) {
	return s.repo.FindByPostID(postID, page, pageSize, "approved")
}

func (s *CommentService) GetPendingComments(page, pageSize int) ([]models.Comment, int64, error) {
	return s.repo.FindByPostID(0, page, pageSize, "pending")
}

func (s *CommentService) ApproveComment(id uint) error {
	comment, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	comment.Status = "approved"
	return s.repo.Update(comment)
}

func (s *CommentService) RejectComment(id uint) error {
	comment, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	comment.Status = "spam"
	return s.repo.Update(comment)
}

func (s *CommentService) DeleteComment(id uint) error {
	return s.repo.Delete(id)
}

func (s *CommentService) GetPendingCount() (int64, error) {
	return s.repo.GetPendingCount()
}
