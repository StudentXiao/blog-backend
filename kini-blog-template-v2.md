我来为你补充这三个核心模块，保持与之前项目架构完全一致。

---

## 一、依赖安装

```bash
go get github.com/redis/go-redis/v9
go get github.com/google/uuid
```

---

## 二、配置文件更新

### `config/config.go` 新增配置
```go
type ConfigStruct struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Redis    RedisConfig    `mapstructure:"redis"`      // 新增
	Upload   UploadConfig   `mapstructure:"upload"`     // 新增
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type UploadConfig struct {
	Path      string   `mapstructure:"path"`       // 上传目录
	MaxSize   int64    `mapstructure:"max_size"`   // 最大文件大小(MB)
	AllowExts []string `mapstructure:"allow_exts"` // 允许后缀
	BaseURL   string   `mapstructure:"base_url"`   // 访问基础URL
}
```

### `config.yaml` 新增配置
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
  expire: 72

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
  pool_size: 10

upload:
  path: ./uploads
  max_size: 5
  allow_exts: [".jpg", ".jpeg", ".png", ".gif", ".webp"]
  base_url: http://localhost:8080/uploads
```

---

## 三、Redis 模块

### `utils/redis.go` — Redis 客户端与缓存工具
```go
package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"go-blog-template/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var ctx = context.Background()

func InitRedis() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Config.Redis.Host, config.Config.Redis.Port),
		Password: config.Config.Redis.Password,
		DB:       config.Config.Redis.DB,
		PoolSize: config.Config.Redis.PoolSize,
	})

	if err := RDB.Ping(ctx).Err(); err != nil {
		panic(fmt.Errorf("redis connection failed: %w", err))
	}
}

// CacheSet 设置缓存
func CacheSet(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RDB.Set(ctx, key, data, expiration).Err()
}

// CacheGet 获取缓存
func CacheGet(key string, dest interface{}) error {
	data, err := RDB.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// CacheDelete 删除缓存
func CacheDelete(keys ...string) error {
	return RDB.Del(ctx, keys...).Err()
}

// CachePatternDelete 按模式删除缓存
func CachePatternDelete(pattern string) error {
	keys, err := RDB.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return RDB.Del(ctx, keys...).Err()
	}
	return nil
}

// IncrViewCount 增加浏览量（先写Redis，定期同步到DB）
func IncrViewCount(articleID uint) error {
	key := fmt.Sprintf("article:view:%d", articleID)
	return RDB.Incr(ctx, key).Err()
}

// GetViewCount 获取浏览量
func GetViewCount(articleID uint) (int64, error) {
	key := fmt.Sprintf("article:view:%d", articleID)
	return RDB.Get(ctx, key).Int64()
}
```

### `model/db.go` 更新（添加Redis初始化）
```go
func InitDB() {
	// ... 原有MySQL连接代码不变 ...
	
	// 初始化Redis
	utils.InitRedis()
}
```

---

## 四、评论系统模块

### `service/comment.go` — 评论业务逻辑
```go
package service

import (
	"fmt"
	"go-blog-template/model"
	"go-blog-template/utils"
)

type CommentService struct{}

type CreateCommentRequest struct {
	Content   string `json:"content" binding:"required,max=2000"`
	ArticleID uint   `json:"article_id" binding:"required"`
	ParentID  *uint  `json:"parent_id"` // 回复某条评论
}

type CommentTree struct {
	model.Comment
	Replies []CommentTree `json:"replies"`
}

func (s *CommentService) Create(userID uint, req CreateCommentRequest) (*model.Comment, error) {
	// 检查文章是否存在
	var article model.Article
	if err := model.DB.First(&article, req.ArticleID).Error; err != nil {
		return nil, utils.NewError("article not found")
	}

	// 如果回复评论，检查父评论是否存在
	if req.ParentID != nil {
		var parent model.Comment
		if err := model.DB.First(&parent, *req.ParentID).Error; err != nil {
			return nil, utils.NewError("parent comment not found")
		}
		if parent.ArticleID != req.ArticleID {
			return nil, utils.NewError("parent comment not in this article")
		}
	}

	comment := model.Comment{
		Content:   req.Content,
		UserID:    userID,
		ArticleID: req.ArticleID,
		ParentID:  req.ParentID,
	}

	if err := model.DB.Create(&comment).Error; err != nil {
		return nil, err
	}

	// 预加载用户信息返回
	model.DB.Preload("User").First(&comment, comment.ID)
	return &comment, nil
}

