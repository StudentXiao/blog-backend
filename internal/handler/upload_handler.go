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
	userId, _ := c.Get("user_id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	uploads, total, err := h.service.GetUserUploads(userId.(uint), page, pageSize)
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
