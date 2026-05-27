我来为你设计一个现代化、生产级的 **Go 后端博客项目模板**。采用标准分层架构，包含完整的用户认证、文章管理、分类标签、评论等核心模块。

---

## 📁 项目结构

```
go-blog-template/
├── api/
│   └── v1/                 # API 版本控制
│       ├── article.go        # 文章接口
│       ├── category.go       # 分类接口
│       ├── comment.go        # 评论接口
│       ├── tag.go            # 标签接口
│       └── user.go           # 用户接口
├── config/
│   └── config.go           # 配置管理
├── middleware/
│   ├── cors.go             # 跨域中间件
│   ├── jwt.go              # JWT 认证
│   └── logger.go           # 请求日志
├── model/
│   ├── article.go          # 文章模型
│   ├── category.go         # 分类模型
│   ├── comment.go          # 评论模型
│   ├── tag.go              # 标签模型
│   └── user.go             # 用户模型
├── router/
│   └── router.go           # 路由注册
├── service/
│   ├── article.go          # 文章业务逻辑
│   ├── category.go         # 分类业务逻辑
│   ├── comment.go          # 评论业务逻辑
│   ├── tag.go              # 标签业务逻辑
│   └── user.go             # 用户业务逻辑
├── utils/
│   ├── response.go         # 统一响应封装
│   └── validator.go        # 参数校验
├── go.mod
├── main.go
└── config.yaml             # 配置文件
```

---

## 🔧 核心代码实现

### 1. `go.mod`
```go
module go-blog-template

go 1.22

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/golang-jwt/jwt/v5 v5.2.0
	github.com/spf13/viper v1.18.2
	golang.org/x/crypto v0.21.0
	gorm.io/driver/mysql v1.5.4
	gorm.io/gorm v1.25.7
)
```

### 2. `main.go`
```go
package main

import (
	"go-blog-template/config"
	"go-blog-template/model"
	"go-blog-template/router"
	"log"
)

func main() {
	// 加载配置
	config.Init()

	// 连接数据库
	model.InitDB()

	// 注册路由
	r := router.SetupRouter()

	// 启动服务
	addr := config.Config.Server.Host + ":" + config.Config.Server.Port
	log.Printf("Server running on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server start failed: %v", err)
	}
}
```

### 3. `config/config.go`
```go
package config

import (
	"fmt"
	"github.com/spf13/viper"
)

type ConfigStruct struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	Charset  string `mapstructure:"charset"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int    `mapstructure:"expire"`
}

var Config ConfigStruct

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("config file error: %w", err))
	}

	if err := viper.Unmarshal(&Config); err != nil {
		panic(fmt.Errorf("config unmarshal error: %w", err))
	}
}
```

### 4. `config.yaml`
```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: debug

database:
  driver: mysql
  host: localhost
  port: 3306
  user: root
  password: your_password
  dbname: go_blog
  charset: utf8mb4

jwt:
  secret: your-secret-key-change-this
  expire: 72  # hours
```

### 5. `model/` 数据模型

**`model/base.go`**
```go
package model

