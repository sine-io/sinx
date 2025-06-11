
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
