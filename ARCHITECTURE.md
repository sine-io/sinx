# Backend 架构说明

> 本文档描述 `backend/` 目录下后端服务的整体分层、初始化流程、请求生命周期、核心组件以及扩展指引，便于新成员快速理解与贡献。

## 目录结构宏观视图

```text
backend/
  main.go                # 进程入口：环境加载 -> 日志 -> Application.Init -> 启动 HTTP / 反向代理
  api/                   # 传输层（Transport）：路由、DTO、Middleware
  application/           # 应用服务层（Use Case orchestration）
  domain/                # 领域模型与领域服务 (纯业务规则，不依赖框架)
  crossdomain/           # 跨领域访问适配 (对外暴露统一接口 SetDefaultSVC)
  infra/                 # 基础设施实现（DB / Cache / MQ / Storage / Model / Vector / 解析 / OCR ...）
  pkg/                   # 通用库（日志、错误、工具、缓存上下文等）
  types/                 # 通用类型、常量、错误码、DDL
  conf/                  # 静态配置 (模型、插件、工作流模板等)
  internal/              # 测试、Mock、仅内部使用
```

### 分层职责对照

| 层级 | 主要目录 | 职责 | 依赖方向 |
|------|----------|------|----------|
| Transport | `api/` | HTTP 适配、参数解析、鉴权、序列化 | 向下依赖 application |
| Application | `application/` | 用例编排、事务边界、调用多个领域服务 | 依赖 domain + infra contract |
| Domain | `domain/` | 业务核心：实体、值对象、领域服务 | 仅依赖基础工具(pkg) |
| Cross Domain | `crossdomain/` | 将多个领域能力以统一接口暴露（避免循环依赖） | 依赖 domain/application |
| Infrastructure | `infra/impl` | 具体技术实现（MySQL、Redis、ES、向量库、OSS、MQ、模型等） | 可被 application/base 初始化装配 |
| Shared Lib | `pkg/`, `types/` | 公共工具、错误码、常量 | 被所有层引用 |

依赖规则：上层只能依赖更低层抽象，不反向；禁止 domain 直接 import 具体 infra 实现。

## 启动与初始化流程

`main.go` 启动顺序（请勿随意改动）：

1. `setCrashOutput()` 注册 panic/crash 输出文件
2. `loadEnv()` 加载 `.env[.ENV]` 环境变量
3. `setLogLevel()` 配置日志级别
4. `application.Init(ctx)`：
   - `appinfra.Init` 构建 `AppDependencies`
   - 初始化内部分层服务 (basic -> primary -> complex)
   - 为 `crossdomain/contract/*` 设置默认实现（`SetDefaultSVC`）
5. 异步启动 Minio / TOS / S3 代理（可选）
6. 构造并启动 Hertz HTTP Server（SSL 可选）

### AppDependencies (Infra 聚合)

`application/base/appinfra/AppDependencies` 封装所有底层资源：

- DB (`gorm.DB`)
- Cache (`redis.Cmdable`)
- ID 生成器 (`idgen` 基于 redis)
- ES Client (全文检索)
- Vector Store 管理器数组 (向量 / 混合检索 Milvus / VikingDB / OceanBase)
- EventBus Producer（Resource / App / Knowledge 三类 Topic）
- Storage(TOS / S3 兼容层)、ImageX、OCR、Parser、Reranker、Rewriter、NL2SQL
- Chat Model 管理 (`ModelMgr` + 内置工作流模型)
- CodeRunner (直接执行或 Sandbox)

所有上层服务通过传入组合结构体（ServiceComponents）受控访问这些依赖，保证解耦与可测性。

### 分阶段服务装配

```text
basicServices      -> 只依赖 infra (user / prompt / template / upload / modelmgr / connector / openauth)
primaryServices    -> 依赖 basic (plugin / memory / knowledge / workflow / shortcut)
complexServices    -> 依赖 primary (singleAgent / app / search / conversation)
```

逐层构造是为了：

- 避免循环依赖
- 控制初始化失败链路（上层失败不污染已就绪下层）
- 提高新增模块时的定位与接入清晰度

### CrossDomain 适配

`crossdomain/contract/*` 定义对外统一接口（面向其他模块 / 未来微服务拆分的“防腐层”），入口在 `application.Init` 最后阶段：

```go
crossconnector.SetDefaultSVC(connectorImpl.InitDomainService(...))
... // 其余资源同理
```

这样：

- `application` 内部仍可直接拿具体服务
- 外部只见 *contract* 接口，减少耦合

