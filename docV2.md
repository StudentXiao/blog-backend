我来为这个博客项目添加更多功能，并说明管理员账户的创建方法。

## 扩展的功能模块

### 1. 更新后的数据模型 (internal/models/models.go)

```go
package models

import (
    "time"
    "gorm.io/gorm"
)

type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Username  string         `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email     string         `gorm:"uniqueIndex;size:100;not null" json:"email"`
    Password  string         `gorm:"not null" json:"-"`
    Avatar    string         `json:"avatar"`
    Bio       string         `gorm:"size:500" json:"bio"`
    Role      string         `gorm:"default:user" json:"role"` // user, admin, editor
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    Posts     []Post         `json:"posts,omitempty"`
    Comments  []Comment      `json:"comments,omitempty"`
}

type Category struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    Name        string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
    Slug        string         `gorm:"uniqueIndex;size:50;not null" json:"slug"`
    Description string         `gorm:"size:200" json:"description"`
    ParentID    *uint          `json:"parent_id"`
    SortOrder   int            `gorm:"default:0" json:"sort_order"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
    Posts       []Post         `gorm:"many2many:post_categories;" json:"posts,omitempty"`
}

type Tag struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Name      string         `gorm:"uniqueIndex;size:50;not null" json:"name"`
    Slug      string         `gorm:"uniqueIndex;size:50;not null" json:"slug"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    Posts     []Post         `gorm:"many2many:post_tags;" json:"posts,omitempty"`
}

type Post struct {
    ID          uint           `gorm:"primaryKey" json:"id"`
    Title       string         `gorm:"size:200;not null" json:"title"`
    Slug        string         `gorm:"uniqueIndex;size:200;not null" json:"slug"`
    Content     string         `gorm:"type:longtext;not null" json:"content"`
    Summary     string         `gorm:"size:500" json:"summary"`
    Cover       string         `json:"cover"`
    Views       int            `gorm:"default:0" json:"views"`
    Likes       int            `gorm:"default:0" json:"likes"`
    Status      string         `gorm:"default:draft" json:"status"` // draft, published, private
    UserID      uint           `gorm:"not null" json:"user_id"`
    User        User           `json:"user,omitempty"`
    CategoryID  *uint          `json:"category_id"`
    Category    *Category      `json:"category,omitempty"`
    Tags        []Tag          `gorm:"many2many:post_tags;" json:"tags,omitempty"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    PublishedAt *time.Time     `json:"published_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
    Comments    []Comment      `json:"comments,omitempty"`
}

type Comment struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Content   string         `gorm:"type:text;not null" json:"content"`
    UserID    uint           `gorm:"not null" json:"user_id"`
    User      User           `json:"user,omitempty"`
    PostID    uint           `gorm:"not null;index" json:"post_id"`
    Post      Post           `json:"-"`
    ParentID  *uint          `json:"parent_id"`
    Status    string         `gorm:"default:pending" json:"status"` // pending, approved, spam, deleted
    IPAddress string         `json:"ip_address"`
    UserAgent string         `json:"user_agent"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    Replies   []Comment      `gorm:"foreignkey:ParentID" json:"replies,omitempty"`
}

type Upload struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Filename  string         `gorm:"size:255;not null" json:"filename"`
    OriginalName string      `gorm:"size:255;not null" json:"original_name"`
    Path      string         `gorm:"size:500;not null" json:"path"`
    URL       string         `gorm:"size:500;not null" json:"url"`
    Size      int64          `json:"size"`
    MimeType  string         `gorm:"size:100" json:"mime_type"`
    UserID    uint           `gorm:"not null" json:"user_id"`
    User      User           `json:"user,omitempty"`
    CreatedAt time.Time      `json:"created_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 2. 新增 Repository 层

#### internal/repository/category_repo.go
```go
package repository

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
    err := database.DB.Where("slug = ?", slug).First(&category).Error
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
```

#### internal/repository/tag_repo.go
```go
package repository

import (
    "blog-backend/internal/database"
    "blog-backend/internal/models"
)

type TagRepository struct{}

func (r *TagRepository) Create(tag *models.Tag) error {
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

func (r *TagRepository) GetOrCreate(name, slug string) (*models.Tag, error) {
    var tag models.Tag
    err := database.DB.Where("slug = ?", slug).First(&tag).Error
    if err == nil {
        return &tag, nil
    }
    
    tag = models.Tag{
        Name: name,
        Slug: slug,
    }
    err = database.DB.Create(&tag).Error
    return &tag, err
}
```

