我来为这个博客项目添加完整的日志记录功能，包括请求日志、错误日志、操作日志和数据库日志。

## 日志记录功能实现

### 1. 安装日志依赖

更新 `go.mod`：

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
    github.com/sirupsen/logrus v1.9.3
    github.com/lestrrat-go/file-rotatelogs v2.4.0+incompatible
    github.com/rifflock/lfshook v0.0.0-20180920164130-b9218ef580f5
    github.com/gin-contrib/requestid v0.0.6
)
```

### 2. 日志配置结构

#### internal/config/config.go (添加日志配置)

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
    
    // 日志配置
    LogLevel      string // debug, info, warn, error
    LogPath       string
    LogMaxAge     int // 日志保留天数
    LogRotationTime int // 日志轮转时间（小时）
}

var AppConfig *Config

func LoadConfig() {
    err := godotenv.Load()
    if err != nil {
        log.Println("No .env file found, using system environment")
    }

    expireHours, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
    logMaxAge, _ := strconv.Atoi(getEnv("LOG_MAX_AGE", "30"))
    logRotationTime, _ := strconv.Atoi(getEnv("LOG_ROTATION_TIME", "24"))
    
    AppConfig = &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "3306"),
        DBUser:     getEnv("DB_USER", "root"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "blog"),
        JWTSecret:  getEnv("JWT_SECRET", "default-secret"),
        JWTExpire:  expireHours,
        Port:       getEnv("PORT", "8080"),
        
        LogLevel:      getEnv("LOG_LEVEL", "info"),
        LogPath:       getEnv("LOG_PATH", "logs"),
        LogMaxAge:     logMaxAge,
        LogRotationTime: logRotationTime,
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

### 3. 日志核心模块

#### pkg/logger/logger.go

```go
package logger