## 请求生命周期 (HTTP)

1. Hertz 接收请求
2. Middleware 链按注册顺序执行：
   - `ContextCacheMW` 放置全局上下文缓存键值
   - `RequestInspectorMW` 判定请求类型（Web API / OpenAPI / 静态资源）
   - `SetHostMW` / `SetLogIDMW` 注入 Host 与日志 Trace ID
   - `cors` 处理跨域
   - `AccessLogMW` 记录访问日志
   - `OpenapiAuthMW`：若匹配需鉴权路径 -> Bearer Token -> MD5 -> 校验权限 -> 设置上下文
   - `SessionAuthMW`：Web 登录态校验（与 OpenAPI 鉴权互斥）
   - `I18nMW`：根据用户 / 语言标头设置本地化
3. Router (由代码生成工具生成) 匹配 handler
4. Handler -> Application Service -> (Domain 规则 / 事件 / 存储) -> 返回 DTO
5. 错误经统一 `errno` -> `errorx` 映射为结构化响应

### 认证类型分流

`RequestInspectorMW` 设置 `RequestAuthType`：

- Web API: 需要 Session 鉴权
- OpenAPI: 需 Header `Authorization: Bearer <api_key>` 校验
- Static: 允许匿名访问

## 事件总线 (EventBus)

| Topic | 用途 | Producer 来源 |
|-------|------|---------------|
| Resource | 资源变更（知识库/插件/变量等） | `infra.ResourceEventProducer` |
| App | 项目 / 应用级事件 | `infra.AppEventProducer` |
| Knowledge | 知识库异步处理 | `infra.KnowledgeEventProducer` |

`search.ResourceEventBus` / `search.ProjectEventBus` 封装 producer，用于触发重建索引、刷新缓存等。

## 知识检索与嵌入 (Knowledge / Search)

1. ParserManager：抽取文本块 (支持 `builtin` / `paddleocr`)
2. OCR：图片 -> 文本 (ve / paddle / 可选)
3. Embedding：根据 `EMBEDDING_TYPE` 选择 OpenAI / Ark / Ollama / Gemini / HTTP 定制
4. Vector Store：`VECTOR_STORE_TYPE` = milvus / vikingdb / oceanbase
5. Rerank：rrf / vikingdb 加权融合
6. Rewriter：将对话消息转为检索 Query (messages2query)
7. NL2SQL：自然语言转 SQL (数据查询场景)

## 错误处理与错误码

`types/errno/*` 中每个子领域预留唯一 code 段，例如：

- Plugin: 109000000 ~ 109999999
- Workflow / User / Memory ... 各有独立文件

注册时调用 `code.Register` 绑定：消息模板、是否影响稳定性等。Application / Domain 中通过 `errorx.New(errno.ErrPluginInvalidParamCode, ...)` 构造统一响应。

## 配置与环境变量分类

| 类别 | 示例变量 | 说明 |
|------|----------|------|
| 运行 | `LISTEN_ADDR` / `MAX_REQUEST_BODY_SIZE` / `APP_ENV` | 端口、环境、请求体大小 |
| 日志 | `LOG_LEVEL` | trace/debug/info 等 |
| 存储 | `MYSQL_*`, `REDIS_*`, `TOS_*` | 数据库、缓存、对象存储 |
| 检索 | `VECTOR_STORE_TYPE`, `EMBEDDING_TYPE`, `RERANK_TYPE` | 检索策略与嵌入模型 |
| 模型 | `OPENAI_*`, `ARK_*`, `OLLAMA_*`, `GEMINI_*` | 多模型后端选择 |
| 安全 | `UseSSL`, `SSLCertFile`, `SSLKeyFile` | HTTPS 支持 |
| OCR/解析 | `OCRType`, `ParserType`, `PPOCRAPIURL`, `PPStructureAPIURL` | 文档解析链路 |
| 代码执行 | `CodeRunnerType` 等 | 直接 or sandbox 隔离 |

（建议为新增模块归类并更新此表）

## 典型用例数据流

### A. 会话对话 (Chat)

HTTP -> Middleware 鉴权 -> ConversationApplication -> SingleAgent -> (模型调用 / 插件 / 变量 / 工作流) -> 持久化消息 + 事件 -> 返回流式或整包结果。

### B. 知识库导入

上传文件 -> UploadService (对象存储) -> KnowledgeApplication 发起解析任务 -> ParserManager + OCR -> 拆分 & 嵌入 -> Vector Store & ES 写入 -> 事件触发可选增量索引 -> 可检索。