#### internal/repository/comment_repo.go
```go
package repository

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
    
    query := database.DB.Model(&models.Comment{}).Where("post_id = ? AND parent_id IS NULL", postID)
    if status != "" {
        query = query.Where("status = ?", status)
    }
    
    query.Count(&total)
    
    offset := (page - 1) * pageSize
    err := query.Preload("User").Preload("Replies.User").
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&comments).Error
    
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
```

#### internal/repository/upload_repo.go
```go
package repository

import (
    "blog-backend/internal/database"
    "blog-backend/internal/models"
)

type UploadRepository struct{}

func (r *UploadRepository) Create(upload *models.Upload) error {
    return database.DB.Create(upload).Error
}

func (r *UploadRepository) FindByID(id uint) (*models.Upload, error) {
    var upload models.Upload
    err := database.DB.First(&upload, id).Error
    return &upload, err
}

func (r *UploadRepository) FindByUserID(userID uint, page, pageSize int) ([]models.Upload, int64, error) {
    var uploads []models.Upload
    var total int64
    
    query := database.DB.Model(&models.Upload{}).Where("user_id = ?", userID)
    query.Count(&total)
    
    offset := (page - 1) * pageSize
    err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&uploads).Error
    
    return uploads, total, err
}

func (r *UploadRepository) Delete(id uint) error {
    return database.DB.Delete(&models.Upload{}, id).Error
}
```

### 3. 新增 Service 层

#### internal/service/category_service.go
```go
package service

import (
    "blog-backend/internal/models"
    "blog-backend/internal/repository"
    "strings"
)

type CategoryService struct {
    repo *repository.CategoryRepository
}

func NewCategoryService() *CategoryService {
    return &CategoryService{
        repo: &repository.CategoryRepository{},
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
```

#### internal/service/tag_service.go
```go
package service

import (
    "blog-backend/internal/models"
    "blog-backend/internal/repository"
    "strings"
)

type TagService struct {
    repo *repository.TagRepository
}

func NewTagService() *TagService {
    return &TagService{
        repo: &repository.TagRepository{},
    }
}

func (s *TagService) Create(tag *models.Tag) error {
    if tag.Slug == "" {
        tag.Slug = strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(tag.Name, " ", "-"), "?", ""))
    }
    return s.repo.Create(tag)
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

func (s *TagService) GetOrCreate(name, slug string) (*models.Tag, error) {
    return s.repo.GetOrCreate(name, slug)
}

func (s *TagService) ProcessTags(tagNames []string) ([]models.Tag, error) {
    var tags []models.Tag
    for _, tagName := range tagNames {
        tagName = strings.TrimSpace(tagName)
        if tagName == "" {
            continue
        }
        slug := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(tagName, " ", "-"), "?", ""))
        tag, err := s.repo.GetOrCreate(tagName, slug)
        if err != nil {
            return nil, err
        }
        tags = append(tags, *tag)
    }
    return tags, nil
}
```

#### internal/service/comment_service.go
```go
package service

import (
    "blog-backend/internal/models"
    "blog-backend/internal/repository"
    "errors"
)

type CommentService struct {
    repo *repository.CommentRepository
}

func NewCommentService() *CommentService {
    return &CommentService{
        repo: &repository.CommentRepository{},
    }
}

func (s *CommentService) Create(comment *models.Comment) error {
    if comment.Status == "" {
        comment.Status = "pending" // Auto-moderate
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
```

