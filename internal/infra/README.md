# `/infra`

Contains the infrastructure layer of the project, divided into packages by technologies or infra domains.

## 目录说明

`/infra` 目录包含项目的基础设施层，按照技术或基础设施域进行划分。这一层负责实现与外部系统、框架、数据库等的交互。

## 设计原则

1. **关注点分离** - 将技术实现与业务逻辑分离
2. **适配器模式** - 实现 `/core` 层定义的接口，适配外部技术
3. **依赖注入** - 在启动时注入具体实现，不在业务逻辑中直接依赖具体技术

## 目录结构示例

```
/infra
  /db           - 数据库相关实现
    /mysql      - MySQL 实现
    /redis      - Redis 缓存实现
  /http         - HTTP 相关实现
    /rest       - RESTful API 实现
    /middleware - HTTP 中间件
  /grpc         - gRPC 相关实现
  /queue        - 消息队列实现
    /kafka      - Kafka 实现
  /auth         - 认证授权实现
  /storage      - 文件存储实现
  ...
```

## 实现仓储接口示例

下面是将 `/core` 层定义的仓储接口在 `/infra` 层实现的示例：

```go
// /infra/db/mysql/user_repository.go
package mysql

import (
    "database/sql"
    "project/internal/core/user/entity"
    "project/internal/core/user/repository"
)

type MySQLUserRepository struct {
    db *sql.DB
}

// 确保实现了接口
var _ repository.UserRepository = (*MySQLUserRepository)(nil)

func NewMySQLUserRepository(db *sql.DB) *MySQLUserRepository {
    return &MySQLUserRepository{db: db}
}

func (r *MySQLUserRepository) FindByID(id string) (*entity.User, error) {
    // 实现从 MySQL 数据库查询用户的逻辑
    // ...
    return &entity.User{ID: id}, nil
}

func (r *MySQLUserRepository) Save(user *entity.User) error {
    // 实现保存用户到 MySQL 数据库的逻辑
    // ...
    return nil
}
```
