### 核心设计理念

该项目结构遵循四大核心理念，旨在实现开发效率、可维护性和部署性能的最大化。

1.  **Monorepo（单体仓库）**: 所有代码（API 定义、后端服务、前端应用、部署配置）共存于一个 Git 仓库。这确保了 **API 的单一事实来源**，简化了跨团队协作和版本控制。
2.  **API-First（API 优先）**: `proto` 目录是项目的“心脏”。所有的数据结构和RPC服务都在此定义。通过 `buf` 自动生成类型安全的 Go 和 TypeScript 代码，确保了前后端契约的强一致性。
3.  **Build Artifact Separation（构建产物分离）** 🏗️: 这是一个关键的性能优化策略。`Makefile` 负责在本地或 CI/CD 环境中编译生成最终产物（Go 二进制文件、包含 JS Bridge 的 Flutter Web 静态文件），并将它们统一存放在根目录的 `build/` 文件夹下。**Docker 的职责被简化为纯粹的打包**，它仅将这些预编译好的产物复制到极简的运行时镜像中，从而使镜像构建过程**极快**且**稳定**。
4.  **JS Bridge for Web（Web 的 JS 桥接）** 🌉: 遵循您提供的规范，Flutter Web 不直接使用 Dart gRPC-Web，而是通过 `dart:js_interop` 与一个专门的 **TypeScript Bridge** 通信。这个 Bridge 使用原生的 `Connect-ES` 库与后端通信，具有**无需 Envoy 代理**、**更小的包体积**和**完整的流支持**等优点。

-----

### 完整文件夹结构

```plaintext
eggybyte-example-project/
├── Makefile                # 自动化指令的统一入口 (核心！)
├── README.md
├── .gitignore
├── .dockerignore           # 优化Docker构建上下文，排除不必要的文件
|
├── build/                  # 存放所有构建产物 (由Makefile生成，gitignored)
│   ├── backend/
│   │   └── user-service    # Go服务的二进制文件
│   └── frontend/
│       └── dashboard/      # Flutter Web构建出的完整静态网站
|
├── buf.yaml                # Buf模块定义
├── buf.gen.yaml            # Buf代码生成配置 (Go + TypeScript)
|
├── proto/                  # API定义 (Source of Truth)
│   └── eggybyte/
│       └── user/
│           └── v1/
│               └── user_service.proto
|
├── backend/                # Go Workspace 根目录
│   ├── go.work             # Go Workspace 定义文件
│   │
│   ├── services/           # 所有独立的Go微服务
│   │   └── user-service/
│   │       ├── cmd/main.go # 服务入口
│   │       ├── internal/   # 服务内部实现
│   │       └── go.mod      # 独立的Go模块
│   │
│   └── gen/                # Buf生成的共享Go代码
│
├── frontend/               # 多前端项目容器
│   └── dashboard/          # Flutter Web应用: "dashboard"
│       ├── lib/            # Dart 源代码
│       │   ├── main.dart
│       │   └── core/
│       │       ├── js_bridge.dart    # Dart JS interop 实现
│       │       └── models.dart       # 手动编写的Dart数据模型
│       │
│       ├── web/            # ⚠️ TypeScript Bridge 项目
│       │   ├── package.json          # npm 依赖 (@connectrpc/connect, etc.)
│       │   ├── tsconfig.json         # TypeScript 配置
│       │   ├── src/
│       │   │   └── bridge.ts         # TS Bridge 核心实现
│       │   ├── gen/                  # buf 生成的TypeScript代码 (gitignored)
│       │   └── index.html            # 加载JS Bridge和Flutter
│       │
│       └── pubspec.yaml
│
└── deploy/                 # 部署相关配置
    ├── templates/          # 存放通用的Dockerfile模板
    │   ├── Dockerfile.go
    │   └── Dockerfile.nginx
    └── charts/
        └── eggybyte-chart/ # 主 Helm Chart
            ├── Chart.yaml
            ├── values.yaml
            └── templates/

```