#### internal/service/upload_service.go
```go
package service

import (
    "blog-backend/internal/config"
    "blog-backend/internal/models"
    "blog-backend/internal/repository"
    "fmt"
    "io"
    "mime/multipart"
    "os"
    "path/filepath"
    "strings"
    "time"
)

type UploadService struct {
    repo *repository.UploadRepository
}

func NewUploadService() *UploadService {
    return &UploadService{
        repo: &repository.UploadRepository{},
    }
}

type UploadResult struct {
    URL       string `json:"url"`
    Filename  string `json:"filename"`
    Size      int64  `json:"size"`
    MimeType  string `json:"mime_type"`
    ID        uint   `json:"id"`
}

func (s *UploadService) UploadFile(file *multipart.FileHeader, userID uint) (*UploadResult, error) {
    // Validate file type
    allowedTypes := map[string]bool{
        "image/jpeg": true,
        "image/png":  true,
        "image/gif":  true,
        "image/webp": true,
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
    filepath := filepath.Join(uploadDir, filename)
    
    // Save file
    src, err := file.Open()
    if err != nil {
        return nil, err
    }
    defer src.Close()
    
    dst, err := os.Create(filepath)
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
        Path:         filepath,
        URL:          fmt.Sprintf("/uploads/%s", filename),
        Size:         file.Size,
        MimeType:     mimeType,
        UserID:       userID,
    }
    
    if err := s.repo.Create(upload); err != nil {
        os.Remove(filepath)
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

// ServeStatic serves static files
func ServeStatic(r *gin.Engine) {
    r.Static("/uploads", "./uploads")
    r.Static("/static", "./static")
}
```

### 4. 更新 PostService 以支持分类和标签

#### internal/service/post_service.go (更新)
```go
package service

import (
    "blog-backend/internal/database"
    "blog-backend/internal/models"
    "blog-backend/internal/repository"
    "errors"
    "time"
)

type PostService struct {
    postRepo *repository.PostRepository
    tagService *TagService
}

func NewPostService() *PostService {
    return &PostService{
        postRepo: &repository.PostRepository{},
        tagService: NewTagService(),
    }
}

func (s *PostService) Create(post *models.Post, tagNames []string) error {
    // Generate slug
    if post.Slug == "" {
        post.Slug = generateSlug(post.Title)
    }
    
    // Set published time if status is published
    if post.Status == "published" && post.PublishedAt == nil {
        now := time.Now()
        post.PublishedAt = &now
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

func (s *PostService) Update(post *models.Post, tagNames []string) error {
    // Update published time if status changed to published
    if post.Status == "published" && post.PublishedAt == nil {
        now := time.Now()
        post.PublishedAt = &now
    }
    
    // Process tags
    if tagNames != nil {
        tags, err := s.tagService.ProcessTags(tagNames)
        if err != nil {
            return err
        }
        // Clear existing tags and add new ones
        if err := database.DB.Model(post).Association("Tags").Clear(); err != nil {
            return err
        }
        post.Tags = tags
    }
    
    return s.postRepo.Update(post)
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

func generateSlug(title string) string {
    slug := strings.ToLower(title)
    slug = strings.ReplaceAll(slug, " ", "-")
    slug = strings.ReplaceAll(slug, "?", "")
    slug = strings.ReplaceAll(slug, "/", "-")
    slug = strings.ReplaceAll(slug, "\\", "-")
    return slug
}
```

### 5. 更新 PostRepository

#### internal/repository/post_repo.go (添加新方法)
```go
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
        query = query.Joins("JOIN post_tags ON post_tags.post_id = posts.id").
            Where("post_tags.tag_id = ?", tagID)
    }
    
    query.Count(&total)
    
    offset := (page - 1) * pageSize
    err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error
    
    return posts, total, err
}
```

### 6. 新增 Handler

