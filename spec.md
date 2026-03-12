## 📚 PalLink 技术架构文档 v1.0 (PostgreSQL版)

> 基于 go-zero 微服务框架的搭子社交平台

---

## 一、项目概述

### 1.1 项目简介
PalLink 是一个连接搭子的社交平台，帮助用户找到兴趣相投的伙伴一起参与活动。核心功能包括用户管理、活动发布、报名签到、即时通讯等。

### 1.2 技术栈选型

| 分类 | 技术 | 用途 |
|------|------|------|
| 框架 | go-zero | 微服务框架 |
| 语言 | Golang 1.20+ | 开发语言 |
| 数据库 | PostgreSQL 15+ | 业务数据存储 |
| 缓存 | Redis 7.0 | 会话管理、计数、分布式锁 |
| 消息队列 | RabbitMQ | 异步处理、削峰填谷 |
| 服务发现 | etcd | 服务注册与发现 |
| 通讯协议 | gRPC | 服务间调用 |
| 实时通讯 | WebSocket | 即时消息 |
| 网关 | go-zero gateway | API 网关 |
| 监控 | Prometheus + Grafana | 监控告警 |

---

## 二、系统架构

### 2.1 整体架构图

```
┌─────────────────────────────────────────────────────┐
│                   客户端 (App/Web)                    │
└─────────────────────┬───────────────────────────────┘
                      │ HTTP/WebSocket
                      ▼
┌─────────────────────────────────────────────────────┐
│                    API 网关 (Gateway)                 │
│                路由、鉴权、限流、熔断                    │
└──┬──────────┬──────────┬──────────┬─────────────────┘
   │          │          │          │
   ▼          ▼          ▼          ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐
│ 用户服务 │ │ 活动服务 │ │  IM服务  │ │ 通知服务 │
│ user-rpc│ │activity-│ │  im-rpc │ │notify-  │
│         │ │   rpc   │ │         │ │   rpc   │
└────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘
     │           │           │           │
     └───────────┼───────────┴───────────┘
                 │
         ┌───────┴───────┐
         │   服务发现      │
         │    etcd       │
         └───────────────┘

┌──────────────────基础设施──────────────────┐
│ PostgreSQL   Redis   RabbitMQ   MinIO(OSS) │
└───────────────────────────────────────────┘
```

### 2.2 服务拆分

| 服务名 | 端口 | 职责 | 依赖 |
|--------|------|------|------|
| user-api | 8001 | 用户HTTP接口 | user-rpc |
| user-rpc | 8002 | 用户核心逻辑 | PostgreSQL, Redis |
| activity-api | 8003 | 活动HTTP接口 | activity-rpc |
| activity-rpc | 8004 | 活动核心逻辑 | PostgreSQL, Redis, RabbitMQ |
| im-api | 8005 | IM HTTP/WebSocket | im-rpc |
| im-rpc | 8006 | IM核心逻辑 | PostgreSQL, Redis, RabbitMQ |
| notify-rpc | 8007 | 通知服务 | RabbitMQ |
| gateway | 8080 | API网关 | 所有服务 |

---

## 三、PostgreSQL 数据库设计

### 3.1 通用配置

```sql
-- 创建数据库
CREATE DATABASE pallink
    WITH 
    OWNER = postgres
    ENCODING = 'UTF8'
    LC_COLLATE = 'en_US.UTF-8'
    LC_CTYPE = 'en_US.UTF-8'
    TEMPLATE = template0;

-- 切换到数据库
\c pallink;

-- 创建扩展
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";      -- UUID生成
CREATE EXTENSION IF NOT EXISTS "postgis";        -- 地理位置支持
CREATE EXTENSION IF NOT EXISTS "btree_gin";      -- 索引优化
CREATE EXTENSION IF NOT EXISTS "pg_trgm";        -- 模糊搜索支持

-- 创建更新时间触发器函数
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';
```

### 3.2 用户服务 (user-rpc)

