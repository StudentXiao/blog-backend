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
		service: &service.CategoryService{},
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
