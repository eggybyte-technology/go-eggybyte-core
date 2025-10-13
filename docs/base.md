### **企业级 Go 微服务基础库 `go-eggybyte-core` & CLI 工具 `ebcctl` 终版设计与实现指南**

#### **第一部分：顶层设计 (High-Level Design)**

##### **1.1. 项目愿景与目标**

`go-eggybyte-core` 生态旨在为 Go 微服务开发提供一个“开箱即用”且“高度规范”的解决方案。其核心目标是：

  * **极致效率**: 通过 `ebcctl` 脚手架和 `core.Bootstrap` 一键启动器，将新服务的创建和基础搭建时间从数小时缩短至数分钟。
  * **高度规范**: 统一日志、配置、监控、数据库交互等最佳实践，降低代码维护成本，提升团队协作效率。
  * **云原生就绪**: 深度集成 Kubernetes，提供动态配置、健康检查等能力，简化云原生环境下的开发与部署。

##### **1.2. 核心设计哲学**

  * **职责分离**: 各功能模块（`log`, `db`, `config`）高度内聚、低耦合。
  * **可插拔架构**: 核心组件（数据库、缓存）均可通过配置按需启用。
  * **配置驱动**: 服务行为由外部配置（环境变量、K8s ConfigMap）定义。
  * **云原生感知**: 内建对 Kubernetes 资源的动态感知能力。
  * **约定优于配置**: 通过自动化机制（如 Repository 自动注册）减少样板代码。

##### **1.3. 整体架构与工作流**

`go-eggybyte-core` 生态由两部分组成：核心库 `go-eggybyte-core` 和命令行工具 `ebcctl`。

**开发者工作流:**

1.  `ebcctl init my-service`  -\> 创建标准化的项目结构。
2.  `ebcctl new repo user`    -\> 在项目中自动生成 `user_repo.go` 模板。
3.  **开发者编写业务逻辑** -\> 在 Handler 和 Repository 中填充具体实现。
4.  `docker build ...`        -\> 使用生成的 Dockerfile 构建镜像。
5.  `kubectl apply ...`       -\> 部署到 Kubernetes。

  (这是一个概念图示)

```
+------------------+     +--------------------------+     +---------------------+
|   Developer      | --> |        ebcctl CLI        | --> |  Generated Project  |
+------------------+     +--------------------------+     +---------------------+
                           | 1. init                  |     | - main.go         |
                           | 2. new repo/service      |     | - go.mod          |
                           +--------------------------+     | - repository/     |
                                                            | ...               |
                                                            +---------------------+
                                                                    |
                                                                    | Imports & Uses
                                                                    v
+-----------------------------------------------------------------------------------+
|                               go-eggybyte-core Library                            |
|-----------------------------------------------------------------------------------|
| /core | /config | /log | /db | /cache | /service | /metrics | /health             |
+-----------------------------------------------------------------------------------+
```

-----

#### **第二部分：`go-eggybyte-core` 库深度实现指南**

这是 `go-eggybyte-core` 库的核心，每个模块的实现都应遵循以下要点。

##### **2.1. 配置模块 (`/config`)**

  * **职责**: 提供统一、多源的配置加载方案，并支持 K8s 动态配置。
  * **核心文件**: `config/env.go`, `config/k8s_watcher.go`, `config/config.go`
  * **实现要点**:
    1.  **基础加载**: 使用 `github.com/kelseyhightower/envconfig` 库实现从环境变量到 Go 结构体的映射。提供 `config.ReadFromEnv(cfg)` 函数。
    2.  **多源决策**: 在 `core.Bootstrap` 中，通过检查环境变量 `$CONFIG_SOURCE` (值为 `env` 或 `kubernetes`) 来决定配置策略。
    3.  **K8s 动态监听**:
          * 引入 `k8s.io/client-go` 库。
          * 实现 `WatchK8sConfig` 函数，该函数使用 `in-cluster` 配置创建 K8s `clientset`。
          * 利用 `informer` 机制监听指定 `namespace` 下的 `ConfigMap`。需要从环境变量 `$K8S_CONFIGMAP_NAME` 和 `$K8S_NAMESPACE` 获取目标信息。
          * 在 `informer` 的 `UpdateFunc` 事件回调中，重新解析 `ConfigMap.Data` 并更新到一个 **全局、线程安全** 的配置实例中。
    4.  **线程安全**: 全局配置实例必须由 `sync.RWMutex` 保护。提供 `config.Get()` 方法（使用读锁）供业务代码安全访问，提供内部的 `update()` 方法（使用写锁）供监听器更新配置。

##### **2.2. 日志模块 (`/log`)**

  * **职责**: 提供标准化的、结构化的、上下文感知的日志接口。
  * **核心文件**: `log/log.go`
  * **实现要点**:
    1.  **接口抽象**: 定义一个 `Logger` 接口，包含 `Info(msg string, fields ...Field)`, `Error(msg string, fields ...Field)` 等方法。`Field` 是一个键值对类型。这使得底层日志库可以被替换。
    2.  **默认实现**: 推荐使用 `go.uber.org/zap` 作为底层实现，因为它性能高、功能强大，并天然支持结构化日志。
    3.  **初始化**: `log.Init(config)` 函数根据配置（如 `level`, `format`）初始化全局的 `Logger` 实例。
    4.  **上下文感知 (Context-aware)**: 提供 `log.WithContext(ctx)` 和 `log.FromContext(ctx)` 方法。这允许将请求 ID (trace id) 等信息注入 `context`，日志库可以自动从 `context` 中提取这些字段并添加到每条日志中，方便链路追踪。