```sql
-- 用户基础表
CREATE TABLE "user" (
    id BIGSERIAL PRIMARY KEY,
    mobile VARCHAR(20) NOT NULL,
    nickname VARCHAR(50) NOT NULL,
    avatar VARCHAR(255),
    gender SMALLINT DEFAULT 0 NOT NULL CHECK (gender IN (0, 1, 2)), -- 0未知 1男 2女
    birthday DATE,
    school VARCHAR(100),
    intro TEXT,
    status SMALLINT DEFAULT 1 NOT NULL CHECK (status IN (1, 2)), -- 1正常 2禁用
    last_login_time TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_user_mobile UNIQUE (mobile)
);

-- 创建索引
CREATE INDEX idx_user_nickname ON "user" USING gin (nickname gin_trgm_ops);
CREATE INDEX idx_user_created_at ON "user" (created_at);

-- 更新时间触发器
CREATE TRIGGER update_user_updated_at
    BEFORE UPDATE ON "user"
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 用户兴趣标签表
CREATE TABLE user_tag (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_user_tag UNIQUE (user_id, tag_id)
);

-- 创建索引
CREATE INDEX idx_user_tag_user_id ON user_tag (user_id);
CREATE INDEX idx_user_tag_tag_id ON user_tag (tag_id);

-- 用户关系表（关注/好友）
CREATE TABLE user_relation (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    target_user_id BIGINT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    relation_type SMALLINT NOT NULL CHECK (relation_type IN (1, 2)), -- 1关注 2好友
    status SMALLINT DEFAULT 1 CHECK (status IN (0, 1)), -- 0待确认 1已确认
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_user_relation UNIQUE (user_id, target_user_id, relation_type)
);

CREATE INDEX idx_user_relation_user_id ON user_relation (user_id);
CREATE INDEX idx_user_relation_target ON user_relation (target_user_id);
```

### 3.3 活动服务 (activity-rpc)

```sql
-- 活动分类表
CREATE TABLE activity_category (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    icon VARCHAR(255),
    sort_order INTEGER DEFAULT 0,
    status SMALLINT DEFAULT 1 CHECK (status IN (0, 1)), -- 0禁用 1启用
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_category_name UNIQUE (name)
);

-- 活动标签表
CREATE TABLE activity_tag (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    category_id INTEGER REFERENCES activity_category(id),
    status SMALLINT DEFAULT 1 CHECK (status IN (0, 1)),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_tag_name UNIQUE (name)
);

-- 活动表
CREATE TABLE activity (
    id BIGSERIAL PRIMARY KEY,
    creator_id BIGINT NOT NULL,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    category_id INTEGER NOT NULL REFERENCES activity_category(id),
    location VARCHAR(255) NOT NULL,
    location_geo GEOGRAPHY(POINT), -- PostGIS地理位置
    address_detail VARCHAR(255),
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ NOT NULL,
    deadline_time TIMESTAMPTZ NOT NULL,
    max_people INTEGER DEFAULT 0, -- 0表示不限
    current_people INTEGER DEFAULT 0,
    min_people INTEGER DEFAULT 1,
    fee_amount DECIMAL(10,2) DEFAULT 0, -- 费用 0表示免费
    fee_type SMALLINT DEFAULT 1 CHECK (fee_type IN (1, 2)), -- 1免费 2AA制 3收费
    status SMALLINT DEFAULT 0 CHECK (status IN (0, 1, 2, 3, 4)), -- 0草稿 1招募中 2进行中 3已结束 4已取消
    views INTEGER DEFAULT 0,
    images TEXT[], -- PostgreSQL数组类型存储多张图片
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT check_time_valid CHECK (start_time < end_time AND deadline_time <= start_time)
);

-- 创建索引
CREATE INDEX idx_activity_creator ON activity (creator_id);
CREATE INDEX idx_activity_status ON activity (status);
CREATE INDEX idx_activity_start_time ON activity (start_time);
CREATE INDEX idx_activity_location ON activity USING GIST (location_geo);
CREATE INDEX idx_activity_title ON activity USING gin (title gin_trgm_ops);

-- 全文搜索索引
CREATE INDEX idx_activity_search ON activity USING gin(
    to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(description,''))
);

-- 更新时间触发器
CREATE TRIGGER update_activity_updated_at
    BEFORE UPDATE ON activity
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- 活动标签关联表
CREATE TABLE activity_tag_relation (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL REFERENCES activity(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES activity_tag(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_activity_tag UNIQUE (activity_id, tag_id)
);

-- 报名表
CREATE TABLE enrollment (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL REFERENCES activity(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    status SMALLINT DEFAULT 1 CHECK (status IN (1, 2, 3)), -- 1已报名 2已签到 3已取消
    enroll_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checkin_time TIMESTAMPTZ,
    checkin_location GEOGRAPHY(POINT), -- 签到地理位置
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_activity_user UNIQUE (activity_id, user_id)
);

-- 创建索引
CREATE INDEX idx_enrollment_user ON enrollment (user_id);
CREATE INDEX idx_enrollment_status ON enrollment (status);
CREATE INDEX idx_enrollment_enroll_time ON enrollment (enroll_time);

-- 活动收藏表
CREATE TABLE activity_favorite (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT NOT NULL REFERENCES activity(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_favorite_activity_user UNIQUE (activity_id, user_id)
);

CREATE INDEX idx_favorite_user ON activity_favorite (user_id);
```

