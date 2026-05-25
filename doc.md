这是一个功能完整的 Go 后端博客项目模板，包含 RESTful API、JWT 认证、数据库操作等核心功能。

## 项目结构

```
blog-backend/
├── cmd/
│   └── server/
│       └── main.go                 # 应用入口
├── internal/
│   ├── config/
│   │   └── config.go               # 配置管理
│   ├── database/
│   │   └── database.go             # 数据库连接
│   ├── models/
│   │   └── models.go               # 数据模型
│   ├── repository/
│   │   ├── user_repo.go            # 用户数据访问
│   │   └── post_repo.go            # 文章数据访问
│   ├── service/
│   │   ├── auth_service.go         # 认证服务
│   │   └── post_service.go         # 文章服务
│   ├── handler/
│   │   ├── auth_handler.go         # 认证处理器
│   │   ├── post_handler.go         # 文章处理器
│   │   └── response.go             # 统一响应
│   ├── middleware/
│   │   ├── auth.go                 # JWT 认证中间件
│   │   ├── logger.go               # 日志中间件
│   │   └── cors.go                 # CORS 中间件
│   └── router/
│       └── router.go               # 路由配置
├── pkg/
│   └── jwt/
│       └── jwt.go                  # JWT 工具
├── go.mod
├── go.sum
└── .env.example
```

## 核心代码实现

### 1. go.mod

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

### 2. .env.example

```env
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=blog
JWT_SECRET=your-secret-key
JWT_EXPIRE_HOURS=24
PORT=8080
```

### 3. internal/config/config.go

```go
package config

import (
    "log"
    "os"
    "strconv"
    "github.com/joho/godotenv"
)

type Config struct {
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    JWTSecret  string
    JWTExpire  int
    Port       string
}

var AppConfig *Config

func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using system environment")
    }

    expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
    
    AppConfig = &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "3306"),
        DBUser:     getEnv("DB_USER", "root"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "blog"),
        JWTSecret:  getEnv("JWT_SECRET", "default-secret"),
        JWTExpire:  expireHours,
        Port:       getEnv("PORT", "8080"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 4. internal/database/database.go

```go
package database

import (
    "fmt"
    "log"
    "blog-backend/internal/config"
    "blog-backend/internal/models"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.AppConfig.DBUser,
        config.AppConfig.DBPassword,
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBName,
    )

    var err error
    DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        log.Fatal("Failed to connect to database: ", err)
    }

    // Auto migrate schemas
    err = DB.AutoMigrate(&models.User{}, &models.Post{}, &models.Comment{})
    if err != nil {
        log.Fatal("Failed to migrate database: ", err)
    }

    log.Println("Database connected and migrated successfully")
}
```

### 5. internal/models/models.go

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
    Role      string         `gorm:"default:user" json:"role"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    Posts     []Post         `json:"posts,omitempty"`
}

type Post struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Title     string         `gorm:"size:200;not null" json:"title"`
    Slug      string         `gorm:"uniqueIndex;size:200;not null" json:"slug"`
    Content   string         `gorm:"type:longtext;not null" json:"content"`
    Summary   string         `gorm:"size:500" json:"summary"`
    Cover     string         `json:"cover"`
    Views     int            `gorm:"default:0" json:"views"`
    Likes     int            `gorm:"default:0" json:"likes"`
    Status    string         `gorm:"default:draft" json:"status"` // draft, published
    UserID    uint           `gorm:"not null" json:"user_id"`
    User      User           `json:"user,omitempty"`
    Tags      string         `json:"tags"` // JSON string or comma separated
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    Comments  []Comment      `json:"comments,omitempty"`
}

type Comment struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    Content   string         `gorm:"type:text;not null" json:"content"`
    UserID    uint           `gorm:"not null" json:"user_id"`
    User      User           `json:"user,omitempty"`
    PostID    uint           `gorm:"not null" json:"post_id"`
    ParentID  *uint          `json:"parent_id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

### 6. pkg/jwt/jwt.go

```go
package jwt

import (
    "errors"
    "time"
    "blog-backend/internal/config"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func GenerateToken(userID uint, username, role string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.AppConfig.JWTExpire) * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func ParseToken(tokenString string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.AppConfig.JWTSecret), nil
    })
    if err != nil {
        return nil, err
    }
    if claims, ok := token.Claims.(*Claims); ok && token.Valid {
        return claims, nil
    }
    return nil, errors.New("invalid token")
}
```

### 7. internal/repository/user_repo.go

```go
package repository

import (
    "blog-backend/internal/database"
    "blog-backend/internal/models"
    "errors"
    "gorm.io/gorm"
)

type UserRepository struct{}

func (r *UserRepository) Create(user *models.User) error {
    return database.DB.Create(user).Error
}

func (r *UserRepository) FindByID(id uint) (*models.User, error) {
    var user models.User
    err := database.DB.First(&user, id).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
    var user models.User
    err := database.DB.Where("email = ?", email).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *UserRepository) FindByUsername(username string) (*models.User, error) {
    var user models.User
    err := database.DB.Where("username = ?", username).First(&user).Error
    if errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, nil
    }
    return &user, err
}