-----

### 关键文件详解

#### 1\. `buf.gen.yaml` (Go + TypeScript 生成)

此文件精确地配置了 Go 和 TypeScript 代码的生成，完全符合您提供的 `Connect-ES v2` 方案。

```yaml
# buf.gen.yaml
version: v1
plugins:
  # ===== Go Backend Plugins =====
  # 1. Go Protocol Buffers (messages and types)
  - local: protoc-gen-go
    out: ../pb
    opt:
      - paths=source_relative
  
  # 2. Connect-RPC for Go (modern RPC framework by Buf)
  #    Generates Connect service handlers and clients
  #    Compatible with gRPC, gRPC-Web, and Connect clients
  - local: protoc-gen-connect-go
    out: ../pb
    opt:
      - paths=source_relative

  # ===== TypeScript/Connect-Web Frontend Plugins =====
  
  - remote: buf.build/bufbuild/es
    out: ../frontend/dashboard/web/gen
    opt:
      - target=ts
```

#### 2\. `frontend/dashboard/web/package.json`

这是 TypeScript Bridge 项目的核心，定义了依赖和构建脚本。

```json
{
    "name": "yao-oracle-dashboard",
    "version": "1.0.0",
    "description": "Yao-Oracle Dashboard - TypeScript Bridge for Flutter Web",
    "type": "module",
    "scripts": {
        "prebuild": "npm run clean",
        "build": "npm run compile && npm run bundle",
        "compile": "tsc",
        "bundle": "esbuild js/src/bridge.js --bundle --format=iife --global-name=yaoOracleAPI --outfile=bridge.bundle.js --sourcemap",
        "watch": "tsc --watch",
        "dev": "concurrently \"npm run watch\" \"npm run watch:bundle\"",
        "watch:bundle": "esbuild js/src/bridge.js --bundle --format=iife --global-name=yaoOracleAPI --outfile=bridge.bundle.js --sourcemap --watch",
        "clean": "rm -rf js bridge.bundle.js bridge.bundle.js.map",
        "lint": "eslint src --ext .ts",
        "format": "prettier --write \"src/**/*.ts\"",
        "test": "node test-api.js",
        "test:help": "node test-api.js --help"
    },
    "dependencies": {
        "@bufbuild/protobuf": "^2.9.0",
        "@connectrpc/connect": "^2.1.0",
        "@connectrpc/connect-web": "^2.1.0"
    },
    "devDependencies": {
        "typescript": "^5.3.3",
        "@types/node": "^20.11.0",
        "esbuild": "^0.19.11",
        "concurrently": "^8.2.2",
        "eslint": "^8.56.0",
        "@typescript-eslint/eslint-plugin": "^6.19.0",
        "@typescript-eslint/parser": "^6.19.0",
        "prettier": "^3.2.4"
    }
}
```

*我们在这里使用 `esbuild`，因为它非常快速且能轻松打包成 `iife` (立即调用函数表达式) 格式，这对在 `index.html` 中安全加载至关重要。*

#### 3\. `Makefile` (升级版)

`Makefile` 是整个工作流的粘合剂，它精确地编排了“构建”和“打包”两个分离的阶段。