### 3.4 IM服务 (im-rpc)

```sql
-- 私聊消息表
CREATE TABLE private_message (
    id BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL,
    to_user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    msg_type SMALLINT DEFAULT 1 CHECK (msg_type IN (1, 2, 3, 4)), -- 1文本 2图片 3语音 4视频
    status SMALLINT DEFAULT 1 CHECK (status IN (1, 2)), -- 1未读 2已读
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- 添加外键约束（需要确保users表存在）
    CONSTRAINT fk_private_message_from FOREIGN KEY (from_user_id) REFERENCES "user"(id),
    CONSTRAINT fk_private_message_to FOREIGN KEY (to_user_id) REFERENCES "user"(id)
);

-- 分区表：按月份分区
CREATE TABLE private_message_partitioned (
    LIKE private_message INCLUDING ALL
) PARTITION BY RANGE (created_at);

-- 创建分区示例
CREATE TABLE private_message_202401 PARTITION OF private_message_partitioned
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- 创建索引
CREATE INDEX idx_private_message_from ON private_message (from_user_id, created_at);
CREATE INDEX idx_private_message_to ON private_message (to_user_id, status, created_at);

-- 群聊表
CREATE TABLE group_chat (
    id BIGSERIAL PRIMARY KEY,
    activity_id BIGINT REFERENCES activity(id) ON DELETE SET NULL,
    name VARCHAR(100) NOT NULL,
    avatar VARCHAR(255),
    creator_id BIGINT NOT NULL REFERENCES "user"(id),
    announcement TEXT,
    max_members INTEGER DEFAULT 100,
    status SMALLINT DEFAULT 1 CHECK (status IN (1, 2)), -- 1正常 2解散
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_chat_activity ON group_chat (activity_id);

-- 群成员表
CREATE TABLE group_member (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES group_chat(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    role SMALLINT DEFAULT 1 CHECK (role IN (1, 2, 3)), -- 1成员 2管理员 3群主
    nickname VARCHAR(50), -- 群昵称
    join_time TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_read_time TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP, -- 最后阅读时间
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    CONSTRAINT uk_group_member UNIQUE (group_id, user_id)
);

CREATE INDEX idx_group_member_user ON group_member (user_id);

-- 群消息表
CREATE TABLE group_message (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES group_chat(id) ON DELETE CASCADE,
    from_user_id BIGINT NOT NULL REFERENCES "user"(id),
    content TEXT NOT NULL,
    msg_type SMALLINT DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_group_message_group ON group_message (group_id, created_at);
```

### 3.5 通知服务 (notify-rpc)

```sql
-- 通知表
CREATE TABLE notification (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES "user"(id) ON DELETE CASCADE,
    type SMALLINT NOT NULL CHECK (type IN (1, 2, 3, 4)), -- 1系统通知 2活动通知 3互动通知 4私信提醒
    title VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    data JSONB, -- PostgreSQL JSONB类型存储额外数据
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX idx_notification_user ON notification (user_id, is_read, created_at);
CREATE INDEX idx_notification_created_at ON notification (created_at);

-- 使用BRIN索引处理大量通知数据
CREATE INDEX idx_notification_created_at_brin ON notification USING BRIN (created_at);
```

---