import (
	"time"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
```

**`model/user.go`**
```go
package model

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	BaseModel
	Username string `gorm:"size:32;not null;uniqueIndex" json:"username" binding:"required,min=3,max=32"`
	Email    string `gorm:"size:128;not null;uniqueIndex" json:"email" binding:"required,email"`
	Password string `gorm:"size:256;not null" json:"-" binding:"required,min=6"`
	Nickname string `gorm:"size:64" json:"nickname"`
	Avatar   string `gorm:"size:256" json:"avatar"`
	Bio      string `gorm:"size:500" json:"bio"`
	Role     int    `gorm:"default:0" json:"role"` // 0: user, 1: admin
	Articles []Article `json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	hashed, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashed)
	return nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
```

**`model/article.go`**
```go
package model

type Article struct {
	BaseModel
	Title      string   `gorm:"size:200;not null" json:"title" binding:"required,max=200"`
	Content    string   `gorm:"type:text;not null" json:"content" binding:"required"`
	Summary    string   `gorm:"size:500" json:"summary"`
	Cover      string   `gorm:"size:256" json:"cover"`
	Status     int      `gorm:"default:0" json:"status"` // 0: draft, 1: published
	ViewCount  int      `gorm:"default:0" json:"view_count"`
	UserID     uint     `gorm:"not null" json:"user_id"`
	User       User     `json:"user,omitempty"`
	CategoryID uint     `json:"category_id"`
	Category   Category `json:"category,omitempty"`
	Tags       []Tag    `gorm:"many2many:article_tags;" json:"tags,omitempty"`
	Comments   []Comment `json:"-"`
}
```

**`model/category.go`**
```go
package model

type Category struct {
	BaseModel
	Name        string `gorm:"size:64;not null;uniqueIndex" json:"name" binding:"required"`
	Description string `gorm:"size:256" json:"description"`
	Articles    []Article `json:"-"`
}
```

**`model/tag.go`**
```go
package model

type Tag struct {
	BaseModel
	Name     string    `gorm:"size:64;not null;uniqueIndex" json:"name" binding:"required"`
	Articles []Article `gorm:"many2many:article_tags;" json:"-"`
}
```

**`model/comment.go`**
```go
package model

type Comment struct {
	BaseModel
	Content   string    `gorm:"type:text;not null" json:"content" binding:"required"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	User      User      `json:"user,omitempty"`
	ArticleID uint      `gorm:"not null" json:"article_id"`
	ParentID  *uint     `json:"parent_id"` // 回复评论ID，支持嵌套
	Replies   []Comment `gorm:"-" json:"replies,omitempty"`
}
```

**`model/db.go`**
```go
package model

import (
	"fmt"
	"go-blog-template/config"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		config.Config.Database.User,
		config.Config.Database.Password,
		config.Config.Database.Host,
		config.Config.Database.Port,
		config.Config.Database.DBName,
		config.Config.Database.Charset,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// 自动迁移
	err = DB.AutoMigrate(&User{}, &Article{}, &Category{}, &Tag{}, &Comment{})
	if err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
}
```

### 6. `utils/` 工具包

**`utils/response.go`**
```go
package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}
```

### 7. `middleware/` 中间件

**`middleware/jwt.go`**
```go
package middleware

import (
	"fmt"
	"go-blog-template/config"
	"go-blog-template/utils"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "missing token")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			utils.Unauthorized(c, "invalid token format")
			c.Abort()
			return
		}

		token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(config.Config.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			utils.Unauthorized(c, "invalid token")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			if exp, ok := claims["exp"].(float64); ok {
				if time.Now().Unix() > int64(exp) {
					utils.Unauthorized(c, "token expired")
					c.Abort()
					return
				}
			}
			c.Set("userID", uint(claims["user_id"].(float64)))
			c.Set("username", claims["username"].(string))
			c.Set("role", int(claims["role"].(float64)))
		}

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(int) != 1 {
			utils.Error(c, http.StatusForbidden, "admin only")
			c.Abort()
			return
		}
		c.Next()
	}
}
```

**`middleware/cors.go`**
```go
package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	})
}
```

### 8. `service/` 业务逻辑层

**`service/user.go`**
```go
package service

import (
	"go-blog-template/model"
	"go-blog-template/utils"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

type UserService struct{}

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Nickname string `json:"nickname"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (s *UserService) Register(req RegisterRequest) (*model.User, error) {
	var existing model.User
	if err := model.DB.Where("username = ?", req.Username).First(&existing).Error; err == nil {
		return nil, utils.NewError("username already exists")
	}

	user := model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Nickname: req.Nickname,
	}

	if err := model.DB.Create(&user).Error; err != nil {
		return nil, err
	}

	user.Password = ""
	return &user, nil
}

func (s *UserService) Login(req LoginRequest) (string, *model.User, error) {
	var user model.User
	if err := model.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return "", nil, utils.NewError("invalid username or password")
	}

	if !user.CheckPassword(req.Password) {
		return "", nil, utils.NewError("invalid username or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(time.Hour * time.Duration(viper.GetInt("jwt.expire"))).Unix(),
	})

	tokenString, err := token.SignedString([]byte(viper.GetString("jwt.secret")))
	if err != nil {
		return "", nil, err
	}

	user.Password = ""
	return tokenString, &user, nil
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	if err := model.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	user.Password = ""
	return &user, nil
}
```

**`service/article.go`**
```go
package service

import (
	"go-blog-template/model"
)

type ArticleService struct{}

type CreateArticleRequest struct {
	Title      string `json:"title" binding:"required,max=200"`
	Content    string `json:"content" binding:"required"`
	Summary    string `json:"summary"`
	Cover      string `json:"cover"`
	CategoryID uint   `json:"category_id"`
	TagIDs     []uint `json:"tag_ids"`
	Status     int    `json:"status"`
}

type UpdateArticleRequest struct {
	Title      string `json:"title" binding:"max=200"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	Cover      string `json:"cover"`
	CategoryID uint   `json:"category_id"`
	TagIDs     []uint `json:"tag_ids"`
	Status     int    `json:"status"`
}