import (
    "blog-backend/internal/config"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
    
    rotatelogs "github.com/lestrrat-go/file-rotatelogs"
    "github.com/rifflock/lfshook"
    "github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// InitLogger 初始化日志系统
func InitLogger() error {
    Log = logrus.New()
    
    // 设置日志格式
    Log.SetFormatter(&logrus.JSONFormatter{
        TimestampFormat: "2006-01-02 15:04:05",
        PrettyPrint:     false,
    })
    
    // 设置日志级别
    level, err := logrus.ParseLevel(config.AppConfig.LogLevel)
    if err != nil {
        level = logrus.InfoLevel
    }
    Log.SetLevel(level)
    
    // 创建日志目录
    logPath := config.AppConfig.LogPath
    if err := os.MkdirAll(logPath, 0755); err != nil {
        return fmt.Errorf("failed to create log directory: %v", err)
    }
    
    // 设置日志轮转
    path := filepath.Join(logPath, "blog.log")
    writer, err := rotatelogs.New(
        path+".%Y%m%d",
        rotatelogs.WithLinkName(path),
        rotatelogs.WithMaxAge(time.Duration(config.AppConfig.LogMaxAge)*24*time.Hour),
        rotatelogs.WithRotationTime(time.Duration(config.AppConfig.LogRotationTime)*time.Hour),
    )
    if err != nil {
        return fmt.Errorf("failed to create rotatelogs: %v", err)
    }
    
    // 设置不同级别的日志输出
    levelMap := lfshook.WriterMap{
        logrus.DebugLevel: writer,
        logrus.InfoLevel:  writer,
        logrus.WarnLevel:  writer,
        logrus.ErrorLevel: writer,
        logrus.FatalLevel: writer,
        logrus.PanicLevel: writer,
    }
    
    // 同时输出到控制台和文件
    Log.SetOutput(io.MultiWriter(os.Stdout, writer))
    Log.AddHook(lfshook.NewHook(levelMap, &logrus.JSONFormatter{
        TimestampFormat: "2006-01-02 15:04:05",
    }))
    
    return nil
}

// 封装常用的日志方法
func Debug(args ...interface{}) {
    Log.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
    Log.Debugf(format, args...)
}

func Info(args ...interface{}) {
    Log.Info(args...)
}

func Infof(format string, args ...interface{}) {
    Log.Infof(format, args...)
}

func Warn(args ...interface{}) {
    Log.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
    Log.Warnf(format, args...)
}

func Error(args ...interface{}) {
    Log.Error(args...)
}

func Errorf(format string, args ...interface{}) {
    Log.Errorf(format, args...)
}

func Fatal(args ...interface{}) {
    Log.Fatal(args...)
}

func Fatalf(format string, args ...interface{}) {
    Log.Fatalf(format, args...)
}

// WithFields 添加结构化字段
func WithFields(fields map[string]interface{}) *logrus.Entry {
    return Log.WithFields(logrus.Fields(fields))
}

// WithField 添加单个字段
func WithField(key string, value interface{}) *logrus.Entry {
    return Log.WithField(key, value)
}
```

### 4. 中间件日志

#### internal/middleware/logger.go (增强版)

```go
package middleware

import (
    "blog-backend/pkg/logger"
    "bytes"
    "io"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/requestid"
)

// 自定义响应写入器，用于捕获响应体
type bodyLogWriter struct {
    gin.ResponseWriter
    body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
    w.body.Write(b)
    return w.ResponseWriter.Write(b)
}

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery
        
        // 读取请求体
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }
        
        // 包装响应写入器
        blw := &bodyLogWriter{
            ResponseWriter: c.Writer,
            body:           bytes.NewBufferString(""),
        }
        c.Writer = blw
        
        // 处理请求
        c.Next()
        
        // 计算响应时间
        latency := time.Since(start)
        
        // 获取请求ID
        reqID := requestid.Get(c)
        
        // 记录日志
        entry := logger.WithFields(map[string]interface{}{
            "request_id":   reqID,
            "method":       c.Request.Method,
            "path":         path,
            "query":        query,
            "ip":           c.ClientIP(),
            "status":       c.Writer.Status(),
            "latency":      latency.String(),
            "user_agent":   c.Request.UserAgent(),
            "error":        c.Errors.String(),
            "request_body": string(requestBody),
        })
        
        // 根据状态码使用不同日志级别
        if c.Writer.Status() >= 500 {
            entry.Error("Request failed")
        } else if c.Writer.Status() >= 400 {
            entry.Warn("Request warning")
        } else {
            entry.Info("Request completed")
        }
        
        // 记录响应体（可选，生产环境建议关闭）
        if gin.Mode() == gin.DebugMode {
            logger.WithFields(map[string]interface{}{
                "request_id": reqID,
                "response":   blw.body.String(),
            }).Debug("Response body")
        }
    }
}

// Recovery 恢复中间件，记录panic日志
func Recovery() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                reqID := requestid.Get(c)
                logger.WithFields(map[string]interface{}{
                    "request_id": reqID,
                    "error":      err,
                    "stack":      string(debug.Stack()),
                }).Error("Panic recovered")
                c.AbortWithStatusJSON(500, gin.H{
                    "code":    500,
                    "message": "Internal server error",
                })
            }
        }()
        c.Next()
    }
}

// 需要导入 runtime/debug
import "runtime/debug"
```

### 5. 数据库日志

#### internal/database/database.go (添加日志)

```go
package database

