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

## 阿里云 ACR 部署

如果你要把自研服务镜像和第三方基础镜像都同步到阿里云 ACR，可以直接使用仓库内置脚本：

```bash
# 先登录 ACR
docker login --username=<your-aliyun-account> registry.cn-hangzhou.aliyuncs.com

# 按 deploy/aliyun/compose.env.example 配置命名空间和标签
make aliyun-sync
```

这条命令会完成三件事：

- 基于 [`docker-compose.yml`](./docker-compose.yml) 生成阿里云版 [`docker-compose.aliyun.yml`](./docker-compose.aliyun.yml)
- 复制第三方镜像的 manifest list 到 `ALIYUN_MIRROR_NAMESPACE`
- 使用 `docker buildx build --platform linux/amd64 --push` 把自研服务推送到 `ALIYUN_APP_NAMESPACE`

执行前请先确认两点：

- [`deploy/aliyun/compose.env.example`](./deploy/aliyun/compose.env.example) 里的 `ALIYUN_REGISTRY` 必须改成你在 ACR 控制台看到的精确登录域名。新个人版、旧个人版、企业版域名格式不同，`docker login` 和 `docker push` 必须使用同一个域名。
- `ALIYUN_APP_NAMESPACE` 和 `ALIYUN_MIRROR_NAMESPACE` 这两个命名空间必须已经在 ACR 中创建好。如果你关闭了 ACR 的自动创建仓库能力，还需要预先创建对应仓库。

如果推送时报 `insufficient_scope: authorization failed`，通常就是以下几类问题：

- `ALIYUN_REGISTRY` 填错了，尤其是把新个人版或企业版实例误写成了 `registry.cn-hangzhou.aliyuncs.com`
- `docker login` 登录的域名和实际 `push` 的域名不一致
- 目标命名空间不存在，或者当前账号没有该命名空间/仓库的推送权限
- 自动建仓已关闭，但目标仓库还没创建

`make aliyun-sync` 默认只同步 `linux/amd64`，因为它的目标就是给云服务器部署，不额外照顾本地 Apple Silicon。

如果部署时报 `no matching manifest for linux/amd64`，通常不是仓库权限问题，而是镜像平台不对：

- 你在 Apple Silicon 上执行了普通的 `docker pull` 或 `docker compose build`
- 然后把本地 `arm64` 单架构镜像直接推到了 ACR
- 最后在 `amd64` ECS 上拉取同一个 tag，就会报这个错

仓库里的同步脚本现在会：

- 第三方镜像通过 `docker buildx imagetools create --platform linux/amd64` 只同步云服务器需要的平台
- 自研服务默认只构建 `linux/amd64`

如果你以后真想发多架构，再临时覆盖即可：

```bash
APP_PLATFORMS=linux/amd64,linux/arm64 \
MIRROR_PLATFORMS=linux/amd64,linux/arm64 \
make aliyun-sync
```

如果你只想重写阿里云版 compose，不想推镜像：

```bash
make aliyun-compose
```

同步完成后可以直接启动：

```bash
make aliyun-up
```

如果你要用子域名和 HTTPS 部署：

- 把 `api.pallink.us.ci`、`grafana.pallink.us.ci`、`pghero.pallink.us.ci` 的 DNS A 记录指到你的服务器公网 IP
- 在服务器安全组和本机防火墙放通 `80/tcp` 和 `443/tcp`
- 使用 [`deploy/aliyun/compose.env`](./deploy/aliyun/compose.env) 里的 `API_HOST`、`GRAFANA_HOST`、`PGHERO_HOST`、`CADDYFILE_PATH`、`CADDY_HTTP_PORT`、`CADDY_HTTPS_PORT` 配置
- Swagger UI 会挂在 `https://api.pallink.us.ci/docs/`
- 业务 API 会挂在 `https://api.pallink.us.ci/`
- Grafana 会挂在 `https://grafana.pallink.us.ci/`
- PgHero 会挂在 `https://pghero.pallink.us.ci/`
- Swagger UI 和 API 现在同源，仓库里的 [`deploy/swagger/swagger.json`](./deploy/swagger/swagger.json) 不写 `host`/`schemes`，直接走相对地址模式
- Swagger UI 默认开启 `persistAuthorization`，刷新页面后会保留浏览器里的授权信息

## 启动后访问地址

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
| PgHero | `http://pghero.localhost:8080/` | `admin / pallink` |
| Prometheus | `http://prometheus.localhost:8080/` | 无 |
| Grafana | `http://grafana.localhost:8080/` | `admin / pallink` |
| RabbitMQ 管理台 | `http://localhost:15672/` | `pallink / pallink` |

说明：

- `pghero.localhost`、`prometheus.localhost`、`grafana.localhost` 走的是 `Caddy` 的 Host 路由。
- 现代系统通常会把 `*.localhost` 自动解析到本机回环地址。

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