func (r *UserRepository) Update(user *models.User) error {
    return database.DB.Save(user).Error
}
```

### 8. internal/repository/post_repo.go

```go
package repository

import (
    "blog-backend/internal/database"
    "blog-backend/internal/models"
    "gorm.io/gorm"
)

type PostRepository struct{}

func (r *PostRepository) Create(post *models.Post) error {
    return database.DB.Create(post).Error
}

func (r *PostRepository) FindByID(id uint) (*models.Post, error) {
    var post models.Post
    err := database.DB.Preload("User").Preload("Comments.User").First(&post, id).Error
    return &post, err
}

func (r *PostRepository) FindBySlug(slug string) (*models.Post, error) {
    var post models.Post
    err := database.DB.Preload("User").Where("slug = ?", slug).First(&post).Error
    return &post, err
}

func (r *PostRepository) FindAll(page, pageSize int, status string) ([]models.Post, int64, error) {
    var posts []models.Post
    var total int64
    
    query := database.DB.Model(&models.Post{}).Preload("User")
    if status != "" {
        query = query.Where("status = ?", status)
    }
    
    query.Count(&total)
    
    offset := (page - 1) * pageSize
    err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&posts).Error
    
    return posts, total, err
}

func (r *PostRepository) Update(post *models.Post) error {
    return database.DB.Save(post).Error
}

func (r *PostRepository) Delete(id uint) error {
    return database.DB.Delete(&models.Post{}, id).Error
}

func (r *PostRepository) IncrementViews(id uint) error {
    return database.DB.Model(&models.Post{}).Where("id = ?", id).UpdateColumn("views", gorm.Expr("views + ?", 1)).Error
}
```

### 9. internal/service/auth_service.go

```go
package service

import (
    "blog-backend/internal/models"
    "blog-backend/internal/repository"
    "blog-backend/pkg/jwt"
    "errors"
    "golang.org/x/crypto/bcrypt"
)

type AuthService struct {
    userRepo *repository.UserRepository
}

func NewAuthService() *AuthService {
    return &AuthService{
        userRepo: &repository.UserRepository{},
    }
}

func (s *AuthService) Register(username, email, password string) (*models.User, string, error) {
    // Check if user exists
    existingUser, _ := s.userRepo.FindByEmail(email)
    if existingUser != nil {
        return nil, "", errors.New("email already registered")
    }
    
    existingUser, _ = s.userRepo.FindByUsername(username)
    if existingUser != nil {
        return nil, "", errors.New("username already taken")
    }
    
    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, "", err
    }
    
    user := &models.User{
        Username: username,
        Email:    email,
        Password: string(hashedPassword),
        Role:     "user",
    }
    
    if err := s.userRepo.Create(user); err != nil {
        return nil, "", err
    }
    
    // Generate token
    token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
    if err != nil {
        return nil, "", err
    }
    
    return user, token, nil
}

func (s *AuthService) Login(email, password string) (*models.User, string, error) {
    user, err := s.userRepo.FindByEmail(email)
    if err != nil || user == nil {
        return nil, "", errors.New("invalid credentials")
    }
    
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
        return nil, "", errors.New("invalid credentials")
    }
    
    token, err := jwt.GenerateToken(user.ID, user.Username, user.Role)
    if err != nil {
        return nil, "", err
    }
    
    return user, token, nil
}
```

### 10. internal/service/post_service.go

```go
package service

import (
    "blog-backend/internal/models"
    "blog-backend/internal/repository"
    "errors"
    "strings"
)

type PostService struct {
    postRepo *repository.PostRepository
}

func NewPostService() *PostService {
    return &PostService{
        postRepo: &repository.PostRepository{},
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
```

### 11. internal/handler/response.go

```go
package handler

import (
    "net/http"
    "github.com/gin-gonic/gin"
)

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    200,
        Message: "success",
        Data:    data,
    })
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    200,
        Message: message,
        Data:    data,
    })
}

func Error(c *gin.Context, code int, message string) {
    c.JSON(code, Response{
        Code:    code,
        Message: message,
    })
}

func BadRequest(c *gin.Context, message string) {
    Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
    Error(c, http.StatusUnauthorized, message)
}

func Forbidden(c *gin.Context, message string) {
    Error(c, http.StatusForbidden, message)
}

func NotFound(c *gin.Context, message string) {
    Error(c, http.StatusNotFound, message)
}

func InternalServerError(c *gin.Context, message string) {
    Error(c, http.StatusInternalServerError, message)
}
```

### 12. internal/handler/auth_handler.go

```go
package handler

import (
    "blog-backend/internal/service"
    "github.com/gin-gonic/gin"
)

type AuthHandler struct {
    authService *service.AuthService
}

func NewAuthHandler() *AuthHandler {
    return &AuthHandler{
        authService: service.NewAuthService(),
    }
}