func (s *CommentService) GetArticleComments(articleID uint, page, pageSize int) ([]CommentTree, int64, error) {
	var comments []model.Comment
	var total int64

	// 查询该文章所有评论（不分页，内存构建树后分页顶层）
	db := model.DB.Model(&model.Comment{}).Where("article_id = ?", articleID).Preload("User")
	db.Count(&total)

	err := db.Order("created_at ASC").Find(&comments).Error
	if err != nil {
		return nil, 0, err
	}

	// 构建评论树
	commentMap := make(map[uint]*CommentTree)
	for i := range comments {
		ct := CommentTree{
			Comment: comments[i],
			Replies: []CommentTree{},
		}
		commentMap[comments[i].ID] = &ct
	}

	var trees []CommentTree
	for i := range comments {
		ct := commentMap[comments[i].ID]
		if comments[i].ParentID == nil {
			trees = append(trees, *ct)
		} else {
			if parent, ok := commentMap[*comments[i].ParentID]; ok {
				parent.Replies = append(parent.Replies, *ct)
			}
		}
	}

	// 顶层评论分页
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(trees) {
		return []CommentTree{}, total, nil
	}
	if end > len(trees) {
		end = len(trees)
	}

	return trees[start:end], total, nil
}

func (s *CommentService) Delete(id uint, userID uint) error {
	var comment model.Comment
	if err := model.DB.First(&comment, id).Error; err != nil {
		return err
	}

	// 只有评论作者或文章作者可以删除
	if comment.UserID != userID {
		// 检查是否是文章作者
		var article model.Article
		model.DB.First(&article, comment.ArticleID)
		if article.UserID != userID {
			return utils.NewError("permission denied")
		}
	}

	// 级联删除子评论
	return model.DB.Where("id = ? OR parent_id = ?", id, id).Delete(&model.Comment{}).Error
}
```

### `api/v1/comment.go` — 评论接口
```go
package v1

import (
	"go-blog-template/middleware"
	"go-blog-template/service"
	"go-blog-template/utils"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CommentAPI struct {
	service service.CommentService
}

func (api *CommentAPI) Create(c *gin.Context) {
	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.BadRequest(c, err.Error())
		return
	}

	userID, _ := c.Get("userID")
	comment, err := api.service.Create(userID.(uint), req)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, comment)
}

func (api *CommentAPI) List(c *gin.Context) {
	articleID, _ := strconv.Atoi(c.Query("article_id"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	comments, total, err := api.service.GetArticleComments(uint(articleID), page, pageSize)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, gin.H{
		"list":  comments,
		"total": total,
		"page":  page,
	})
}

func (api *CommentAPI) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	userID, _ := c.Get("userID")

	if err := api.service.Delete(uint(id), userID.(uint)); err != nil {
		utils.Error(c, 403, err.Error())
		return
	}

	utils.Success(c, nil)
}
```

---

## 五、文件上传模块

### `service/upload.go` — 上传业务逻辑
```go
package service

import (
	"fmt"
	"go-blog-template/config"
	"go-blog-template/utils"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UploadService struct{}

type UploadResult struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
}

