package service

import (
	"blog-backend/internal/models"
	"blog-backend/internal/repositorty"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadService struct {
	repo *repositorty.UploadRepository
}

func NewUploadService() *UploadService {
	return &UploadService{
		repo: &repositorty.UploadRepository{},
	}
}

type UploadResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	MimeType string `json:"mime_type"`
	ID       uint   `json:"id"`
}

func (s *UploadService) UploadFile(file *multipart.FileHeader, userID uint) (*UploadResult, error) {
	// validate file type
	allowedTypes := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"application/pdf": true,
	}

	mimeType := file.Header.Get("Content-Type")
	if !allowedTypes[mimeType] {
		return nil, fmt.Errorf("file type not allowed")
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file too large (max 10MB)")
	}

	// Create upload directory if not exists
	uploadDir := "uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), userID, ext)
	destPath := filepath.Join(uploadDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	// Create upload record
	upload := &models.Upload{
		Filename:     filename,
		OriginalName: file.Filename,
		Path:         destPath,
		URL:          fmt.Sprintf("/uploads/%s", filename),
		Size:         file.Size,
		MimeType:     mimeType,
		UserID:       userID,
	}

	if err := s.repo.Create(upload); err != nil {
		os.Remove(destPath)
		return nil, err
	}

	return &UploadResult{
		URL:      upload.URL,
		Filename: filename,
		Size:     file.Size,
		MimeType: mimeType,
		ID:       upload.ID,
	}, nil
}

func (s *UploadService) GetUserUploads(userID uint, page, pageSize int) ([]models.Upload, int64, error) {
	return s.repo.FindByUserID(userID, page, pageSize)
}

func (s *UploadService) DeleteUpload(id uint, userID uint, isAdmin bool) error {
	upload, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}

	// Check permission
	if upload.UserID != userID && !isAdmin {
		return fmt.Errorf("permission denied")
	}

	// Delete file
	if err := os.Remove(upload.Path); err != nil && !os.IsNotExist(err) {
		return err
	}

	return s.repo.Delete(id)
}

// ServerStatic serves static files
func ServerStatic(r *gin.Engine) {
	r.Static("/uploads", "./uploads")
	r.Static("/static", "./static")
}
