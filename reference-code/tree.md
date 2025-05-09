# 目录树

## 顶层目录说明

1. api: 认证、上下文、存储相关接口和实现
2. config: 配置模型、解析、序列化与通用工具
3. controller: 控制器核心逻辑、导出、加载、调度、服务、任务单元
4. controller-web: 控制器Web前端与接口
5. core: 核心基准测试、模型、协议、服务、工具
6. core-web: 核心Web异常处理
7. driver: 驱动执行器、生成器、操作器、服务、工具
8. driver-web: 驱动Web前端与接口
9. sineio: SineIO扩展存储协议实现

+---api
|   # 认证、上下文、存储API
|   +---auth           # 认证相关接口与异常
|   |       AuthAPI.java                    # 认证API接口
|   |       AuthAPIFactory.java             # 认证API工厂类
|   |       AuthAPIService.java             # 认证API服务类
|   |       AuthBadException.java           # 认证错误异常
|   |       AuthConstants.java              # 认证常量
|   |       AuthException.java              # 认证通用异常
|   |       AuthInterruptedException.java   # 认证中断异常
|   |       AuthTimeoutException.java       # 认证超时异常
|   |       NoneAuth.java                   # 无认证实现
|   |
|   +---context        # 上下文对象定义
|   |       AuthContext.java                # 认证上下文
|   |       Context.java                    # 通用上下文接口
|   |       DefaultAuthContext.java         # 默认认证上下文实现
|   |
|   \---storage        # 存储API及异常
|           NoneStorage.java                # 无存储实现
|           StorageAPI.java                 # 存储API接口
|           StorageAPIFactory.java          # 存储API工厂类
|           StorageAPIService.java          # 存储API服务类
|           StorageConstants.java           # 存储常量
|           StorageException.java           # 存储通用异常
|           StorageInterruptedException.java# 存储中断异常
|           StorageTimeoutException.java    # 存储超时异常
|
+---config
|   # 配置模型、解析、序列化
|   |   Auth.java                         # 认证配置模型
|   |   Config.java                       # 顶层配置模型
|   |   ConfigConstants.java              # 配置常量
|   |   ConfigException.java              # 配置异常
|   |   Mission.java                      # 任务配置模型
|   |   MissionResolver.java              # 任务配置解析器接口
|   |   MissionWriter.java                # 任务配置写入器接口
|   |   Operation.java                    # 操作配置模型
|   |   Stage.java                        # 阶段配置模型
|   |   Storage.java                      # 存储配置模型
|   |   Work.java                         # 工作单元配置模型
|   |   Workflow.java                     # 工作流配置模型
|   |   Workload.java                     # 工作负载配置模型
|   |   WorkloadResolver.java             # 工作负载配置解析器接口
|   |   WorkloadWriter.java               # 工作负载配置写入器接口
|   |   XmlConfig.java                    # XML配置基类
|   |
|   +---castor         # Castor工具与映射文件
|   |       auth-mapping.xml              # 认证映射文件
|   |       CastorConfigBase.java         # Castor配置基类
|   |       CastorConfigResolver.java     # Castor配置解析器基类
|   |       CastorConfigTools.java        # Castor配置工具类
|   |       CastorConfigWriter.java       # Castor配置写入器基类
|   |       CastorMappings.java           # Castor映射管理
|   |       CastorMissionResolver.java    # Castor任务配置解析器
|   |       CastorMissionWriter.java      # Castor任务配置写入器
|   |       CastorWorkloadResolver.java   # Castor工作负载配置解析器
|   |       CastorWorkloadWriter.java     # Castor工作负载配置写入器
|   |       mission-mapping.xml           # 任务映射文件
|   |       operation-mapping.xml         # 操作映射文件
|   |       stage-mapping.xml             # 阶段映射文件
|   |       storage-mapping.xml           # 存储映射文件
|   |       work-mapping.xml              # 工作单元映射文件
|   |       workflow-mapping.xml          # 工作流映射文件
|   |       workload-mapping.xml          # 工作负载映射文件
|   |
|   \---common         # 通用配置工具
|           ConfigUtils.java              # 配置工具类
|           COSBConfigApator.java         # COSBench配置适配器 (拼写可能为Adaptor)
|           INIConfigParser.java          # INI格式配置解析器
|           KVConfigParser.java           # Key-Value格式配置解析器
|
+---controller
|   # 控制器核心逻辑
|   +---archiver       # 工作负载归档器
|   |       SimpleWorkloadArchiver.java   # 简单工作负载归档器实现
|   |       WorkloadArchiver.java         # 工作负载归档器接口
|   |
|   +---exporter       # 数据导出器
|   |       AbstractAllDriversExporter.java   # 抽象所有驱动导出器基类
|   |       AbstractLatencyExporter.java      # 抽象延迟导出器基类
|   |       AbstractMatrixExporter.java       # 抽象矩阵导出器基类
|   |       AbstractRunExporter.java          # 抽象运行信息导出器基类
|   |       AbstractStageExporter.java        # 抽象阶段信息导出器基类
|   |       AbstractStageExtraExporter.java   # 抽象阶段附加信息导出器基类
|   |       AbstractTaskExporter.java         # 抽象任务信息导出器基类
|   |       AbstractWorkerExporter.java       # 抽象工作节点导出器基类
|   |       AbstractWorkloadExporter.java     # 抽象工作负载导出器基类
|   |       AllDriversExporter.java           # 所有驱动信息导出器接口
|   |       CSVAllDriversExporter.java        # CSV格式所有驱动信息导出器
|   |       CSVLatencyExporter.java           # CSV格式延迟数据导出器
|   |       CSVMatrixExporter.java            # CSV格式性能矩阵导出器
|   |       CSVRunExporter.java               # CSV格式运行信息导出器
|   |       CSVStageExporter.java             # CSV格式阶段信息导出器
|   |       CSVStageExtraExporter.java        # CSV格式阶段附加信息导出器
|   |       CSVTaskExporter.java              # CSV格式任务信息导出器
|   |       CSVWorkerExporter.java            # CSV格式工作节点信息导出器
|   |       CSVWorkloadExporter.java          # CSV格式工作负载信息导出器
|   |       Exporters.java                    # 导出器工具类/工厂类
|   |       ExportException.java              # 导出异常类
|   |       Formats.java                      # 导出格式定义
|   |       LatencyExporter.java              # 延迟数据导出器接口
|   |       LogExporter.java                  # 日志导出器接口
|   |       MatrixExporter.java               # 性能矩阵导出器接口
|   |       RunExporter.java                  # 运行信息导出器接口
|   |       ScriptsLogExporter.java           # 脚本日志导出器
|   |       SimpleLogExporter.java            # 简单日志导出器
|   |       StageExporter.java                # 阶段信息导出器接口
|   |       StageExtraExporter.java           # 阶段附加信息导出器接口
|   |       TaskExporter.java                 # 任务信息导出器接口
|   |       WorkerExporter.java               # 工作节点信息导出器接口
|   |       WorkloadExporter.java             # 工作负载信息导出器接口
|   |
|   +---loader         # 数据加载器
|   |       AbstractAllDriversFileLoader.java # 抽象所有驱动文件加载器基类
|   |       AbstractRunLoader.java            # 抽象运行信息加载器基类
|   |       AbstractSnapshotLoader.java       # 抽象快照加载器基类
|   |       AbstractTaskInfoFileLoader.java   # 抽象任务信息文件加载器基类
|   |       AbstractWorkloadFileLoader.java   # 抽象工作负载文件加载器基类
|   |       AllDriversFileLoader.java         # 所有驱动文件加载器接口
|   |       CSVAllDriversFileLoader.java      # CSV格式所有驱动文件加载器
|   |       CSVRunLoader.java                 # CSV格式运行信息加载器
|   |       CSVSnapshotLoader.java            # CSV格式快照加载器
|   |       CSVTaskInfoFileLoader.java        # CSV格式任务信息文件加载器
|   |       CSVWorkloadFileLoader.java        # CSV格式工作负载文件加载器
|   |       Formats.java                      # 加载格式定义
|   |       Loaders.java                      # 加载器工具类/工厂类
|   |       RunLoader.java                    # 运行信息加载器接口
|   |       SimpleWorkloadLoader.java         # 简单工作负载加载器
|   |       SnapshotLoader.java               # 快照加载器接口
|   |       TaskInfoFileLoader.java           # 任务信息文件加载器接口
|   |       WorkloadFileLoader.java           # 工作负载文件加载器接口
|   |
|   +---model          # 控制器模型定义
|   |       ControllerContext.java          # 控制器上下文
|   |       DriverContext.java              # 驱动上下文
|   |       DriverRegistry.java             # 驱动注册表
|   |       ErrorSummary.java               # 错误摘要
|   |       SchedulePlan.java               # 调度计划
|   |       ScheduleRegistry.java           # 调度注册表
|   |       SnapshotRegistry.java           # 快照注册表
|   |       StageContext.java               # 阶段上下文
|   |       StageListener.java              # 阶段监听器接口
|   |       StageRegistry.java              # 阶段注册表
|   |       TaskContext.java                # 任务上下文
|   |       TaskRegistry.java               # 任务注册表
|   |       WorkloadContext.java            # 工作负载上下文
|   |       WorkloadListener.java           # 工作负载监听器接口
|   |
|   +---repository     # 工作负载存储库
|   |       RAMWorkloadRepository.java      # 基于内存的工作负载存储库
|   |       SimpleWorkloadList.java         # 简单工作负载列表实现
|   |       WorkloadList.java               # 工作负载列表接口
|   |       WorkloadRepository.java         # 工作负载存储库接口
|   |
|   +---schedule       # 调度器
|   |       AbstractScheduler.java          # 抽象调度器基类
|   |       BalancedScheduler.java          # 均衡调度器实现
|   |       readme.md                       # 调度器说明文档
|   |       ScheduleException.java          # 调度异常
|   |       Schedulers.java                 # 调度器工具类/工厂类
|   |       WorkScheduler.java              # 工作调度器接口
|   |
|   +---service        # 控制器服务
|   |       ControllerThread.java           # 控制器线程
|   |       COSBControllerService.java      # COSBench控制器服务实现
|   |       COSBControllerServiceFactory.java # COSBench控制器服务工厂
|   |       OrderFuture.java                # 有序Future
|   |       OrderFutureComparator.java      # 有序Future比较器
|   |       OrderThreadPoolExecutor.java    # 有序线程池执行器
|   |       PingDriverRunner.java           # Ping驱动运行器
|   |       StageCallable.java              # 阶段Callable任务
|   |       StageChecker.java               # 阶段检查器
|   |       StageException.java             # 阶段异常
|   |       StageRunner.java                # 阶段运行器
|   |       TriggerRunner.java              # 触发器运行器
|   |       WorkloadException.java          # 工作负载异常
|   |       WorkloadProcessor.java          # 工作负载处理器
|   |
|   \---tasklet        # 任务单元
|           Aborter.java                    # 中止任务单元
|           AbstractCommandTasklet.java     # 抽象命令任务单元基类
|           AbstractHttpTasklet.java        # 抽象HTTP任务单元基类
|           AbstractTasklet.java            # 抽象任务单元基类
|           Authenticator.java              # 认证任务单元
|           Bootor.java                     # 启动任务单元 (拼写可能为Booter)
|           Closer.java                     # 关闭任务单元
|           Launcher.java                   # 启动任务单元
|           Querier.java                    # 查询任务单元
|           Submitter.java                  # 提交任务单元
|           Tasklet.java                    # 任务单元接口
|           TaskletException.java           # 任务单元异常
|           Tasklets.java                   # 任务单元工具类/工厂类
|           Trigger.java                    # 触发任务单元
|           TriggerHttpTasklet.java         # HTTP触发任务单元
|
+---controller-web
|   # 控制器Web前端
|   |   favicon.ico                       # 网站图标
|   |   index.html                        # 主页HTML
|   |
|   +---handler        # Web接口处理器
|   |       AbstractClientHandler.java      # 抽象客户端处理器基类
|   |       CancelHandler.java              # 取消操作处理器
|   |       ConfigHandler.java              # 配置操作处理器
|   |       IndexHandler.java               # 主页处理器
|   |       SubmitHandler.java              # 提交操作处理器
|   |       WorkloadHandler.java            # 工作负载操作处理器
|   |
|   +---resources      # 静态资源
|   |   |   bg-footer.png                 # 页脚背景图
|   |   |   bg-header.png                 # 页头背景图
|   |   |   cosbench.css                  # CSS样式表
|   |   |   down_arrow.png                # 向下箭头图标
|   |   |   up_arrow.png                  # 向上箭头图标
|   |   |
|   |   \---build
|   |       \---dist
|   |           |   echarts.js              # ECharts库文件
|   |           |
|   |           \---chart
|   |                   bar.js              # 柱状图配置
|   |                   line.js             # 折线图配置
|   |
|   +---web            # Web控制器
|   |       AdvancedConfigPageController.java # 高级配置页面控制器
|   |       CancelWorkloadController.java   # 取消工作负载控制器
|   |       CliLoginFilter.java             # 命令行登录过滤器
|   |       ConfigPageController.java       # 配置页面控制器
|   |       DownloadConfigController.java   # 下载配置控制器
|   |       DownloadLogController.java      # 下载日志控制器
|   |       IndexPageController.java        # 主页控制器
|   |       LogonErrorPageController.java   # 登录错误页面控制器
|   |       LogonPageController.java        # 登录页面控制器
|   |       MatrixPageController.java       # 性能矩阵页面控制器
|   |       PrometheusController.java       # Prometheus指标控制器
|   |       StagePageController.java        # 阶段页面控制器
|   |       SubmitPageController.java       # 提交页面控制器
|   |       TimelineCSVController.java      # 时间线CSV数据控制器
|   |       TimelinePageController.java     # 时间线页面控制器
|   |       WorkloadConfigGenerator.java    # 工作负载配置生成器
|   |       WorkloadConfigurationController.java # 工作负载配置控制器
|   |       WorkloadMatrixConfigurationController.java # 工作负载矩阵配置控制器
|   |       WorkloadPageController.java     # 工作负载页面控制器
|   |       WorkloadSubmissionController.java # 工作负载提交控制器
|   |
|   \---WEB-INF        # Web配置与模板
|       |   web.xml                         # Web应用部署描述符
|       |
|       +---freemarker
|       |       400.ftl                     # 400错误页面模板
|       |       404.ftl                     # 404错误页面模板
|       |       500.ftl                     # 500错误页面模板
|       |       advanced-config.ftl         # 高级配置页面模板
|       |       config.ftl                  # 配置页面模板
|       |       finalchart.ftl              # 最终图表模板
|       |       footer.ftl                  # 页脚模板
|       |       forchart.ftl                # 图表模板 (可能用于循环生成)
|       |       head.ftl                    # HTML头部模板
|       |       header.ftl                  # 页头模板
|       |       index.ftl                   # 主页模板
|       |       logon.ftl                   # 登录页面模板
|       |       logonError.ftl              # 登录错误页面模板
|       |       matrix.ftl                  # 性能矩阵页面模板
|       |       metrics.ftl                 # 指标页面模板
|       |       runningchart.ftl            # 运行中图表模板
|       |       stage.ftl                   # 阶段页面模板
|       |       submit.ftl                  # 提交页面模板
|       |       timeline-metrics.ftl        # 时间线指标模板
|       |       timeline.ftl                # 时间线页面模板
|       |       workload.ftl                # 工作负载页面模板
|       |
|       \---spring
|               controller-handler-context.xml # 控制器处理器Spring配置
|               controller-web-context.xml     # 控制器Web Spring配置
|               controller-web-osgi-context.xml# 控制器Web OSGi Spring配置
|
+---core
|   # 核心基准测试与模型
|   +---bench          # 基准测试工具
|   |       Aggregator.java                 # 指标聚合器
|   |       Benchmark.java                  # 基准测试核心类
|   |       Counter.java                    # 计数器
|   |       ErrorStatistics.java            # 错误统计
|   |       Histogram.java                  # 直方图 (用于延迟分布)
|   |       Mark.java                       # 时间标记
|   |       Metrics.java                    # 性能指标
|   |       Report.java                     # 测试报告
|   |       ReportMerger.java               # 测试报告合并器
|   |       Result.java                     # 单次操作结果
|   |       Sample.java                     # 采样数据
|   |       Snapshot.java                   # 性能快照
|   |       SnapshotMerger.java             # 性能快照合并器
|   |       Status.java                     # 状态枚举
|   |       TaskReport.java                 # 任务报告
|   |
|   +---model          # 核心模型定义
|   |       ControllerInfo.java             # 控制器信息模型
|   |       DriverInfo.java                 # 驱动信息模型
|   |       LifeCycle.java                  # 生命周期接口
|   |       MissionInfo.java                # 任务信息模型
|   |       MissionState.java               # 任务状态枚举
|   |       ScheduleInfo.java               # 调度信息模型
|   |       StageInfo.java                  # 阶段信息模型
|   |       StageState.java                 # 阶段状态枚举
|   |       StateInfo.java                  # 状态信息基类
|   |       StateRegistry.java              # 状态注册表接口
|   |       TaskInfo.java                   # 任务信息模型
|   |       TaskState.java                  # 任务状态枚举
|   |       WorkerInfo.java                 # 工作节点信息模型
|   |       WorkloadInfo.java               # 工作负载信息模型
|   |       WorkloadState.java              # 工作负载状态枚举
|   |
|   +---protocol       # 通信协议
|   |       AbortResponse.java              # 中止响应
|   |       CloseResponse.java              # 关闭响应
|   |       LaunchResponse.java             # 启动响应
|   |       LoginResponse.java              # 登录响应
|   |       PingResponse.java               # Ping响应
|   |       QueryResponse.java              # 查询响应
|   |       Response.java                   # 通用响应基类
|   |       SubmitResponse.java             # 提交响应
|   |       TriggerResponse.java            # 触发响应
|   |
|   +---service        # 核心服务
|   |       AbortedException.java           # 中止异常
|   |       AbstractServiceFactory.java     # 抽象服务工厂基类
|   |       CancelledException.java         # 取消异常
|   |       ControllerService.java          # 控制器服务接口
|   |       ControllerServiceFactory.java   # 控制器服务工厂接口
|   |       DriverService.java              # 驱动服务接口
|   |       DriverServiceFactory.java       # 驱动服务工厂接口
|   |       IllegalStateException.java      # 非法状态异常
|   |       TimeoutException.java           # 超时异常
|   |       UnexpectedException.java        # 未预期异常
|   |       WorkloadLoader.java             # 工作负载加载器接口 (与controller.loader中重复?)
|   |
|   \---utils          # 核心工具
|           AuthValidator.java              # 认证验证器
|           ListRegistry.java               # 列表注册表实现
|           MapRegistry.java                # Map注册表实现
|
+---core-web
|   # 核心Web异常处理
|       AbstractController.java           # 抽象Web控制器基类
|       BadRequestException.java          # 错误请求异常 (400)
|       NotFoundException.java            # 未找到异常 (404)
|
+---driver
|   # 驱动执行器与工具
|   +---agent          # 驱动代理
|   |       AbstractAgent.java              # 抽象代理基类
|   |       Agent.java                      # 代理接口
|   |       AgentException.java             # 代理异常
|   |       Agents.java                     # 代理工具类/工厂类
|   |       AuthAgent.java                  # 认证代理
|   |       WatchDog.java                   # 看门狗 (用于监控)
|   |       WorkAgent.java                  # 工作代理
|   |
|   +---generator      # 数据生成器
|   |       ConstantIntGenerator.java       # 常量整数生成器
|   |       DefaultSizeGenerator.java       # 默认大小生成器
|   |       Generators.java                 # 生成器工具类/工厂类
|   |       HistogramIntGenerator.java      # 直方图整数生成器
|   |       IntGenerator.java               # 整数生成器接口
|   |       NameGenerator.java              # 名称生成器接口
|   |       NumericNameGenerator.java       # 数字名称生成器
|   |       RandomInputStream.java          # 随机输入流
|   |       RangeIntGenerator.java          # 范围整数生成器
|   |       SequentialIntGenerator.java     # 顺序整数生成器
|   |       SizeGenerator.java              # 大小生成器接口
|   |       StreamUtils.java                # 流工具类
|   |       UniformIntGenerator.java        # 均匀分布整数生成器
|   |       XferCountingInputStream.java    # 传输计数输入流
|   |
|   +---iterator       # 数据迭代器
|   |       EmptyIterator.java              # 空迭代器
|   |       IntIterator.java                # 整数迭代器接口
|   |       Iterators.java                  # 迭代器工具类/工厂类
|   |       NameIterator.java               # 名称迭代器接口
|   |       NumericNameIterator.java        # 数字名称迭代器
|   |       RangeIterator.java              # 范围迭代器
|   |
|   +---model          # 驱动模型定义
|   |       DriverContext.java              # 驱动上下文
|   |       MissionContext.java             # 任务上下文
|   |       MissionListener.java            # 任务监听器接口
|   |       OperatorContext.java            # 操作器上下文
|   |       OperatorRegistry.java           # 操作器注册表
|   |       WorkerContext.java              # 工作节点上下文
|   |       WorkerRegistry.java             # 工作节点注册表
|   |
|   +---operator       # 操作器
|   |       AbstractOperator.java           # 抽象操作器基类
|   |       Cleaner.java                    # 清理操作器
|   |       Deleter.java                    # 删除操作器
|   |       Disposer.java                   # 处置操作器 (资源释放)
|   |       FileWriter.java                 # 文件写入操作器
|   |       Header.java                     # 获取头部信息操作器
|   |       Initializer.java                # 初始化操作器
|   |       Lister.java                     # 列举操作器
|   |       LocalWriter.java                # 本地写入操作器
|   |       MFileWriter.java                # 多文件写入操作器
|   |       MPreparer.java                  # 多文件准备操作器
|   |       MWriter.java                    # 多文件写入操作器 (与MFileWriter重复?)
|   |       OperationListener.java          # 操作监听器接口
|   |       Operator.java                   # 操作器接口
|   |       Operators.java                  # 操作器工具类/工厂类
|   |       Preparer.java                   # 准备操作器
|   |       Reader.java                     # 读取操作器
|   |       Restorer.java                   # 恢复操作器
|   |       Session.java                    # 会话管理
|   |       Writer.java                     # 写入操作器
|   |
|   +---repository     # 驱动存储库
|   |       MissionList.java                # 任务列表接口
|   |       MissionRepository.java          # 任务存储库接口
|   |       RAMMissionRepository.java       # 基于内存的任务存储库
|   |       SimpleMissionList.java          # 简单任务列表实现
|   |
|   +---service        # 驱动服务
|   |       COSBAuthAPIService.java         # COSBench认证API服务实现
|   |       COSBDriverService.java          # COSBench驱动服务实现
|   |       COSBDriverServiceFactory.java   # COSBench驱动服务工厂
|   |       COSBStorageAPIService.java      # COSBench存储API服务实现
|   |       MissionException.java           # 任务异常
|   |       MissionHandler.java             # 任务处理器
|   |
|   \---util           # 驱动工具
|           AuthCachePool.java              # 认证缓存池
|           ContainerPicker.java            # 容器选择器
|           Defaults.java                   # 默认值常量
|           Division.java                   # 分区/分片工具
|           FilePicker.java                 # 文件选择器
|           HashedFileInputStream.java      # 带哈希计算的文件输入流
|           HashUtil.java                   # 哈希工具类
|           ObjectPicker.java               # 对象选择器
|           ObjectScanner.java              # 对象扫描器
|           OperationPicker.java            # 操作选择器
|           SizePicker.java                 # 大小选择器
|
+---driver-web
|   # 驱动Web前端
|   |   index.html                        # 主页HTML
|   |
|   +---resources      # 静态资源
|   |       bg-footer.png                 # 页脚背景图
|   |       bg-header.png                 # 页头背景图
|   |       cosbench.css                  # CSS样式表
|   |
|   +---src            # Web接口与控制器
|   |   \---com
|   |       \---intel
|   |           \---cosbench
|   |               \---driver
|   |                   +---handler
|   |                   |       AbortHandler.java             # 中止任务处理器
|   |                   |       AbstractCommandHandler.java   # 抽象命令处理器基类
|   |                   |       CloseHandler.java             # 关闭任务处理器
|   |                   |       LaunchHandler.java            # 启动任务处理器
|   |                   |       LoginHandler.java             # 登录处理器
|   |                   |       MissionHandler.java           # 任务处理器
|   |                   |       PingHandler.java              # Ping处理器
|   |                   |       QueryHandler.java             # 查询处理器
|   |                   |       SubmitHandler.java            # 提交任务处理器
|   |                   |       TriggerHandler.java           # 触发处理器
|   |                   |
|   |                   \---web
|   |                           AbortMissionController.java   # 中止任务控制器
|   |                           CloseMissionController.java   # 关闭任务控制器
|   |                           DownloadLogController.java    # 下载日志控制器
|   |                           IndexPageController.java      # 主页控制器
|   |                           LaunchMissionController.java  # 启动任务控制器
|   |                           MissionPageController.java    # 任务页面控制器
|   |                           MissionSubmissionController.java # 任务提交控制器
|   |                           PerformLoginController.java   # 执行登录控制器
|   |                           SubmitPageController.java     # 提交页面控制器
|   |                           TriggerController.java        # 触发控制器
|   |                           WorkersPageController.java    # 工作节点页面控制器
|   |
|   \---WEB-INF        # Web配置与模板
|       |   web.xml                         # Web应用部署描述符
|       |
|       +---freemarker
|       |       400.ftl                     # 400错误页面模板
|       |       404.ftl                     # 404错误页面模板
|       |       500.ftl                     # 500错误页面模板
|       |       footer.ftl                  # 页脚模板
|       |       head.ftl                    # HTML头部模板
|       |       header.ftl                  # 页头模板
|       |       index.ftl                   # 主页模板
|       |       metrics.ftl                 # 指标页面模板
|       |       mission.ftl                 # 任务页面模板
|       |       submit.ftl                  # 提交页面模板
|       |       workers.ftl                 # 工作节点页面模板
|       |
|       \---spring
|               driver-handler-context.xml    # 驱动处理器Spring配置
|               driver-web-context.xml        # 驱动Web Spring配置
|               driver-web-osgi-context.xml   # 驱动Web OSGi Spring配置
|
\---sineio
    # SineIO扩展存储协议
    +---api            # 存储API
    |       SIOStorage.java                 # SineIO存储API实现
    |       SIOStorageFactory.java          # SineIO存储API工厂
    |
    \---client         # 客户端工具
            SIOConstants.java               # SineIO常量