func (s *UploadService) UploadImage(fileHeader *multipart.FileHeader) (*UploadResult, error) {
	// 检查文件大小
	maxSize := config.Config.Upload.MaxSize * 1024 * 1024
	if fileHeader.Size > maxSize {
		return nil, utils.NewError(fmt.Sprintf("file too large, max %dMB", config.Config.Upload.MaxSize))
	}

	// 检查文件后缀
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowed := false
	for _, allowExt := range config.Config.Upload.AllowExts {
		if ext == allowExt {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, utils.NewError("file type not allowed")
	}

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建上传目录（按年月分文件夹）
	now := time.Now()
	dir := filepath.Join(config.Config.Upload.Path, "images", now.Format("200601"))
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	// 生成唯一文件名
	filename := uuid.New().String() + ext
	dst := filepath.Join(dir, filename)

	// 保存文件
	out, err := os.Create(dst)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	if _, err = io.Copy(out, file); err != nil {
		return nil, err
	}

	// 构建访问URL
	url := fmt.Sprintf("%s/images/%s/%s", config.Config.Upload.BaseURL, now.Format("200601"), filename)

	return &UploadResult{
		URL:      url,
		Filename: filename,
		Size:     fileHeader.Size,
	}, nil
}
```

### `api/v1/upload.go` — 上传接口
```go
package v1

import (
	"go-blog-template/service"
	"go-blog-template/utils"

	"github.com/gin-gonic/gin"
)

type UploadAPI struct {
	service service.UploadService
}

func (api *UploadAPI) UploadImage(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		utils.BadRequest(c, "please select a file")
		return
	}

	result, err := api.service.UploadImage(fileHeader)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	utils.Success(c, result)
}
```

---

## 六、文章模块增强（加入Redis缓存）

### `service/article.go` 更新 — 缓存集成
```go
package service

import (
	"encoding/json"
	"fmt"
	"go-blog-template/model"
	"go-blog-template/utils"
	"time"
)

// ... 原有结构体不变 ...

func (s *ArticleService) GetByID(id uint) (*model.Article, error) {
	cacheKey := fmt.Sprintf("article:%d", id)

	// 尝试从缓存获取
	var cached model.Article
	if err := utils.CacheGet(cacheKey, &cached); err == nil {
		// 合并Redis中的实时浏览量
		if viewCount, err := utils.GetViewCount(id); err == nil {
			cached.ViewCount = int(viewCount)
		}
		return &cached, nil
	}

	// 缓存未命中，查数据库
	var article model.Article
	if err := model.DB.Preload("User").Preload("Category").Preload("Tags").First(&article, id).Error; err != nil {
		return nil, err
	}

	// 写入缓存（10分钟TTL）
	utils.CacheSet(cacheKey, article, 10*time.Minute)

	return &article, nil
}

func (s *ArticleService) List(query ArticleListQuery) ([]model.Article, int64, error) {
	cacheKey := fmt.Sprintf("articles:list:%d:%d:%d:%d:%s:%d",
		query.Page, query.PageSize, query.CategoryID, query.TagID, query.Keyword, query.Status)

	// 尝试从缓存获取列表
	var cached struct {
		Articles []model.Article
		Total    int64
	}
	if err := utils.CacheGet(cacheKey, &cached); err == nil {
		return cached.Articles, cached.Total, nil
	}

	// 数据库查询（原有逻辑）
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
	if err != nil {
		return nil, 0, err
	}

	// 写入缓存（5分钟TTL，列表缓存时间短一些）
	cached.Articles = articles
	cached.Total = total
	utils.CacheSet(cacheKey, cached, 5*time.Minute)

	return articles, total, nil
}

// 更新和删除时清除缓存
func (s *ArticleService) clearArticleCache(id uint) {
	utils.CacheDelete(fmt.Sprintf("article:%d", id))
	utils.CachePatternDelete("articles:list:*")
}

func (s *ArticleService) Update(id uint, userID uint, req UpdateArticleRequest) (*model.Article, error) {
	// ... 原有更新逻辑 ...
	// 更新成功后清除缓存
	s.clearArticleCache(id)
	return s.GetByID(id)
}

func (s *ArticleService) Delete(id uint, userID uint) error {
	// ... 原有删除逻辑 ...
	// 删除成功后清除缓存
	s.clearArticleCache(id)
	return nil
}

// IncrementView 增加浏览量（使用Redis计数，异步持久化）
func (s *ArticleService) IncrementView(id uint) error {
	return utils.IncrViewCount(id)
}
```

### `api/v1/article.go` 新增浏览量接口
```go
func (api *ArticleAPI) IncrementView(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := api.service.IncrementView(uint(id)); err != nil {
		utils.Error(c, 500, err.Error())
		return
	}
	utils.Success(c, nil)
}
```

---

## 七、路由注册更新

### `router/router.go` 完整更新版
```go
package router