#### internal/handler/category_handler.go
```go
package handler

import (
    "blog-backend/internal/models"
    "blog-backend/internal/service"
    "strconv"
    "github.com/gin-gonic/gin"
)

type CategoryHandler struct {
    service *service.CategoryService
}

func NewCategoryHandler() *CategoryHandler {
    return &CategoryHandler{
        service: service.NewCategoryService(),
    }
}

type CreateCategoryRequest struct {
    Name        string `json:"name" binding:"required"`
    Slug        string `json:"slug"`
    Description string `json:"description"`
    ParentID    *uint  `json:"parent_id"`
    SortOrder   int    `json:"sort_order"`
}

func (h *CategoryHandler) Create(c *gin.Context) {
    var req CreateCategoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    category := &models.Category{
        Name:        req.Name,
        Slug:        req.Slug,
        Description: req.Description,
        ParentID:    req.ParentID,
        SortOrder:   req.SortOrder,
    }
    
    if err := h.service.Create(category); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, category)
}

func (h *CategoryHandler) GetByID(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    category, err := h.service.GetByID(uint(id))
    if err != nil {
        NotFound(c, "Category not found")
        return
    }
    
    Success(c, category)
}

func (h *CategoryHandler) GetAll(c *gin.Context) {
    categories, err := h.service.GetAll()
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, categories)
}

func (h *CategoryHandler) Update(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    category, err := h.service.GetByID(uint(id))
    if err != nil {
        NotFound(c, "Category not found")
        return
    }
    
    var req CreateCategoryRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    if req.Name != "" {
        category.Name = req.Name
    }
    if req.Slug != "" {
        category.Slug = req.Slug
    }
    if req.Description != "" {
        category.Description = req.Description
    }
    if req.ParentID != nil {
        category.ParentID = req.ParentID
    }
    category.SortOrder = req.SortOrder
    
    if err := h.service.Update(category); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, category)
}

func (h *CategoryHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    if err := h.service.Delete(uint(id)); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    SuccessWithMessage(c, "Category deleted successfully", nil)
}
```

#### internal/handler/tag_handler.go
```go
package handler

import (
    "blog-backend/internal/models"
    "blog-backend/internal/service"
    "strconv"
    "github.com/gin-gonic/gin"
)

type TagHandler struct {
    service *service.TagService
}

func NewTagHandler() *TagHandler {
    return &TagHandler{
        service: service.NewTagService(),
    }
}

type CreateTagRequest struct {
    Name string `json:"name" binding:"required"`
    Slug string `json:"slug"`
}

func (h *TagHandler) Create(c *gin.Context) {
    var req CreateTagRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    tag := &models.Tag{
        Name: req.Name,
        Slug: req.Slug,
    }
    
    if err := h.service.Create(tag); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, tag)
}

func (h *TagHandler) GetAll(c *gin.Context) {
    tags, err := h.service.GetAll()
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, tags)
}

func (h *TagHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    if err := h.service.Delete(uint(id)); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    SuccessWithMessage(c, "Tag deleted successfully", nil)
}
```

#### internal/handler/comment_handler.go
```go
package handler

import (
    "blog-backend/internal/models"
    "blog-backend/internal/service"
    "strconv"
    "github.com/gin-gonic/gin"
)

type CommentHandler struct {
    service *service.CommentService
}

func NewCommentHandler() *CommentHandler {
    return &CommentHandler{
        service: service.NewCommentService(),
    }
}

type CreateCommentRequest struct {
    Content  string `json:"content" binding:"required"`
    PostID   uint   `json:"post_id" binding:"required"`
    ParentID *uint  `json:"parent_id"`
}

func (h *CommentHandler) Create(c *gin.Context) {
    var req CreateCommentRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    userID, exists := c.Get("user_id")
    if !exists {
        Unauthorized(c, "Please login to comment")
        return
    }
    
    comment := &models.Comment{
        Content:   req.Content,
        UserID:    userID.(uint),
        PostID:    req.PostID,
        ParentID:  req.ParentID,
        IPAddress: c.ClientIP(),
        UserAgent: c.GetHeader("User-Agent"),
    }
    
    if err := h.service.Create(comment); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, comment)
}

func (h *CommentHandler) GetByPost(c *gin.Context) {
    postID, err := strconv.ParseUint(c.Param("post_id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid post ID")
        return
    }
    
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    comments, total, err := h.service.GetByPostID(uint(postID), page, pageSize)
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "comments":  comments,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}

// Admin only endpoints
func (h *CommentHandler) GetPending(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    comments, total, err := h.service.GetPendingComments(page, pageSize)
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "comments":  comments,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}

func (h *CommentHandler) Approve(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    if err := h.service.ApproveComment(uint(id)); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    SuccessWithMessage(c, "Comment approved", nil)
}

func (h *CommentHandler) Reject(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    if err := h.service.RejectComment(uint(id)); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    SuccessWithMessage(c, "Comment rejected", nil)
}
```