import (
    "blog-backend/internal/config"
    "blog-backend/internal/models"
    "blog-backend/pkg/logger"
    "fmt"
    "time"
    
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// 自定义GORM日志
type GORMLogger struct {
    SlowThreshold time.Duration
}

func (l *GORMLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
    return l
}

func (l *GORMLogger) Info(ctx context.Context, msg string, data ...interface{}) {
    logger.Infof(msg, data...)
}

func (l *GORMLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
    logger.Warnf(msg, data...)
}

func (l *GORMLogger) Error(ctx context.Context, msg string, data ...interface{}) {
    logger.Errorf(msg, data...)
}

func (l *GORMLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
    elapsed := time.Since(begin)
    sql, rows := fc()
    
    fields := map[string]interface{}{
        "sql":      sql,
        "rows":     rows,
        "duration": elapsed,
    }
    
    if err != nil {
        fields["error"] = err
        logger.WithFields(fields).Error("Database query error")
    } else if elapsed > l.SlowThreshold {
        fields["slow_query"] = true
        logger.WithFields(fields).Warn("Slow SQL query")
    } else if gin.Mode() == gin.DebugMode {
        logger.WithFields(fields).Debug("Database query")
    }
}

func InitDB() {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
        config.AppConfig.DBUser,
        config.AppConfig.DBPassword,
        config.AppConfig.DBHost,
        config.AppConfig.DBPort,
        config.AppConfig.DBName,
    )
    
    // 配置GORM日志
    gormLog := &GORMLogger{
        SlowThreshold: 200 * time.Millisecond, // 慢查询阈值
    }
    
    var err error
    DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
        Logger: gormLog.LogMode(gormlogger.Info),
    })
    if err != nil {
        logger.Fatal("Failed to connect to database: ", err)
    }
    
    // 配置连接池
    sqlDB, err := DB.DB()
    if err != nil {
        logger.Fatal("Failed to get database instance: ", err)
    }
    
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    // Auto migrate schemas
    err = DB.AutoMigrate(
        &models.User{}, 
        &models.Post{}, 
        &models.Comment{},
        &models.Category{},
        &models.Tag{},
        &models.Upload{},
        &models.OperationLog{}, // 添加操作日志表
    )
    if err != nil {
        logger.Fatal("Failed to migrate database: ", err)
    }
    
    logger.Info("Database connected and migrated successfully")
}
```

### 6. 操作日志模型

#### internal/models/operation_log.go

```go
package models

import (
    "time"
)

// OperationLog 操作日志模型
type OperationLog struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    UserID     uint      `gorm:"index" json:"user_id"`
    Username   string    `gorm:"size:50" json:"username"`
    Operation  string    `gorm:"size:50;not null" json:"operation"` // CREATE, UPDATE, DELETE, LOGIN, LOGOUT
    Resource   string    `gorm:"size:100" json:"resource"`          // post, user, category, etc.
    ResourceID uint      `json:"resource_id"`
    Method     string    `gorm:"size:10" json:"method"`             // GET, POST, PUT, DELETE
    Path       string    `gorm:"size:500" json:"path"`
    IP         string    `gorm:"size:50" json:"ip"`
    UserAgent  string    `gorm:"size:500" json:"user_agent"`
    Request    string    `gorm:"type:text" json:"request"`          // 请求参数
    Response   string    `gorm:"type:text" json:"response"`         // 响应结果
    StatusCode int       `json:"status_code"`
    Duration   int64     `json:"duration"`                          // 执行时间(ms)
    Error      string    `gorm:"type:text" json:"error"`
    CreatedAt  time.Time `json:"created_at"`
}

func (OperationLog) TableName() string {
    return "operation_logs"
}
```

### 7. 操作日志中间件

#### internal/middleware/operation_log.go

```go
package middleware

import (
    "blog-backend/internal/database"
    "blog-backend/internal/models"
    "blog-backend/pkg/logger"
    "bytes"
    "encoding/json"
    "io"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/requestid"
)