### C. 工作流运行

WorkflowService 校验/编排 -> 触发节点执行 (可能调用 CodeRunner / 插件 / 知识库检索 / 模型推理) -> Checkpoint Redis 持久化状态 -> 失败可恢复 / 流式输出。

## 并发与可靠性

- `safego.Go(ctx, fn)`：集中管理 goroutine，防止 panic 失联
- 事件最终一致性：写库后投递 MQ；搜索索引异步更新
- Redis 既用于缓存也用于 ID 发号与工作流 checkpoint

## 扩展实践指南

### 新增一个领域服务 (Domain Service)

1. 在 `domain/<new_domain>/` 定义实体、仓储接口、领域服务
2. 在 `infra/impl/<component>` 实现仓储（若需要新存储）
3. 在 `application/<new_domain>/` 编写 ApplicationService：对外用例 + 组合领域服务
4. 在 `application/application.go`：
   - 归类到 basic / primary / complex 层
   - 通过 `crossdomain` 暴露（可选）
5. 在 `api/router` 添加路由 & Handler
6. 在 `types/errno` 申请 code 段并注册错误码

### 新增一个 HTTP API

1. 在 `api/router/` 或生成工具模板添加路由
2. DTO / Request 校验
3. 调用对应 Application Service 方法
4. 统一返回结构 & 错误码

### 新增向量库支持

1. 在 `infra/impl/document/searchstore/<engine>` 实现 `Manager`
2. 在 `appinfra.getVectorStore` 分支中增加 case
3. 增加相关 env 变量说明

### 选择放置目录的原则

| 需求 | 放置位置 |
|------|----------|
| 业务规则/聚合根 | `domain/` |
| 跨多个领域聚合用例 | `application/` |
| HTTP、鉴权、中间件 | `api/` |
| 资源持久化技术细节 | `infra/impl/...` |
| 工具 / 无业务状态 | `pkg/` |
| 常量、枚举、错误码 | `types/` |

## 代码风格与约定

- 初始化函数统一返回 `(T, error)`，错误向上抛出到 `application.Init`
- 不在 `domain` 直接引用具体第三方库（除非为纯值对象处理）
- 中间件顺序敏感：`ContextCacheMW` / `RequestInspectorMW` 必须最前
- 错误码一旦对外公开不可复用或删除，只能新增

## 测试建议

- Domain 层：纯单元测试（mock 仓储接口）
- Application 层：用内存 / docker compose 启动依赖的集成测试
- API 层：黑盒回归（关键鉴权 / 并发场景）
- 针对事件处理/检索链路可使用“短路”配置（如内置 Parser / rrf Rerank）缩短反馈

## 常见问题 (FAQ)

Q: 为什么要有 `crossdomain` 层？
A: 解决“多个领域服务互相调用”导致的循环依赖，并为未来微服务拆分预留隔离层。

Q: 新增模块放哪一层？
A: 是否包含业务规则？是 -> domain；只是组合多个现有领域服务构成用例 -> application；仅是技术实现 -> infra。

Q: 如何调试向量检索？
A: 确认 `VECTOR_STORE_TYPE` 与对应 env；查看初始化日志；开启 `LOG_LEVEL=debug`，必要时在 `SearchApplication` 增加临时调试输出。

## 术语表

| 术语 | 解释 |
|------|------|
| Application Service | 用例编排层，协调多个领域服务完成一项业务动作 |
| Domain Service | 纯业务逻辑封装，围绕实体 / 值对象进行操作 |
| CrossDomain Service | 统一访问入口与适配层，避免循环依赖 |
| EventBus | RocketMQ 封装，用于资源 / 项目 / 知识相关异步事件 |
| Vector Store | 用于语义检索的向量数据库或混合检索引擎 |
| Rewriter | 将多轮对话上下文转为检索 Query 的组件 |
| Reranker | 对初步检索结果进行重排提升相关度 |
| CodeRunner | 运行工作流节点代码（沙箱或直跑） |

## 后续改进建议 (Roadmap 提示)

- 将 `crossdomain` 服务按模块逐步抽象为 gRPC / HTTP Service Interface（支持独立进程化）
- 增加统一链路追踪 (OpenTelemetry) 采集模型调用与检索耗时
- 工作流节点执行状态事件化，支持可视化 DAG 追踪
- 统一资源事件与业务审计日志

---

文档版本：v1.0  (生成日期：2025-09-05)