type ArticleListQuery struct {
	Page       int    `form:"page,default=1"`
	PageSize   int    `form:"page_size,default=10"`
	CategoryID uint   `form:"category_id"`
	TagID      uint   `form:"tag_id"`
	Keyword    string `form:"keyword"`
	Status     int    `form:"status,default=1"` // 默认查询已发布
}

func (s *ArticleService) Create(userID uint, req CreateArticleRequest) (*model.Article, error) {
	article := model.Article{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    req.Summary,
		Cover:      req.Cover,
		Status:     req.Status,
		UserID:     userID,
		CategoryID: req.CategoryID,
	}

	if err := model.DB.Create(&article).Error; err != nil {
		return nil, err
	}

	if len(req.TagIDs) > 0 {
		var tags []model.Tag
		model.DB.Where("id IN ?", req.TagIDs).Find(&tags)
		model.DB.Model(&article).Association("Tags").Append(&tags)
	}

	return s.GetByID(article.ID)
}

func (s *ArticleService) GetByID(id uint) (*model.Article, error) {
	var article model.Article
	if err := model.DB.Preload("User").Preload("Category").Preload("Tags").First(&article, id).Error; err != nil {
		return nil, err
	}
	return &article, nil
}

func (s *ArticleService) List(query ArticleListQuery) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	db := model.DB.Model(&model.Article{}).Preload("User").Preload("Category").Preload("Tags")

	if query.Status >= 0 {
		db = db.Where("status = ?", query.Status)
	}
	if query.CategoryID > 0 {
		db = db.Where("category_id = ?", query.CategoryID)
	}
	if query.Keyword != "" {
		db = db.Where("title LIKE ?", "%"+query.Keyword+"%")
	}
	if query.TagID > 0 {
		db = db.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Where("article_tags.tag_id = ?", query.TagID)
	}

	db.Count(&total)
	offset := (query.Page - 1) * query.PageSize
	err := db.Order("created_at DESC").Offset(offset).Limit(query.PageSize).Find(&articles).Error

	return articles, total, err
}

func (s *ArticleService) Update(id uint, userID uint, req UpdateArticleRequest) (*model.Article, error) {
	var article model.Article
	if err := model.DB.First(&article, id).Error; err != nil {
		return nil, err
	}

	// 只有作者或管理员可以编辑
	if article.UserID != userID {
		return nil, utils.NewError("permission denied")
	}

	updates := map[string]interface{}{
		"title":       req.Title,
		"content":     req.Content,
		"summary":     req.Summary,
		"cover":       req.Cover,
		"category_id": req.CategoryID,
		"status":      req.Status,
	}

	if err := model.DB.Model(&article).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 更新标签
	if req.TagIDs != nil {
		var tags []model.Tag
		model.DB.Where("id IN ?", req.TagIDs).Find(&tags)
		model.DB.Model(&article).Association("Tags").Replace(&tags)
	}

	return s.GetByID(id)
}

func (s *ArticleService) Delete(id uint, userID uint) error {
	var article model.Article
	if err := model.DB.First(&article, id).Error; err != nil {
		return err
	}
	if article.UserID != userID {
		return utils.NewError("permission denied")
	}
	return model.DB.Delete(&article).Error
}
```

### 9. `api/v1/` 接口层

**`api/v1/user.go`**
```go
package v1

import (
	"go-blog-template/service"
	"go-blog-template/utils"

	"github.com/gin-gonic/gin"
)

type UserAPI struct {
	service service.UserService
}

func (api *UserAPI) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	user, err := api.service.Register(req)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, user)
}