// OperationLog 操作日志中间件
func OperationLog() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 跳过健康检查和静态资源
        if c.Request.URL.Path == "/health" || 
           c.Request.URL.Path == "/metrics" ||
           len(c.Request.URL.Path) >= 8 && c.Request.URL.Path[:8] == "/uploads" {
            c.Next()
            return
        }
        
        start := time.Now()
        
        // 获取用户信息
        userID, _ := c.Get("user_id")
        username, _ := c.Get("username")
        
        // 读取请求体
        var requestBody []byte
        if c.Request.Body != nil {
            requestBody, _ = io.ReadAll(c.Request.Body)
            c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
        }
        
        // 包装响应写入器以捕获响应
        blw := &bodyLogWriter{
            ResponseWriter: c.Writer,
            body:           bytes.NewBufferString(""),
        }
        c.Writer = blw
        
        // 处理请求
        c.Next()
        
        // 计算耗时
        duration := time.Since(start).Milliseconds()
        
        // 异步保存操作日志（避免影响响应性能）
        go func() {
            // 只记录写操作和重要操作
            method := c.Request.Method
            if method == "GET" && c.Writer.Status() < 400 {
                // GET请求只记录错误和特定资源
                return
            }
            
            // 构建操作日志
            opLog := &models.OperationLog{
                UserID:     getUint(userID),
                Username:   getString(username),
                Operation:  getOperation(method, c.Writer.Status()),
                Resource:   getResource(c.Request.URL.Path),
                Method:     method,
                Path:       c.Request.URL.Path,
                IP:         c.ClientIP(),
                UserAgent:  c.Request.UserAgent(),
                Request:    truncateString(string(requestBody), 1000),
                Response:   truncateString(blw.body.String(), 1000),
                StatusCode: c.Writer.Status(),
                Duration:   duration,
            }
            
            if c.Writer.Status() >= 400 {
                opLog.Error = blw.body.String()
            }
            
            // 保存到数据库
            if err := database.DB.Create(opLog).Error; err != nil {
                logger.Errorf("Failed to save operation log: %v", err)
            }
        }()
    }
}

func getOperation(method string, statusCode int) string {
    if statusCode >= 400 {
        return "ERROR"
    }
    
    switch method {
    case "POST":
        return "CREATE"
    case "PUT", "PATCH":
        return "UPDATE"
    case "DELETE":
        return "DELETE"
    case "GET":
        return "READ"
    default:
        return method
    }
}

func getResource(path string) string {
    // 从路径中提取资源类型
    parts := bytes.Split([]byte(path), []byte("/"))
    for i, part := range parts {
        if string(part) == "api" && i+2 < len(parts) {
            return string(parts[i+2])
        }
    }
    return "unknown"
}

func getUint(v interface{}) uint {
    if v == nil {
        return 0
    }
    if u, ok := v.(uint); ok {
        return u
    }
    return 0
}

func getString(v interface{}) string {
    if v == nil {
        return ""
    }
    if s, ok := v.(string); ok {
        return s
    }
    return ""
}

func truncateString(s string, maxLen int) string {
    if len(s) > maxLen {
        return s[:maxLen] + "..."
    }
    return s
}
```

### 8. 操作日志查询服务

#### internal/service/log_service.go

```go
package service

import (
    "blog-backend/internal/database"
    "blog-backend/internal/models"
    "time"
)

type LogService struct{}

func NewLogService() *LogService {
    return &LogService{}
}

// GetOperationLogs 获取操作日志列表
func (s *LogService) GetOperationLogs(page, pageSize int, filters map[string]interface{}) ([]models.OperationLog, int64, error) {
    var logs []models.OperationLog
    var total int64
    
    query := database.DB.Model(&models.OperationLog{})
    
    // 应用过滤器
    if userID, ok := filters["user_id"]; ok && userID != 0 {
        query = query.Where("user_id = ?", userID)
    }
    if operation, ok := filters["operation"]; ok && operation != "" {
        query = query.Where("operation = ?", operation)
    }
    if resource, ok := filters["resource"]; ok && resource != "" {
        query = query.Where("resource = ?", resource)
    }
    if statusCode, ok := filters["status_code"]; ok && statusCode != 0 {
        query = query.Where("status_code = ?", statusCode)
    }
    if startDate, ok := filters["start_date"]; ok && startDate != nil {
        query = query.Where("created_at >= ?", startDate)
    }
    if endDate, ok := filters["end_date"]; ok && endDate != nil {
        query = query.Where("created_at <= ?", endDate)
    }
    
    query.Count(&total)
    
    offset := (page - 1) * pageSize
    err := query.Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&logs).Error
    
    return logs, total, err
}