type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    user, token, err := h.authService.Register(req.Username, req.Email, req.Password)
    if err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "user":  user,
        "token": token,
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        BadRequest(c, err.Error())
        return
    }
    
    user, token, err := h.authService.Login(req.Email, req.Password)
    if err != nil {
        Unauthorized(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "user":  user,
        "token": token,
    })
}
```

### 13. internal/handler/post_handler.go

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
    Title   string `json:"title" binding:"required"`
    Slug    string `json:"slug"`
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
        "posts": posts,
        "total": total,
        "page":  page,
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
```

### 14. internal/middleware/auth.go

```go
package middleware

import (
    "blog-backend/internal/handler"
    "blog-backend/pkg/jwt"
    "strings"
    "github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            handler.Unauthorized(c, "Missing authorization header")
            c.Abort()
            return
        }
        
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            handler.Unauthorized(c, "Invalid authorization format")
            c.Abort()
            return
        }
        
        claims, err := jwt.ParseToken(parts[1])
        if err != nil {
            handler.Unauthorized(c, "Invalid or expired token")
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("user_role", claims.Role)
        c.Next()
    }
}

func AdminOnly() gin.HandlerFunc {
    return func(c *gin.Context) {
        role, exists := c.Get("user_role")
        if !exists || role != "admin" {
            handler.Forbidden(c, "Admin access required")
            c.Abort()
            return
        }
        c.Next()
    }
}
```

### 15. internal/middleware/logger.go

```go
package middleware

import (
    "log"
    "time"
    "github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        startTime := time.Now()
        
        c.Next()
        
        endTime := time.Now()
        latency := endTime.Sub(startTime)
        
        log.Printf("[%s] %s %s %d %v",
            c.Request.Method,
            c.Request.URL.Path,
            c.ClientIP(),
            c.Writer.Status(),
            latency,
        )
    }
}
```

### 16. internal/middleware/cors.go

```go
package middleware

import (
    "github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
        
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        
        c.Next()
    }
}
```

### 17. internal/router/router.go

```go
package router

import (
    "blog-backend/internal/handler"
    "blog-backend/internal/middleware"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    router := gin.Default()
    
    // Global middleware
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())
    
    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // Public routes
    api := router.Group("/api/v1")
    {
        authHandler := handler.NewAuthHandler()
        postHandler := handler.NewPostHandler()
        
        // Auth endpoints
        api.POST("/auth/register", authHandler.Register)
        api.POST("/auth/login", authHandler.Login)
        
        // Public post endpoints
        api.GET("/posts", postHandler.List)
        api.GET("/posts/slug/:slug", postHandler.GetBySlug)
        api.GET("/posts/:id", postHandler.GetByID)
        
        // Protected routes
        protected := api.Group("/")
        protected.Use(middleware.Auth())
        {
            // Post management
            protected.POST("/posts", postHandler.Create)
            protected.PUT("/posts/:id", postHandler.Update)
            protected.DELETE("/posts/:id", postHandler.Delete)
        }
        
        // Admin only routes
        admin := api.Group("/admin")
        admin.Use(middleware.Auth(), middleware.AdminOnly())
        {
            admin.GET("/posts/all", postHandler.List) // List all posts including drafts
        }
    }
    
    return router
}
```

### 18. cmd/server/main.go

```go
package main

import (
    "fmt"
    "log"
    "blog-backend/internal/config"
    "blog-backend/internal/database"
    "blog-backend/internal/router"
)

func main() {
    // Load configuration
    config.LoadConfig()
    
    // Initialize database
    database.InitDB()
    
    // Setup router
    r := router.SetupRouter()
    
    // Start server
    addr := fmt.Sprintf(":%s", config.AppConfig.Port)
    log.Printf("Server starting on %s", addr)
    if err := r.Run(addr); err != nil {
        log.Fatal("Failed to start server: ", err)
    }
}
```

## 使用方法

### 1. 安装依赖
```bash
go mod tidy
```

### 2. 配置数据库
```bash
cp .env.example .env
# 编辑 .env 文件，配置数据库连接信息
```

### 3. 创建数据库
```sql
CREATE DATABASE blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 4. 运行项目
```bash
go run cmd/server/main.go
```

## API 端点

### 公开端点
- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/posts` - 获取文章列表
- `GET /api/v1/posts/:id` - 获取单篇文章
- `GET /api/v1/posts/slug/:slug` - 通过slug获取文章

### 需要认证的端点
- `POST /api/v1/posts` - 创建文章
- `PUT /api/v1/posts/:id` - 更新文章
- `DELETE /api/v1/posts/:id` - 删除文章

### 管理员端点
- `GET /api/v1/admin/posts/all` - 获取所有文章（包括草稿）

## 测试示例

### 注册用户
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "john",
    "email": "john@example.com",
    "password": "123456"
  }'
```

### 登录
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "123456"
  }'
```

### 创建文章（需要token）
```bash
curl -X POST http://localhost:8080/api/v1/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "My First Post",
    "content": "This is my first blog post",
    "status": "published"
  }'
```

这个模板提供了一个完整的博客后端基础框架，包含了用户认证、文章管理、数据库操作等核心功能。你可以根据需要添加更多功能，如评论系统、分类管理、标签系统、文件上传等。