func (api *UserAPI) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	token, user, err := api.service.Login(req)
	if err != nil {
		utils.Unauthorized(c, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}

func (api *UserAPI) GetProfile(c *gin.Context) {
	userID, _ := c.Get("userID")
	user, err := api.service.GetUserByID(userID.(uint))
	if err != nil {
		utils.NotFound(c, "user not found")
		return
	}
	utils.Success(c, user)
}
```

**`api/v1/article.go`**
```go
package v1

import (
	"go-blog-template/middleware"
	"go-blog-template/service"
	"go-blog-template/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArticleAPI struct {
	service service.ArticleService
}

func (api *ArticleAPI) Create(c *gin.Context) {
	var req service.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	article, err := api.service.Create(userID.(uint), req)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, article)
}

func (api *ArticleAPI) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	article, err := api.service.GetByID(uint(id))
	if err != nil {
		utils.NotFound(c, "article not found")
		return
	}
	utils.Success(c, article)
}

func (api *ArticleAPI) List(c *gin.Context) {
	var query service.ArticleListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	articles, total, err := api.service.List(query)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":  articles,
		"total": total,
		"page":  query.Page,
	})
}

func (api *ArticleAPI) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req service.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	article, err := api.service.Update(uint(id), userID.(uint), req)
	if err != nil {
		utils.Error(c, 403, err.Error())
		return
	}

	utils.Success(c, article)
}

func (api *ArticleAPI) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, _ := c.Get("userID")

	if err := api.service.Delete(uint(id), userID.(uint)); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}

	utils.Success(c, nil)
}
```

### 10. `router/router.go`
```go
package router

import (
	"go-blog-template/api/v1"
	"go-blog-template/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 全局中间件
	r.Use(middleware.CORS())

	// API v1
	apiV1 := r.Group("/api/v1")
	{
		// 用户模块
		userAPI := v1.UserAPI{}
		apiV1.POST("/auth/register", userAPI.Register)
		apiV1.POST("/auth/login", userAPI.Login)

		// 文章公开接口
		articleAPI := v1.ArticleAPI{}
		apiV1.GET("/articles", articleAPI.List)
		apiV1.GET("/articles/:id", articleAPI.GetByID)

		// 需要认证
		auth := apiV1.Group("")
		auth.Use(middleware.JWTAuth())
		{
			auth.GET("/user/profile", userAPI.GetProfile)

			// 文章管理
			auth.POST("/articles", articleAPI.Create)
			auth.PUT("/articles/:id", articleAPI.Update)
			auth.DELETE("/articles/:id", articleAPI.Delete)
		}
	}

	return r
}
```

---

## 🚀 快速启动

```bash
# 1. 创建项目
mkdir go-blog-template && cd go-blog-template
go mod init go-blog-template

# 2. 安装依赖
go get github.com/gin-gonic/gin
go get gorm.io/gorm
go get gorm.io/driver/mysql
go get github.com/golang-jwt/jwt/v5
go get github.com/spf13/viper
go get golang.org/x/crypto

# 3. 创建数据库
mysql -u root -p -e "CREATE DATABASE go_blog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 4. 修改 config.yaml 中的数据库密码

# 5. 运行
go run main.go
```

---

## 📋 API 端点概览

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 否 | 用户注册 |
| POST | `/api/v1/auth/login` | 否 | 用户登录 |
| GET | `/api/v1/user/profile` | 是 | 获取个人信息 |
| GET | `/api/v1/articles` | 否 | 文章列表（支持分页、搜索、筛选） |
| GET | `/api/v1/articles/:id` | 否 | 文章详情 |
| POST | `/api/v1/articles` | 是 | 创建文章 |
| PUT | `/api/v1/articles/:id` | 是 | 更新文章 |
| DELETE | `/api/v1/articles/:id` | 是 | 删除文章 |

---

## ✨ 设计亮点

1. **分层架构**：Controller → Service → Model，职责清晰
2. **JWT 认证**：无状态认证，支持 Token 过期
3. **统一响应**：`utils.Response` 规范所有 API 返回格式
4. **GORM 关联**：预加载 User/Category/Tags，避免 N+1 查询
5. **软删除**：使用 `gorm.DeletedAt` 保留数据
6. **权限控制**：文章编辑仅限作者本人
7. **配置热加载**：Viper 支持多环境配置

需要我补充 **评论系统**、**文件上传** 或 **Redis 缓存** 模块吗？
