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
	Title   string `json:"title" binding:"required"`
	Slug    string `json:"sulg"`
	Content string `json:"content" binding:"required"`
	Summary string `json:"summary"`
	Cover   string `json:"cover"`
	Tags    string `json:"tags"`
	Status  string `json:"status"`
}

func (h *PostHandler) Create(c *gin.Context) {
	var req CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	post := &models.Post{
		Title:   req.Title,
		Slug:    req.Slug,
		Content: req.Content,
		Summary: req.Summary,
		Cover:   req.Cover,
		Tags:    req.Tags,
		Status:  req.Status,
		UserID:  userID.(uint),
	}

	if post.Status == "" {
		post.Status = "draft"
	}

	if err := h.postService.Create(post); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, post)
}

func (h *PostHandler) GetByID(c *gin.Context) {
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

	Success(c, post)
}

func (h *PostHandler) GetBySlug(c *gin.Context) {
	slug := c.Param("slug")

	post, err := h.postService.GetBySlug(slug)
	if err != nil {
		NotFound(c, "Post not found")
		return
	}

	Success(c, post)
}

func (h *PostHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	status := c.Query("status")

	// Only show published posts for non-admin
	if _, exists := c.Get("user_role"); !exists && status == "" {
		status = "published"
	}

	posts, total, err := h.postService.List(page, pageSize, status)
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
	if req.Tags != "" {
		post.Tags = req.Tags
	}
	if req.Status != "" {
		post.Status = req.Status
	}

	if err := h.postService.Update(post); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, post)
}

func (h *PostHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		BadRequest(c, "Invalid ID")
		return
	}

	if err := h.postService.Delete(uint(id)); err != nil {
		InternalServerError(c, err.Error())
		return
	}

	SuccessWithMessage(c, "Post deleted successfully", nil)
}
