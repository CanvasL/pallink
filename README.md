# PalLink

PalLink 是一个基于 Go + go-zero 的活动社交练习项目，采用单仓库、多服务架构，当前通过 `Docker Compose` 进行本地联调和整体启动。

项目核心能力包括：

- 用户注册、登录、个人资料维护
- 活动创建、更新、报名、取消报名、签到、评论
- 公共活动浏览与详情查询
- 站内通知与已读处理
- 私聊会话、消息发送、消息列表、WebSocket 实时通信
- RabbitMQ 异步审核、通知投递
- Prometheus / Grafana / PgHero 运维观测

## 系统架构图

```text
+------------------------+         +------------------------+
| 浏览器 / App / 调试客户端 | ----> |   Caddy 统一入口层      |
+------------------------+         +-----------+------------+
                                               |
                                               v
                                    +----------+-----------+
                                    |      gateway-api     |
                                    | JWT / 限流 / 熔断 /  |
                                    | 超时 / WebSocket     |
                                    +----+--------+--------+
                                         |        |
                +------------------------+        +-------------------------+
                |                                                         |
                v                                                         v
      +---------+---------+   +---------------+   +----------------+   +--+--------------+
      |     user-rpc      |   | activity-rpc  |   | notification-rpc|   |     im-rpc     |
      | 注册 / 登录 / 资料 |   | 活动 / 报名 /  |   | 通知查询 / 已读 |   | 会话 / 消息 / WS |
      |                   |   | 签到 / 评论    |   | 消费通知事件     |   | 实时消息         |
      +---------+---------+   +-------+-------+   +--------+-------+   +--------+--------+
                |                     |                        |                    |
                +---------------------+------------------------+--------------------+
                                              |
                                              v
                                     +--------+--------+
                                     |    PostgreSQL   |
                                     | 用户 / 活动 /   |
                                     | 评论 / 通知 / IM |
                                     +--------+--------+
                                              ^
                                              |
                                     +--------+--------+
                                     |      PgHero     |
                                     | SQL / 慢查询分析 |
                                     +-----------------+

                +-------------------- gateway-api --------------------+
                |                                                     |
                v                                                     v
        +-------+--------+                                   +--------+-------+
        |     Redis      |                                   |      etcd      |
        | 限流 / 缓存预留 / |                                   |   服务发现      |
        | 分布式锁预留      |                                   +----------------+
        +-----------------+

      +------------------------ RabbitMQ -----------------------+
      |                                                         |
      v                                                         v
+-----+-----------+    活动消息     +-----------------+   通知 / 实时消息   +------------------+
|   activity-rpc  | -------------> |  audit-worker   | <---------------- | notification-rpc |
+-----------------+                | 自动审核 user /   |                   +------------------+
                                   | activity / comment |
                                   +---------+-------+
                                             |
                                             +--------------------+
                                             |                    |
                                             v                    v
                                       +-----+----+         +-----+-------+
                                       | user-rpc |         | activity-rpc |
                                       +----------+         +-------------+

      +----------------------- 监控观测链路 -----------------------+
      |                                                          |
      v                                                          v
+-----+-----------+    抓取 metrics    +-----------------+   展示仪表盘   +-----------------+
|   gateway-api   | -----------------> |   Prometheus    | -----------> |     Grafana     |
| user/activity/  |                    +-----------------+              +-----------------+
| notification/im |
+-----------------+
```

## 功能模块

| 模块 | 服务 | 说明 |
| --- | --- | --- |
| 接入层 | `gateway-api` | 统一 REST 入口，负责 JWT 鉴权、Redis 限流、go-zero 熔断/超时/过载保护，以及 WebSocket 接入 |
| 用户模块 | `user-rpc` | 注册、登录、获取个人信息、更新资料；注册后会投递审核事件 |
| 活动模块 | `activity-rpc` | 创建活动、更新活动、我的活动、公共活动、报名/取消报名、签到、参与者列表、评论 |
| 通知模块 | `notification-rpc` | 查询通知、标记已读；消费评论和私信事件并生成站内通知 |
| 即时通讯 | `im-rpc` | 打开会话、会话列表、消息发送、消息列表、会话已读；通过 WebSocket 推送实时消息 |
| 审核模块 | `audit-worker` | 消费 RabbitMQ 审核队列，更新 `user` / `activity` / `comment` 的审核状态 |
| 观测运维 | `Caddy` `Swagger UI` `Prometheus` `Grafana` `PgHero` | 统一入口、API 文档、指标抓取、仪表盘和数据库分析 |

## 当前实现说明

