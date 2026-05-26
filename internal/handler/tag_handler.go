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
	var req CreateCategoryRequest
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