##### **2.3. 数据库模块 (`/db`)**

  * **职责**: 封装数据库客户端，提供自动化表初始化、日志、指标等功能。
  * **核心文件**: `db/db.go`, `db/registry.go`, `db/tidb.go`
  * **实现要点**:
    1.  **Repository 接口**: 定义 `db.Repository` 接口，包含 `TableName() string` 和 `InitTable(ctx, db) error` 两个方法。这是实现自动化初始化的契约。
    2.  **自动注册机制**:
          * 在 `db/registry.go` 中，维护一个全局的 `var registeredRepositories []Repository` 切片和一个互斥锁 `sync.Mutex`。
          * 提供 `db.RegisterRepository(repo Repository)` 函数，业务代码通过它来注册自己的 Repository 实例。
          * **关键模式**: 业务仓库 (`user_repo.go`) 必须使用 `init()` 函数来调用 `db.RegisterRepository`，实现自动、解耦的注册。
    3.  **初始化器 (Initializer)**:
          * 实现 `db.TiDBInitializer`，它遵循 `service.Initializer` 接口。
          * 在其 `Init(ctx)` 方法中，首先连接数据库（推荐使用 `gorm.io/gorm`），然后调用 `db.InitializeAllTables(ctx, db)`。
          * `InitializeAllTables` 函数遍历 `registeredRepositories` 切片，依次执行每个 Repository 的 `InitTable` 方法。
    4.  **客户端封装**: 创建 `LoggerDB` 结构体，内嵌 `*gorm.DB`。通过重写 GORM 的钩子 (Callbacks) 或方法，在每次数据库操作（`Query`, `Exec`）前后自动记录包含 SQL、耗时、影响行数的日志，并更新 Prometheus 指标（如 `db_query_duration_seconds`）。
    5.  **全局实例**: 提供 `db.SetDB(*LoggerDB)` 和 `db.GetDB() *LoggerDB`，用于在初始化后注入和在业务代码中获取数据库实例。

##### **2.4. 服务生命周期模块 (`/service`)**

  * **职责**: 管理所有后台服务（HTTP, gRPC, Metrics 等）的启动、运行和优雅关闭。
  * **核心文件**: `service/launcher.go`, `service/http.go`
  * **实现要点**:
    1.  **Service 接口**: 定义 `type Service interface { Start(ctx) error; Stop(ctx) error }`。任何需要独立运行和关闭的单元（如 HTTP 服务器）都应实现此接口。
    2.  **Initializer 接口**: 定义 `type Initializer interface { Init(ctx) error }`。任何需要在服务启动前完成的初始化步骤（如连接数据库）都应实现此接口。
    3.  **Launcher 结构体**:
          * 包含 `initializers []Initializer` 和 `services []Service` 两个切片。
          * `AddInitializer()` 和 `AddService()` 方法用于注册。
          * `Init()` 方法按顺序执行所有 `initializers`。
          * `Run()` 方法是核心：
              * 使用 `sync.WaitGroup` 和 `errgroup` 来并发启动所有 `services`。
              * 创建一个 `chan os.Signal` 来监听 `syscall.SIGINT` 和 `syscall.SIGTERM` 信号。
              * 当接收到退出信号时，调用一个 `shutdown()` 方法。
              * `shutdown()` 方法会创建一个带超时的 `context`，然后并发或顺序地调用所有 `services` 的 `Stop()` 方法，实现优雅关闭。

##### **2.5. 可观测性模块 (`/metrics`, `/health`)**

  * **职责**: 提供标准的健康检查和 Prometheus 指标端点。
  * **核心文件**: `metrics/health.go`, `metrics/metrics.go`
  * **实现要点**:
    1.  **独立端口**: Metrics 和 Health 服务必须监听在一个独立的端口上（如 `9090`），与业务端口分离。这确保了即使业务逻辑线程池耗尽，可观测性端点依然可用。
    2.  **健康检查 (`/healthz`, `/readyz`)**:
          * 实现 `health.Service`，它启动一个 HTTP 服务器。
          * `/healthz` (Liveness): 用于检查服务进程是否存活，通常直接返回 `200 OK`。
          * `/readyz` (Readiness): 用于检查服务是否准备好接收流量。应实现检查逻辑，例如 Ping 数据库和缓存，确认依赖项是否正常。当检查失败时返回 `503 Service Unavailable`。
    3.  **Prometheus 指标 (`/metrics`)**:
          * 引入 `github.com/prometheus/client_golang` 库。
          * 实现 `metrics.Service`，它启动一个 HTTP 服务器。
          * 注册 `promhttp.Handler()` 到 `/metrics` 路径。
          * 注册一些默认的 Go 进程指标 (`collectors.NewGoCollector()`)。
          * 在其他模块（如 `/db`）中，定义并注册自定义指标（如 `Counter`, `Gauge`, `Histogram`），并在相应操作时更新它们。

