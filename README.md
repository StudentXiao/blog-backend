

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



## 接口文档生成
swag init --dir ./ --parseDependency --parseInternal --output ./docs