// GetErrorLogs 获取错误日志
func (s *LogService) GetErrorLogs(page, pageSize int) ([]models.OperationLog, int64, error) {
    var logs []models.OperationLog
    var total int64
    
    query := database.DB.Model(&models.OperationLog{}).
        Where("status_code >= 400 OR error != ''")
    
    query.Count(&total)
    
    offset := (page - 1) * pageSize
    err := query.Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&logs).Error
    
    return logs, total, err
}

// CleanOldLogs 清理旧日志
func (s *LogService) CleanOldLogs(days int) error {
    cutoffTime := time.Now().AddDate(0, 0, -days)
    return database.DB.Where("created_at < ?", cutoffTime).Delete(&models.OperationLog{}).Error
}

// GetStatistics 获取日志统计
func (s *LogService) GetStatistics() (map[string]interface{}, error) {
    stats := make(map[string]interface{})
    
    // 总请求数
    var totalCount int64
    database.DB.Model(&models.OperationLog{}).Count(&totalCount)
    stats["total_requests"] = totalCount
    
    // 错误数
    var errorCount int64
    database.DB.Model(&models.OperationLog{}).Where("status_code >= 400").Count(&errorCount)
    stats["error_requests"] = errorCount
    
    // 各操作类型统计
    var operations []struct {
        Operation string
        Count     int64
    }
    database.DB.Model(&models.OperationLog{}).
        Select("operation, count(*) as count").
        Group("operation").
        Scan(&operations)
    stats["operations"] = operations
    
    // 今日请求数
    today := time.Now().Format("2006-01-02")
    var todayCount int64
    database.DB.Model(&models.OperationLog{}).
        Where("DATE(created_at) = ?", today).
        Count(&todayCount)
    stats["today_requests"] = todayCount
    
    return stats, nil
}
```

### 9. 操作日志处理器

#### internal/handler/log_handler.go

```go
package handler

import (
    "blog-backend/internal/service"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
)

type LogHandler struct {
    logService *service.LogService
}

func NewLogHandler() *LogHandler {
    return &LogHandler{
        logService: service.NewLogService(),
    }
}

// GetOperationLogs 获取操作日志列表（仅管理员）
func (h *LogHandler) GetOperationLogs(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    filters := make(map[string]interface{})
    
    if userID := c.Query("user_id"); userID != "" {
        id, _ := strconv.ParseUint(userID, 10, 32)
        filters["user_id"] = uint(id)
    }
    if operation := c.Query("operation"); operation != "" {
        filters["operation"] = operation
    }
    if resource := c.Query("resource"); resource != "" {
        filters["resource"] = resource
    }
    if statusCode := c.Query("status_code"); statusCode != "" {
        code, _ := strconv.Atoi(statusCode)
        filters["status_code"] = code
    }
    if startDate := c.Query("start_date"); startDate != "" {
        t, _ := time.Parse("2006-01-02", startDate)
        filters["start_date"] = t
    }
    if endDate := c.Query("end_date"); endDate != "" {
        t, _ := time.Parse("2006-01-02", endDate)
        filters["end_date"] = t.Add(24 * time.Hour)
    }
    
    logs, total, err := h.logService.GetOperationLogs(page, pageSize, filters)
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "logs":      logs,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}

// GetErrorLogs 获取错误日志
func (h *LogHandler) GetErrorLogs(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
    
    logs, total, err := h.logService.GetErrorLogs(page, pageSize)
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, gin.H{
        "logs":      logs,
        "total":     total,
        "page":      page,
        "page_size": pageSize,
    })
}