#### internal/handler/upload_handler.go
```go
package handler

import (
    "blog-backend/internal/service"
    "strconv"
    "github.com/gin-gonic/gin"
)

type UploadHandler struct {
    service *service.UploadService
}

func NewUploadHandler() *UploadHandler {
    return &UploadHandler{
        service: service.NewUploadService(),
    }
}

func (h *UploadHandler) Upload(c *gin.Context) {
    file, err := c.FormFile("file")
    if err != nil {
        BadRequest(c, "No file uploaded")
        return
    }
    
    userID, _ := c.Get("user_id")
    
    result, err := h.service.UploadFile(file, userID.(uint))
    if err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    Success(c, result)
}

func (h *UploadHandler) GetUserUploads(c *gin.Context) {
    userID, _ := c.Get("user_id")
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    uploads, total, err := h.service.GetUserUploads(userID.(uint), page, pageSize)
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "uploads":   uploads,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}

func (h *UploadHandler) Delete(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    userID, _ := c.Get("user_id")
    role, _ := c.Get("user_role")
    isAdmin := role == "admin"
    
    if err := h.service.DeleteUpload(uint(id), userID.(uint), isAdmin); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    SuccessWithMessage(c, "File deleted successfully", nil)
}
```

### 7. 创建管理员账户的方法

有几种方式可以创建管理员账户：

#### 方法1: 创建一个命令行脚本 (cmd/create_admin/main.go)

```go
package main

import (
    "blog-backend/internal/config"
    "blog-backend/internal/database"
    "blog-backend/internal/models"
    "fmt"
    "log"
    "os"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    // Load config
    config.LoadConfig()
    
    // Initialize database
    database.InitDB()
    
    if len(os.Args) < 4 {
        fmt.Println("Usage: go run cmd/create_admin/main.go <username> <email> <password>")
        os.Exit(1)
    }
    
    username := os.Args[1]
    email := os.Args[2]
    password := os.Args[3]
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal("Failed to hash password:", err)
    }
    
    // Create admin user
    admin := &models.User{
        Username: username,
        Email:    email,
        Password: string(hashedPassword),
        Role:     "admin",
    }
    
    result := database.DB.Create(admin)
    if result.Error != nil {
        log.Fatal("Failed to create admin:", result.Error)
    }
    
    fmt.Printf("Admin user created successfully!\nID: %d\nUsername: %s\nEmail: %s\n", admin.ID, admin.Username, admin.Email)
}
```

使用方法：
```bash
go run cmd/create_admin/main.go admin admin@example.com mypassword123
```

#### 方法2: 创建一个 API 端点（仅开发环境使用）

在 `internal/handler/auth_handler.go` 中添加：

```go
// CreateAdmin (development only - should be protected in production)
func (h *AuthHandler) CreateAdmin(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    // Check if admin already exists
    existingUser, _ := h.authService.userRepo.FindByEmail(req.Email)
    if existingUser != nil {
        BadRequest(c, "User already exists")
        return
    }
    
    hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    
    user := &models.User{
        Username: req.Username,
        Email:    req.Email,
        Password: string(hashedPassword),
        Role:     "admin",
    }
    
    if err := h.authService.userRepo.Create(user); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, user)
}
```

在路由中添加（仅开发环境）：
```go
// Development only - remove in production
if gin.Mode() == gin.DebugMode {
    api.POST("/admin/create", authHandler.CreateAdmin)
}
```

#### 方法3: 直接通过 SQL 创建

```sql
-- 使用 bcrypt 加密密码 'admin123' 的结果
INSERT INTO users (username, email, password, role, created_at, updated_at) 
VALUES ('admin', 'admin@example.com', '$2a$10$YourHashedPasswordHere', 'admin', NOW(), NOW());
```

### 8. 更新路由配置

#### internal/router/router.go (完整更新)