```makefile
# Makefile

# --- Configuration ---
BACKEND_SERVICES := user-service
FRONTEND_APPS := dashboard
DOCKER_REGISTRY := your-registry.com/eggybyte

# --- High-Level Commands ---
.PHONY: all
all: proto build docker-build

.PHONY: build
build: build-backend build-frontend

.PHONY: docker-build
docker-build:
	@echo "🐳 Building all Docker images from pre-built artifacts..."
	@$(foreach service,$(BACKEND_SERVICES), \
		docker build -t $(DOCKER_REGISTRY)/$$service:latest -f ./deploy/templates/Dockerfile.go --build-arg SERVICE_NAME=$$service .; \
	)
	@$(foreach app,$(FRONTEND_APPS), \
		docker build -t $(DOCKER_REGISTRY)/$$app:latest -f ./deploy/templates/Dockerfile.nginx --build-arg APP_NAME=$$app .; \
	)

# --- Atomic Build Steps ---
.PHONY: proto
proto:
	@echo "🚀 Generating Go & TypeScript code from proto files..."
	@buf generate

.PHONY: build-backend
build-backend:
	@echo "🛠️ Compiling Go services to ./build/backend..."
	@mkdir -p build/backend
	@$(foreach service,$(BACKEND_SERVICES), \
		CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/backend/$$service ./backend/services/$$service/cmd; \
	)

.PHONY: build-frontend
build-frontend: proto
	@echo "📦 Building Flutter Web app with JS Bridge to ./build/frontend..."
	@mkdir -p build/frontend
	@$(foreach app,$(FRONTEND_APPS), \
		echo "  -> Building JS Bridge for $$app..."; \
		(cd frontend/$$app/web && npm install && npm run build); \
		echo "  -> Building Flutter app $$app..."; \
		(cd frontend/$$app && flutter build web --release --web-renderer canvaskit -o ../../build/frontend/$$app); \
	)

# --- Utility ---
.PHONY: clean
clean:
	@echo "🧹 Cleaning all generated code and build artifacts..."
	@rm -rf build backend/gen frontend/*/web/gen frontend/*/web/node_modules frontend/*/web/dist
```

#### 4\. 通用 Dockerfile 模板

这些文件位于 `deploy/templates/` 目录，由 `Makefile` 在构建时使用。

##### `deploy/templates/Dockerfile.go`

```dockerfile
# deploy/templates/Dockerfile.go
ARG SERVICE_NAME
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /app

# 关键：只复制预先编译好的二进制文件
COPY build/backend/${SERVICE_NAME} .

EXPOSE 8080
CMD ["/app/${SERVICE_NAME}"]
```

##### `deploy/templates/Dockerfile.nginx`

```dockerfile
# deploy/templates/Dockerfile.nginx
ARG APP_NAME
FROM nginx:1.29.2-alpine

# 关键：只复制预先构建好的整个Web应用
COPY build/frontend/${APP_NAME} /usr/share/nginx/html

# 添加SPA重定向配置 (可选，但推荐)
COPY deploy/nginx/default.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

-----

### 开发者工作流

1.  **API 设计**: 在 `proto/` 目录中修改或创建 `.proto` 文件。
2.  **代码生成**: 运行 `make proto`。`buf` 会立即更新 `backend/gen` 中的 Go 代码和 `frontend/dashboard/web/gen` 中的 TypeScript 代码。
3.  **后端开发**: 在 `backend/services/user-service/` 中实现 Connect RPC 服务。
4.  **前端开发**:
      * 在 `frontend/dashboard/web/src/bridge.ts` 中，使用新生成的 TS 类型封装对后端的调用，并暴露给 `window` 对象。
      * 在 `frontend/dashboard/lib/` 中，编写 Dart 代码，通过 `js_bridge.dart` 调用 TypeScript 函数，实现业务逻辑和 UI。
5.  **完整构建**: 准备部署时，在项目根目录运行 `make build`。此命令会：
      * 编译所有 Go 服务。
      * 构建 JS Bridge。
      * 构建 Flutter Web 应用。
      * 所有产物都干净地存放在 `build/` 目录下。
6.  **镜像打包**: 运行 `make docker-build`。此命令会为每个服务和应用执行一个**极快**的 `docker build` 过程，因为它只涉及文件复制。
7.  **部署**: 将镜像推送到仓库，并使用 Helm Chart (`deploy/charts/`) 进行部署。

这个结构为您提供了一个健壮、高效且完全符合现代云原生实践的项目基础。