- 网关已经开启 `Breaker`、`Shedding`、`Timeout`、`MaxConns`、`Prometheus` 等 go-zero 中间件。
- 网关限流基于 go-zero 自带的 Redis `PeriodLimit`，已对登录、注册、公开接口、用户读写接口和 WebSocket 做了分组限流。
- Redis 配置已拆分为 `RateLimit`、`Cache`、`DistributedLock` 三组；当前代码实际接入的是限流，缓存和分布式锁作为后续扩展预留。
- `audit-worker` 当前是学习版实现：消费审核消息后会自动把用户、活动、评论更新为通过状态。
- PostgreSQL 启用了 `pg_stat_statements`，用于配合 PgHero 观察慢查询和 SQL 统计。

## 目录结构

```text
.
├── gateway/                    # HTTP 网关、Swagger、WebSocket 接入、限流
├── user/                       # 用户 RPC
├── activity/                   # 活动 RPC
├── notification/               # 通知 RPC
├── im/                         # IM RPC
├── audit/                      # RabbitMQ 审核 Worker
├── common/                     # 公共组件：auth、mq、postgres
├── deploy/                     # Docker、Caddy、监控配置
├── include/                    # protobuf 依赖
├── docker-compose.yml          # 本地整体编排
├── Makefile                    # 启动、日志、sqlc、goctl 命令
└── sqlc.yaml                   # sqlc 生成配置
```

SQL 与代码生成约定：

- API 网关入口定义在 [`gateway/gateway.api`](./gateway/gateway.api)
- RPC 定义分别在各服务目录下的 `*.proto`
- SQL 源文件位于 `*/sqlc/queries.sql`
- sqlc 生成代码位于 `*/internal/dao/sqlc`

## 快速启动

### 1. 启动全部服务

```bash
# 如果你的 Docker 构建依赖代理，可先开启本机代理，例如 proxyOn
make up-build

# 或者直接使用 docker compose
docker compose up --build -d
```

### 2. 查看运行状态

```bash
docker compose ps
make logs
```

### 3. 常用开发命令

```bash
make sqlc
make goctl-api
make goctl-rpc
make goctl
make down
```

说明：

- 首次启动时，PostgreSQL 会自动执行 [`deploy/docker/init.sql`](./deploy/docker/init.sql) 完成建表。
- 如果你之前已经保留了旧的 PostgreSQL 数据卷，而 PgHero 看不到 SQL 统计，可手动执行：

```bash
docker compose exec postgres psql -U pallink -d pallink -c 'CREATE EXTENSION IF NOT EXISTS pg_stat_statements;'
```

## Docker 公网部署

如果你要在公网主机上用子域名和 HTTPS 部署，可以直接复用仓库里的通用配置：

```bash
cp ./deploy/docker/compose.env.example ./deploy/docker/compose.env
docker compose --env-file ./deploy/docker/compose.env up -d
```

执行前请先调整 [`deploy/docker/compose.env.example`](./deploy/docker/compose.env.example) 里的主机名，并确认：

- `CADDYFILE_PATH=./deploy/caddy/Caddyfile.prod`
- `80/tcp` 和 `443/tcp` 已对公网放通
- `api`、`prometheus`、`grafana`、`pghero`、`rabbitmq` 这些子域名都已经解析到你的服务器公网 IP

默认入口：

- Swagger UI: `https://api.example.com/docs/`
- API: `https://api.example.com/`
- Prometheus: `https://prometheus.example.com/`
- Grafana: `https://grafana.example.com/`
- PgHero: `https://pghero.example.com/`
- RabbitMQ 管理台: `https://rabbitmq.example.com/`

说明：

- Swagger UI 和 API 同源，仓库里的 [`deploy/swagger/swagger.json`](./deploy/swagger/swagger.json) 不写 `host`/`schemes`，直接走相对地址模式。
- Swagger UI 默认开启 `persistAuthorization`，刷新页面后会保留浏览器里的授权信息。

## Kubernetes 本地测试

如果你想在本地先跑一套 K8s 环境，仓库里已经补了 `kind + kustomize` 方案：

前置要求：

- 已安装 `docker`
- 已安装 `kubectl`
- 已安装 `kind`
- 可选：已安装 `k9s`

一条命令拉起：

```bash
make k8s-kind-up
```

这条命令会自动：

- 创建或复用本地 `kind` 集群 `pallink-local`
- 安装本地 `ingress-nginx` controller
- 构建本地业务镜像和 K8s 专用运维镜像
- 把镜像 load 进集群
- 应用 [`deploy/k8s/overlays/local`](./deploy/k8s/overlays/local) 这套清单

本地 Ingress 入口：

- API: `http://localhost:8080/`
- Swagger UI: `http://localhost:8080/docs/`

本地直连端口：