// GetStatistics 获取日志统计
func (h *LogHandler) GetStatistics(c *gin.Context) {
    stats, err := h.logService.GetStatistics()
    if err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    Success(c, stats)
}

// CleanLogs 清理旧日志
func (h *LogHandler) CleanLogs(c *gin.Context) {
    days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
    
    if err := h.logService.CleanOldLogs(days); err != nil {
        InternalServerError(c, err.Error())
        return
    }
    
    SuccessWithMessage(c, "Logs cleaned successfully", nil)
}
```

### 10. 更新 .env 配置

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=blog

# JWT Configuration
JWT_SECRET=your-secret-key
JWT_EXPIRE_HOURS=24

# Server Configuration
PORT=8080

# Log Configuration
LOG_LEVEL=info          # debug, info, warn, error
LOG_PATH=logs           # 日志文件路径
LOG_MAX_AGE=30          # 日志保留天数
LOG_ROTATION_TIME=24    # 日志轮转时间（小时）
```

### 11. 更新路由配置

#### internal/router/router.go (更新)

```go
package router

import (
    "blog-backend/internal/handler"
    "blog-backend/internal/middleware"
    "blog-backend/internal/service"
    "blog-backend/pkg/logger"
    
    "github.com/gin-contrib/requestid"
    "github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
    // 设置Gin模式
    gin.SetMode(gin.ReleaseMode)
    
    router := gin.New()
    
    // 添加中间件
    router.Use(middleware.Recovery())      // Panic恢复
    router.Use(requestid.New())            // 请求ID
    router.Use(middleware.Logger())        // 请求日志
    router.Use(middleware.CORS())          // CORS
    router.Use(middleware.OperationLog())  // 操作日志
    
    // 静态文件服务
    service.ServeStatic(router)
    
    // 健康检查
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    // API路由
    api := router.Group("/api/v1")
    {
        authHandler := handler.NewAuthHandler()
        postHandler := handler.NewPostHandler()
        categoryHandler := handler.NewCategoryHandler()
        tagHandler := handler.NewTagHandler()
        commentHandler := handler.NewCommentHandler()
        uploadHandler := handler.NewUploadHandler()
        
        // 公开路由
        api.POST("/auth/register", authHandler.Register)
        api.POST("/auth/login", authHandler.Login)
        api.GET("/posts", postHandler.List)
        api.GET("/posts/slug/:slug", postHandler.GetBySlug)
        api.GET("/posts/:id", postHandler.GetByID)
        api.GET("/categories", categoryHandler.GetAll)
        api.GET("/tags", tagHandler.GetAll)
        api.GET("/posts/:post_id/comments", commentHandler.GetByPost)
        
        // 需要认证的路由
        protected := api.Group("/")
        protected.Use(middleware.Auth())
        {
            protected.POST("/posts", postHandler.Create)
            protected.PUT("/posts/:id", postHandler.Update)
            protected.DELETE("/posts/:id", postHandler.Delete)
            protected.POST("/comments", commentHandler.Create)
            protected.POST("/upload", uploadHandler.Upload)
            protected.GET("/uploads", uploadHandler.GetUserUploads)
            protected.DELETE("/upload/:id", uploadHandler.Delete)
        }
        
        // 管理员路由
        admin := api.Group("/admin")
        admin.Use(middleware.Auth(), middleware.AdminOnly())
        {
            // 内容管理
            admin.GET("/posts/all", postHandler.List)
            admin.POST("/categories", categoryHandler.Create)
            admin.PUT("/categories/:id", categoryHandler.Update)
            admin.DELETE("/categories/:id", categoryHandler.Delete)
            admin.POST("/tags", tagHandler.Create)
            admin.DELETE("/tags/:id", tagHandler.Delete)
            
            // 评论管理
            admin.GET("/comments/pending", commentHandler.GetPending)
            admin.POST("/comments/:id/approve", commentHandler.Approve)
            admin.POST("/comments/:id/reject", commentHandler.Reject)
            
            // 日志管理
            logHandler := handler.NewLogHandler()
            admin.GET("/logs/operations", logHandler.GetOperationLogs)
            admin.GET("/logs/errors", logHandler.GetErrorLogs)
            admin.GET("/logs/statistics", logHandler.GetStatistics)
            admin.DELETE("/logs/clean", logHandler.CleanLogs)
        }
    }
    
    return router
}
```