```go
package router

import (
    "blog-backend/internal/handler"
    "blog-backend/internal/middleware"
    "blog-backend/internal/service"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    router := gin.Default()
    
    // Global middleware
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())
    
    // Serve static files
    service.ServeStatic(router)
    
    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // Public routes
    api := router.Group("/api/v1")
    {
        authHandler := handler.NewAuthHandler()
        postHandler := handler.NewPostHandler()
        categoryHandler := handler.NewCategoryHandler()
        tagHandler := handler.NewTagHandler()
        commentHandler := handler.NewCommentHandler()
        
        // Auth endpoints
        api.POST("/auth/register", authHandler.Register)
        api.POST("/auth/login", authHandler.Login)
        
        // Public post endpoints
        api.GET("/posts", postHandler.List)
        api.GET("/posts/slug/:slug", postHandler.GetBySlug)
        api.GET("/posts/:id", postHandler.GetByID)
        
        // Public category endpoints
        api.GET("/categories", categoryHandler.GetAll)
        api.GET("/categories/:id", categoryHandler.GetByID)
        
        // Public tag endpoints
        api.GET("/tags", tagHandler.GetAll)
        
        // Public comment endpoints
        api.GET("/posts/:post_id/comments", commentHandler.GetByPost)
        
        // Protected routes
        protected := api.Group("/")
        protected.Use(middleware.Auth())
        {
            // Post management
            protected.POST("/posts", postHandler.Create)
            protected.PUT("/posts/:id", postHandler.Update)
            protected.DELETE("/posts/:id", postHandler.Delete)
            
            // Comment management
            protected.POST("/comments", commentHandler.Create)
            
            // Upload management
            uploadHandler := handler.NewUploadHandler()
            protected.POST("/upload", uploadHandler.Upload)
            protected.GET("/uploads", uploadHandler.GetUserUploads)
            protected.DELETE("/upload/:id", uploadHandler.Delete)
        }
        
        // Admin only routes
        admin := api.Group("/admin")
        admin.Use(middleware.Auth(), middleware.AdminOnly())
        {
            // Post management
            admin.GET("/posts/all", postHandler.List) // List all posts including drafts
            
            // Category management
            admin.POST("/categories", categoryHandler.Create)
            admin.PUT("/categories/:id", categoryHandler.Update)
            admin.DELETE("/categories/:id", categoryHandler.Delete)
            
            // Tag management
            admin.POST("/tags", tagHandler.Create)
            admin.DELETE("/tags/:id", tagHandler.Delete)
            
            // Comment moderation
            admin.GET("/comments/pending", commentHandler.GetPending)
            admin.POST("/comments/:id/approve", commentHandler.Approve)
            admin.POST("/comments/:id/reject", commentHandler.Reject)
            
            // User management (can add later)
            // admin.GET("/users", userHandler.List)
            // admin.PUT("/users/:id/role", userHandler.UpdateRole)
        }
    }
    
    return router
}
```

### 9. 更新 PostHandler

#### internal/handler/post_handler.go (更新)

```go
package handler

import (
    "blog-backend/internal/models"
    "blog-backend/internal/service"
    "strconv"
    "github.com/gin-gonic/gin"
)

type PostHandler struct {
    postService *service.PostService
}

func NewPostHandler() *PostHandler {
    return &PostHandler{
        postService: service.NewPostService(),
    }
}

type CreatePostRequest struct {
    Title      string   `json:"title" binding:"required"`
    Slug       string   `json:"slug"`
    Content    string   `json:"content" binding:"required"`
    Summary    string   `json:"summary"`
    Cover      string   `json:"cover"`
    Tags       []string `json:"tags"`
    CategoryID *uint    `json:"category_id"`
    Status     string   `json:"status"`
}

func (h *PostHandler) Create(c *gin.Context) {
    var req CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    userID, _ := c.Get("user_id")
    post := &models.Post{
        Title:      req.Title,
        Slug:       req.Slug,
        Content:    req.Content,
        Summary:    req.Summary,
        Cover:      req.Cover,
        CategoryID: req.CategoryID,
        Status:     req.Status,
        UserID:     userID.(uint),
    }
    
    if post.Status == "" {
        post.Status = "draft"
    }
    
    if err := h.postService.Create(post, req.Tags); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, post)
}

func (h *PostHandler) List(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
    status := c.Query("status")
    
    var categoryID *uint
    if catID := c.Query("category_id"); catID != "" {
        id, _ := strconv.ParseUint(catID, 10, 32)
        categoryID = new(uint)
        *categoryID = uint(id)
    }
    
    var tagID *uint
    if tID := c.Query("tag_id"); tID != "" {
        id, _ := strconv.ParseUint(tID, 10, 32)
        tagID = new(uint)
        *tagID = uint(id)
    }
    
    // Only show published posts for non-admin
    role, hasRole := c.Get("user_role")
    if !hasRole || role != "admin" {
        status = "published"
    }
    
    posts, total, err := h.postService.List(page, pageSize, status, categoryID, tagID)
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "posts":     posts,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}

func (h *PostHandler) Update(c *gin.Context) {
    id, err := strconv.ParseUint(c.Param("id"), 10, 32)
    if err != nil {
        BadRequest(c, "Invalid ID")
        return
    }
    
    post, err := h.postService.GetByID(uint(id))
    if err != nil {
        NotFound(c, "Post not found")
        return
    }
    
    var req CreatePostRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    if req.Title != "" {
        post.Title = req.Title
    }
    if req.Slug != "" {
        post.Slug = req.Slug
    }
    if req.Content != "" {
        post.Content = req.Content
    }
    if req.Summary != "" {
        post.Summary = req.Summary
    }
    if req.Cover != "" {
        post.Cover = req.Cover
    }
    if req.CategoryID != nil {
        post.CategoryID = req.CategoryID
    }
    if req.Status != "" {
        post.Status = req.Status
    }
    
    if err := h.postService.Update(post, req.Tags); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, post)
}

// ... other methods (GetByID, GetBySlug, Delete) remain the same
```