##### **2.6. 核心启动器 (`/core`)**

  * **职责**: 作为框架的唯一入口，编排所有模块的初始化和启动。
  * **核心文件**: `core/bootstrap.go`
  * **实现要点**:
    1.  **单一入口**: `Bootstrap` 函数是业务 `main.go` 唯一需要调用的核心库函数。
    2.  **编排流程**: 严格按照顺序执行：配置加载 -\> 日志初始化 -\> Launcher 创建 -\> 核心服务（Metrics/Health）注册 -\> 可选组件（DB/Cache）的 Initializer 注册 -\> 执行所有 Initializer -\> 注册业务 Service。
    3.  **简洁性**: `Bootstrap` 隐藏了所有复杂的初始化细节，向开发者提供了一个极其简单的接口。

-----

#### **第三部分：`ebcctl` CLI 工具实现指南**

##### **3.1. 定位与职责**

`ebcctl` 是 `go-eggybyte-core` 的官方脚手架和开发工具，旨在自动化项目创建和模块生成，强制执行项目结构规范。

##### **3.2. 仓储与安装**

  * **仓储**: 源代码位于 `go-eggybyte-core` 主仓库的 `cmd/ebcctl` 目录下。
  * **安装**: 开发者通过在 `go-eggybyte-core` 根目录执行 `go install ./cmd/ebcctl` 来安装到 `$GOPATH/bin`。

##### **3.3. 命令设计**

| 命令 | 描述 |
| :--- | :--- |
| `ebcctl init <service-name>` | 在当前目录创建并初始化一个完整的微服务项目骨架。 |
| `ebcctl new repo <model-name>` | 在当前项目的 `internal/repository` 目录下生成一个 Repository 模板文件。 |
| `ebcctl new service <service-name> --type <http\|grpc>` | （可选高级功能）生成一个服务模块，包含 handler, router 等。 |

##### **3.4. 实现要点**

1.  **技术栈**:
      * **命令解析**: `github.com/spf13/cobra`
      * **模板引擎**: Go 原生的 `text/template`
2.  **模板驱动**:
      * 在 `cmd/ebcctl/templates` 目录下存放所有代码模板文件（`.tpl` 后缀）。
      * 模板中应使用占位符，如 `{{.ServiceName}}`, `{{.ModelName}}`, `{{.TableName}}`。
3.  **`init` 命令实现**:
      * 接收 `service-name` 参数。
      * 创建项目根目录。
      * 递归地创建子目录结构（`/cmd`, `/internal/repository` 等）。
      * 遍历 `templates` 目录中用于项目初始化的模板（`main.go.tpl`, `go.mod.tpl`, `Dockerfile.tpl` 等）。
      * 为每个模板创建一个数据结构（如 `struct { ServiceName string }`），并使用 `text/template` 包的 `Execute` 方法将渲染结果写入目标文件。
4.  **`new repo` 命令实现**:
      * 接收 `model-name` 参数。
      * 从参数中推断出结构体名（如 `user` -\> `User`）和表名（如 `user` -\> `users`）。
      * 读取 `repo.go.tpl` 模板。
      * 渲染模板并将结果写入 `internal/repository/<model-name>_repo.go`。

-----

#### **第四部分：最佳实践与部署**

##### **4.1. 开发者工作流**

1.  `go install github.com/eggybyte-technology/go-eggybyte-core/cmd/ebcctl@latest`
2.  `ebcctl init user-service`
3.  `cd user-service`
4.  `ebcctl new repo user`
5.  `ebcctl new repo product`
6.  在 `main.go` 中，**匿名导入** `_ "user-service/internal/repository"`。
7.  在 `internal/repository/user_repo.go` 和 `product_repo.go` 中编写 `CREATE TABLE` SQL。
8.  编写业务 Handler，并通过 `db.GetDB()` 调用数据库。
9.  `go run ./cmd/user-service` 本地运行和测试。

##### **4.2. Kubernetes 部署**

  * 使用 `ebcctl init` 生成的 `Dockerfile` 构建多阶段、精简的镜像。
  * 编写 `Deployment`, `Service`, `ConfigMap`, `Secret` 等 YAML 文件。
  * 在 `Deployment.yaml` 中：
      * **必须** 设置 `CONFIG_SOURCE: "kubernetes"` 等环境变量。
      * **必须** 配置 `livenessProbe` 和 `readinessProbe` 指向 `health` 服务端口的 `/healthz` 和 `/readyz`。
      * **强烈建议** 为 Pod 配置一个有权读取 `ConfigMap` 的 `ServiceAccount`，并遵守最小权限原则 (RBAC)。
      * 通过 `secretKeyRef` 将数据库 DSN 等敏感信息注入环境变量。

这份终版指南为您和您的团队提供了一份从理论到实践、从编码到部署的全面路线图。遵循此指南，将能构建出一个健壮、高效、可维护的 Go 微服务体系。