### 12. 更新 main.go

#### cmd/server/main.go (更新)

```go
package main

import (
    "blog-backend/internal/config"
    "blog-backend/internal/database"
    "blog-backend/internal/router"
    "blog-backend/pkg/logger"
    "fmt"
)

func main() {
    // 初始化日志
    if err := logger.InitLogger(); err != nil {
        panic(fmt.Sprintf("Failed to initialize logger: %v", err))
    }
    
    logger.Info("Starting blog backend server...")
    
    // 加载配置
    config.LoadConfig()
    logger.Infof("Configuration loaded: env=%s, log_level=%s", 
        gin.Mode(), config.AppConfig.LogLevel)
    
    // 初始化数据库
    database.InitDB()
    logger.Info("Database initialized")
    
    // 设置路由
    r := router.SetupRouter()
    
    // 启动服务器
    addr := fmt.Sprintf(":%s", config.AppConfig.Port)
    logger.Infof("Server starting on %s", addr)
    
    if err := r.Run(addr); err != nil {
        logger.Fatalf("Failed to start server: %v", err)
    }
}
```

### 13. 使用示例

#### 在代码中使用日志

```go
// 在 Service 中使用
func (s *PostService) Create(post *models.Post) error {
    logger.WithFields(map[string]interface{}{
        "title": post.Title,
        "user_id": post.UserID,
    }).Info("Creating new post")
    
    if err := s.repo.Create(post); err != nil {
        logger.WithField("error", err).Error("Failed to create post")
        return err
    }
    
    logger.WithField("post_id", post.ID).Info("Post created successfully")
    return nil
}

// 在 Handler 中使用
func (h *PostHandler) Create(c *gin.Context) {
    reqID := requestid.Get(c)
    logger.WithField("request_id", reqID).Debug("Processing create post request")
    
    // ... 处理逻辑
    
    logger.WithField("request_id", reqID).Info("Post created")
}
```

### 14. 日志查询 API

```bash
# 查询操作日志（管理员）
curl -X GET "http://localhost:8080/api/v1/admin/logs/operations?page=1&page_size=20&operation=CREATE" \
  -H "Authorization: Bearer ADMIN_TOKEN"

# 查询错误日志
curl -X GET "http://localhost:8080/api/v1/admin/logs/errors?page=1&page_size=20" \
  -H "Authorization: Bearer ADMIN_TOKEN"

# 获取日志统计
curl -X GET "http://localhost:8080/api/v1/admin/logs/statistics" \
  -H "Authorization: Bearer ADMIN_TOKEN"

# 清理30天前的日志
curl -X DELETE "http://localhost:8080/api/v1/admin/logs/clean?days=30" \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

### 15. 日志文件结构

```
logs/
├── blog.log              # 当前日志（软链接）
├── blog.log.20250101     # 按日期轮转的日志
├── blog.log.20250102
└── blog.log.20250103
```

这个完整的日志系统提供了：

✅ **请求日志**：记录所有HTTP请求的详细信息
✅ **错误日志**：捕获和记录错误信息
✅ **操作日志**：记录用户的CRUD操作
✅ **数据库日志**：记录SQL查询和慢查询
✅ **Panic恢复**：捕获panic并记录堆栈信息
✅ **日志轮转**：自动按日期轮转和清理
✅ **结构化日志**：JSON格式，便于分析
✅ **日志统计**：提供日志分析API
✅ **性能监控**：记录请求耗时和慢查询

你可以根据实际需求调整日志级别、保留时间和轮转策略。
