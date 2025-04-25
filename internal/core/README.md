# `/core`

The business logic layer of the project, divided into packages by business domains.

## 目录说明

`/core` 目录包含项目的核心业务逻辑，按照业务领域（Domain）进行划分。每个子目录代表一个独立的业务领域。

## 设计原则

1. **领域驱动设计** - 每个子目录应专注于一个特定业务领域
2. **高内聚低耦合** - 业务逻辑应尽量独立，减少跨域依赖
3. **依赖反转** - 使用接口隔离具体实现，方便测试和扩展

## 目录结构示例

```
/core
  /user         - 用户相关业务逻辑
    /entity     - 用户领域实体
    /repository - 用户数据访问接口
    /service    - 用户服务实现
    /dto        - 数据传输对象
  /order        - 订单相关业务逻辑
  /product      - 产品相关业务逻辑
  ...
```

## 新增业务领域步骤

1. 在 `/core` 下创建新的目录，以业务领域命名
2. 定义该领域的核心实体和值对象
3. 定义领域服务接口和实现
4. 定义仓储接口（具体实现放在 `/infra` 层）

## 示例代码

```go
// /core/user/entity/user.go
package entity

type User struct {
    ID       string
    Username string
    Email    string
    // 其他属性
}

// /core/user/repository/repository.go
package repository

import "project/internal/core/user/entity"

type UserRepository interface {
    FindByID(id string) (*entity.User, error)
    Save(user *entity.User) error
    // 其他方法
}
```
