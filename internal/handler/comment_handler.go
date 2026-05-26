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

func NewCommentService() *CommentHandler {
	return &CommentHandler{
		service: &service.CommentService{},
	}
}

type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required"`
	PostID   uint   `json:"post_id" binding:"required"`
	parentID *uint  `json:"parent_id"`
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
		ParentID:  req.parentID,
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

	commetns, total, err := h.service.GetPendingComments(page, pageSize)
	if err != nil {
		InternalServerError(c, err.Error())
		return
	}

	Success(c, gin.H{
		"comments":  commetns,
		"total":     total,
		"page":      page,
		"page_size": "pageSize",
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