- PostgreSQL: `localhost:5432`
- Redis: `localhost:6379`
- RabbitMQ: `localhost:5672`
- RabbitMQ 管理台: `http://localhost:15672/`
- PgHero: `http://localhost:8081/`
- Prometheus: `http://localhost:9090/`
- Grafana: `http://localhost:3000/`

可选调试工具：

```bash
make k8s-k9s
```

这会直接打开 `k9s`，默认连接到 `kind-pallink-local` context 和 `pallink` namespace。

说明：

- K8s 清单放在 [`deploy/k8s`](./deploy/k8s)。
- 本地 K8s 方案不再依赖 `caddy` 和 `etcd`。
- RPC 服务发现改成了 K8s Service 直连。
- 对外入口由 [`deploy/k8s/overlays/local/ingress.yaml`](./deploy/k8s/overlays/local/ingress.yaml) 定义，本地只保留 `localhost:8080` 的 API 和 Swagger。
- `kind` 里通过 `ingress-nginx` 接收 `8080 -> 80` 和 `8443 -> 443` 的宿主机映射。
- PostgreSQL、Redis、RabbitMQ、PgHero、Prometheus、Grafana 通过 `NodePort + kind extraPortMappings` 暴露给宿主机，配置在 [`deploy/k8s/overlays/local/nodeports`](./deploy/k8s/overlays/local/nodeports) 和 [`deploy/k8s/overlays/local/kind-config.yaml`](./deploy/k8s/overlays/local/kind-config.yaml)。
- RabbitMQ 的 `5672` 是 AMQP TCP 端口，不走 Ingress，直接用 `localhost:5672`。
- 如果你只想把清单打到当前 kube context，而不是走 kind 工作流，可以直接执行 `make k8s-apply`。这种模式下你需要自己准备镜像可拉取地址，并确保集群里已经有 Ingress controller。
- 销毁本地集群执行 `make k8s-kind-down`。

## Docker Compose 启动后访问地址

### 业务与文档入口

| 名称 | 地址 | 说明 |
| --- | --- | --- |
| 网关基址 | `http://localhost:8080` | 业务接口统一入口 |
| Swagger UI | `http://localhost:8080/docs/` | 在线接口文档 |
| Swagger JSON | `http://localhost:8080/docs/swagger.json` | 原始 OpenAPI 文档 |
| 公共活动列表示例 | `http://localhost:8080/activity/public/list` | 无需登录即可访问 |
| IM WebSocket | `ws://localhost:8080/im/ws?token=<JWT>` | 实时消息入口 |

### 运维与中间件入口

| 名称 | 地址 | 默认账号 |
| --- | --- | --- |
| PgHero | `http://localhost:8081/` | `admin / pallink` |
| Prometheus | `http://localhost:9090/` | 无 |
| Grafana | `http://localhost:3000/` | `admin / pallink` |
| RabbitMQ 管理台 | `http://localhost:15672/` | `admin / pallink` |

说明：

- 本地 Docker Compose 默认只用 `localhost` 和端口，不需要额外域名或 `/etc/hosts`。

### 基础设施连接信息

| 组件 | 地址 | 默认信息 |
| --- | --- | --- |
| PostgreSQL | `localhost:5432` | `db=pallink user=pallink password=pallink` |
| Redis | `localhost:6379` | 无密码 |
| etcd | `localhost:2379` | 服务发现 |

### RPC 服务端口

| 服务 | 端口 |
| --- | --- |
| `user-rpc` | `8002` |
| `activity-rpc` | `8004` |
| `im-rpc` | `8005` |
| `notification-rpc` | `8006` |

## 接口范围概览

公开接口：

- `POST /user/register`
- `POST /user/login`
- `GET /activity/public/list`
- `GET /activity/public/detail`
- `GET /activity/public/comment/list`

特殊接口：

- `GET /im/ws` 不走 JWT 中间件，但连接时仍需通过 `token` 参数或 `Authorization` 头传入 JWT

登录后接口：

- `GET /user/me`
- `POST /user/update`
- `POST /activity/create`
- `POST /activity/update`
- `GET /activity/list`
- `GET /activity/my`
- `GET /activity/enrolled`
- `GET /activity/detail`
- `POST /activity/enroll`
- `POST /activity/cancelEnroll`
- `POST /activity/checkIn`
- `GET /activity/participants`
- `POST /activity/comment/create`
- `GET /notification/list`
- `POST /notification/read`
- `POST /im/conversation/open`
- `GET /im/conversation/list`
- `POST /im/message/send`
- `GET /im/message/list`
- `POST /im/conversation/read`

更完整的接口字段说明，请直接查看 Swagger：

- [`deploy/swagger/swagger.json`](./deploy/swagger/swagger.json)
- `http://localhost:8080/docs/`