import (
	"go-blog-template/api/v1"
	"go-blog-template/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 静态文件服务（上传文件访问）
	r.Static("/uploads", "./uploads")

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
		apiV1.POST("/articles/:id/view", articleAPI.IncrementView) // 浏览量

		// 评论公开接口
		commentAPI := v1.CommentAPI{}
		apiV1.GET("/comments", commentAPI.List)

		// 需要认证
		auth := apiV1.Group("")
		auth.Use(middleware.JWTAuth())
		{
			auth.GET("/user/profile", userAPI.GetProfile)

			// 文章管理
			auth.POST("/articles", articleAPI.Create)
			auth.PUT("/articles/:id", articleAPI.Update)
			auth.DELETE("/articles/:id", articleAPI.Delete)

			// 评论管理
			auth.POST("/comments", commentAPI.Create)
			auth.DELETE("/comments/:id", commentAPI.Delete)

			// 文件上传
			uploadAPI := v1.UploadAPI{}
			auth.POST("/upload/image", uploadAPI.UploadImage)
		}
	}

	return r
}
```

---

## 八、Redis 数据同步任务（可选）

### `task/sync_view.go` — 定时同步浏览量到MySQL
```go
package task

import (
	"context"
	"fmt"
	"go-blog-template/model"
	"go-blog-template/utils"
	"strconv"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
)

func StartCron() {
	c := cron.New()
	// 每5分钟同步一次浏览量
	c.AddFunc("*/5 * * * *", syncViewCount)
	c.Start()
}

func syncViewCount() {
	ctx := context.Background()
	keys, err := utils.RDB.Keys(ctx, "article:view:*").Result()
	if err != nil || len(keys) == 0 {
		return
	}

	for _, key := range keys {
		parts := strings.Split(key, ":")
		if len(parts) != 3 {
			continue
		}
		articleID, _ := strconv.Atoi(parts[2])
		count, err := utils.RDB.Get(ctx, key).Int64()
		if err != nil {
			continue
		}

		// 更新数据库
		model.DB.Model(&model.Article{}).Where("id = ?", articleID).
			UpdateColumn("view_count", model.DB.Raw("view_count + ?", count))

		// 删除Redis计数器
		utils.RDB.Del(ctx, key)
	}
}
```

---

## 九、API 端点完整概览

| 方法 | 路径 | 认证 | 说明 |
|------|------|------|------|
| POST | `/api/v1/auth/register` | 否 | 注册 |
| POST | `/api/v1/auth/login` | 否 | 登录 |
| GET | `/api/v1/user/profile` | 是 | 个人信息 |
| GET | `/api/v1/articles` | 否 | 文章列表（Redis缓存） |
| GET | `/api/v1/articles/:id` | 否 | 文章详情（Redis缓存） |
| POST | `/api/v1/articles/:id/view` | 否 | 增加浏览量（Redis计数） |
| POST | `/api/v1/articles` | 是 | 创建文章 |
| PUT | `/api/v1/articles/:id` | 是 | 更新文章（清缓存） |
| DELETE | `/api/v1/articles/:id` | 是 | 删除文章（清缓存） |
| GET | `/api/v1/comments` | 否 | 评论列表（嵌套树形） |
| POST | `/api/v1/comments` | 是 | 发表评论（支持回复） |
| DELETE | `/api/v1/comments/:id` | 是 | 删除评论（级联子评论） |
| POST | `/api/v1/upload/image` | 是 | 上传图片 |

---

## 十、设计亮点总结

1. **评论嵌套**：内存构建树形结构，支持无限层级回复，级联删除子评论
2. **文件上传**：按年月分目录存储，UUID防重名，白名单限制文件类型，大小限制
3. **Redis缓存策略**：
   - 文章详情缓存 10 分钟
   - 文章列表缓存 5 分钟
   - 更新/删除时主动失效缓存（Pattern Delete）
   - 浏览量先写Redis，定时批量同步到MySQL（减少DB写压力）
4. **静态资源**：通过 `r.Static` 暴露 uploads 目录，可直接访问上传的图片

需要我再补充 **Elasticsearch 搜索**、**消息通知（WebSocket）** 或 **Docker 部署配置** 吗？