### 10. 更新 go.mod 依赖

```go
module blog-backend

go 1.21

require (
    github.com/gin-gonic/gin v1.9.1
    github.com/golang-jwt/jwt/v5 v5.2.0
    github.com/joho/godotenv v1.5.1
    golang.org/x/crypto v0.18.0
    gorm.io/driver/mysql v1.5.2
    gorm.io/gorm v1.25.5
)
```

## 创建管理员账户的操作步骤

### 方法1: 使用命令行脚本（推荐）

1. 创建脚本文件 `cmd/create_admin/main.go`
2. 运行命令：
```bash
go run cmd/create_admin/main.go admin admin@example.com your_password
```

### 方法2: 直接修改数据库

1. 进入 MySQL：
```bash
mysql -u root -p blog
```

2. 运行 SQL（密码需要使用 bcrypt 加密）：
```sql
-- 如果还没有用户表，先运行迁移
-- 然后插入管理员
INSERT INTO users (username, email, password, role, created_at, updated_at) 
VALUES ('admin', 'admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MrJ6Yqj7JkQZqHd6J5q6S7q8q9q10q', 'admin', NOW(), NOW());
```

### 方法3: 程序启动时自动创建

在 `database/database.go` 的 `InitDB` 函数中添加：

```go
// 创建默认管理员（如果不存在）
func createDefaultAdmin() {
    var admin models.User
    result := DB.Where("role = ?", "admin").First(&admin)
    if result.Error != nil && errors.Is(result.Error, gorm.ErrRecordNotFound) {
        hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
        admin = models.User{
            Username: "admin",
            Email:    "admin@example.com",
            Password: string(hashedPassword),
            Role:     "admin",
        }
        DB.Create(&admin)
        log.Println("Default admin created: admin@example.com / admin123")
    }
}
```

## API 测试示例

### 创建文章（带分类和标签）
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "My Advanced Post",
    "content": "This is my advanced blog post with categories and tags",
    "summary": "A brief summary",
    "category_id": 1,
    "tags": ["golang", "web", "api"],
    "status": "published"
  }'
```

### 上传文件
```bash
curl -X POST http://localhost:8080/api/v1/upload \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@/path/to/your/image.jpg"
```

### 发表评论
```bash
curl -X POST http://localhost:8080/api/v1/comments \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "content": "Great article!",
    "post_id": 1,
    "parent_id": null
  }'
```

这个完整的模板现在包含了：
- ✅ 用户认证和授权（普通用户、管理员）
- ✅ 文章管理（CRUD）
- ✅ 分类管理
- ✅ 标签系统（支持多对多）
- ✅ 评论系统（嵌套评论、审核）
- ✅ 文件上传（图片、PDF等）
- ✅ 管理员账户创建（多种方法）
- ✅ RESTful API 设计
- ✅ 完整的权限控制

你可以根据需要进一步扩展功能，比如添加用户管理界面、文章搜索、邮件通知等。
