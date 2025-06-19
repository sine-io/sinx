
scoop bucket add extras
scoop install extras/protobuf

git config --global http.proxy <http://ip:port>
git config --global https.proxy <http://ip:port>

git config --global --unset http.proxy
git config --global --unset https.proxy

git config --global --get http.proxy
git config --global --get https.proxy

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

Agent 模块
    集群管理（Serf、Raft）
    节点间通信
    服务发现与健康检查
    领导选举与领导者监控
Scheduler 模块
    纯粹的任务调度功能
    Cron 表达式解析与管理
    调度时间计算
    任务执行委派给执行器
Job 模块
    任务定义与状态
    任务间的依赖关系
    任务执行器选择与配置
    任务验证逻辑

# 模块分析

## 模块划分

1. **Agent 模块**：系统的核心组件
2. **Store 模块**：数据存储层
3. **Job 模块**：任务定义和管理
4. **Scheduler 模块**：调度系统
5. **RPC/通信模块**：节点间通信
6. **Execution 模块**：执行实例的管理
7. **Raft 模块**：一致性和领导选举

## 模块调用关系图

```text
                +-------------+
                |    Agent    |
                +------+------+
                       |
       +---------------+---------------+
       |               |               |
+------v------+ +------v------+ +------v------+
|    Store    | |  Scheduler  | |    Raft     |
+------+------+ +------+------+ +------+------+
       |               |               |
       |        +------v------+        |
       +------->|     Job     |<-------+
                +------+------+
                       |
                +------v------+
                | Execution   |
                +------+------+
                       |
                +------v------+
                |  RPC/gRPC   |
                +-------------+
```

## 各模块责任

### 1. Agent 模块 (agent.go, serf.go)

- **主要责任**：作为系统核心，协调所有其他组件的工作
- **功能**：
  - 节点管理和集群组织
  - 协调领导选举过程
  - 管理节点间通信
  - 处理节点加入/离开事件
  - 启动和管理其他模块

### 2. Store 模块 (store.go)

- **主要责任**：提供持久化存储功能
- **功能**：
  - 存储任务定义和配置
  - 存储执行记录和结果
  - 提供查询和检索能力
  - 使用 BuntDB 作为底层存储引擎

### 3. Job 模块 (job.go)

- **主要责任**：定义和管理任务
- **功能**：
  - 任务元数据管理
  - 任务依赖关系处理
  - 父子任务关系管理
  - 任务验证

### 4. Scheduler 模块 (scheduler.go)

- **主要责任**：负责任务的调度和执行
- **功能**：
  - 基于 cron 表达式的任务调度
  - 管理任务执行时间
  - 触发任务运行

### 5. Execution 模块 (execution.go)

- **主要责任**：管理任务的执行实例
- **功能**：
  - 执行状态跟踪
  - 执行结果存储
  - 重试逻辑实现

### 6. RPC/通信模块 (grpc.go, grpc_client.go, server_lookup.go)

- **主要责任**：提供节点间通信机制
- **功能**：
  - 实现 gRPC 服务器和客户端
  - 任务执行命令的传递
  - 集群状态同步

### 7. Raft 模块 (raft_grpc.go, fsm.go)

- **主要责任**：提供分布式一致性和领导选举
- **功能**：
  - 实现领导选举
  - 保证集群状态一致性
  - 日志复制
  - 实现有限状态机 (FSM) 处理日志应用

## 整体架构特点

采用分布式架构，使用 Raft 算法保证一致性，使用 Serf 进行成员管理。其核心是 Agent 模块，协调其他所有模块工作。任务执行分散在集群的不同节点上，通过 RPC 进行通信，数据通过 Store 模块持久化。这种设计能够实现高可用、容错和可扩展的任务调度系统。
