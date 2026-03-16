# FSVChart Notify

[![License](https://img.shields.io/github/license/jicki/fsvchart-notify)](https://github.com/jicki/fsvchart-notify/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/jicki/fsvchart-notify)](https://github.com/jicki/fsvchart-notify/releases)

FSVChart Notify 是一个专注于 PromQL 监控指标可视化和飞书通知的开源工具。它能够帮助团队监控关键指标并绘制图表，通过飞书机器人推送到指定群组，提升运维效率。

## 功能特性

- **PromQL 查询管理** — 创建、分类、复用预定义 PromQL 查询，支持语法高亮
- **图表可视化** — 多种图表模板，支持图表/文本/混合展示模式
- **推送任务** — 灵活配置定时推送，支持多数据源、多 WebHook、多 PromQL 组合
- **飞书通知** — 通过飞书机器人 WebHook 推送图表卡片到群组
- **发送记录** — 完整的推送历史记录，便于追溯和排查
- **权限管理** — Admin/User 角色分级，Admin 管理系统配置，User 查看数据
- **用户管理** — 管理员可查看用户列表、修改角色、重置本地用户密码
- **LDAP 认证** — 支持 LDAP 统一认证，自动创建本地账户，同步角色
- **系统管理** — 数据源与 WebHook 统一管理页面
- **SQLite 存储** — 零依赖持久化，自动 Schema 迁移
- **单一二进制** — 前端通过 Go embed 嵌入，单文件部署

## 快速开始

### 环境要求

- Go 1.23+
- Node.js 16+（前端构建）
- Docker（可选）

### 安装与构建

```bash
git clone https://github.com/jicki/fsvchart-notify.git
cd fsvchart-notify

# 一键构建（前端 + statik + 后端）
make build
```

### 配置

创建 `config.yaml`：

```yaml
server:
  address: "0.0.0.0"
  port: 8080

auth:
  jwt_secret: "your-secret-key"    # JWT 签名密钥，生产环境务必修改
  token_expiry_hours: 24           # Token 有效期（小时）
  ldap:
    enabled: false                 # 是否启用 LDAP 认证
    host: "ldap.example.com"
    port: 389
    use_tls: false
    bind_dn: "cn=admin,dc=example,dc=com"
    bind_password: ""
    base_dn: "ou=users,dc=example,dc=com"
    user_filter: "(uid=%s)"
    display_name_attr: "cn"
    email_attr: "mail"
    default_role: "user"           # LDAP 用户默认角色：user 或 admin
    admin_group_dn: ""             # 可选，LDAP Admin 组 DN
```

### 运行

```bash
# 直接运行
make run

# 或使用 Docker
make docker-build
docker run -p 8080:8080 -v ./data:/app/data -v ./config.yaml:/app/config.yaml fsvchart-notify
```

访问 `http://localhost:8080`，默认管理员账户：`admin` / `123456`。

## 项目结构

```
fsvchart-notify
├── build/                  # Dockerfile
├── cmd/                    # 程序入口
├── internal/               # 内部包
│   ├── config/            # 配置管理
│   ├── database/          # 数据库与自动迁移
│   ├── handler/           # 业务处理器
│   ├── middleware/         # JWT 认证、权限中间件
│   ├── models/            # 数据模型
│   ├── scheduler/         # 定时任务调度
│   ├── server/            # HTTP 路由与 API
│   └── service/           # 业务逻辑（认证、LDAP）
├── frontend/               # Vue 3 + TypeScript 前端
│   ├── src/
│   │   ├── components/    # 通用组件（AppLayout、ModalDialog、icons）
│   │   ├── composables/   # 组合式函数（useCrudList、usePolling 等）
│   │   ├── stores/        # Pinia 状态管理
│   │   ├── views/         # 页面视图
│   │   └── styles/        # 全局样式（暗色主题）
│   └── ...
├── statik/                 # 嵌入的前端静态资源
└── config.yaml             # 运行时配置
```

## 配置说明

### 服务器配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `server.address` | 监听地址 | `0.0.0.0` |
| `server.port` | 监听端口 | `8080` |

### 认证配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.jwt_secret` | JWT 签名密钥 | `fsvchart-notify-secret-key` |
| `auth.token_expiry_hours` | Token 过期时间（小时） | `24` |

### LDAP 配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `auth.ldap.enabled` | 启用 LDAP | `false` |
| `auth.ldap.host` | LDAP 服务器地址 | - |
| `auth.ldap.port` | LDAP 端口 | `389` |
| `auth.ldap.use_tls` | 使用 TLS | `false` |
| `auth.ldap.bind_dn` | Bind DN | - |
| `auth.ldap.bind_password` | Bind 密码 | - |
| `auth.ldap.base_dn` | 搜索 Base DN | - |
| `auth.ldap.user_filter` | 用户搜索过滤器 | `(uid=%s)` |
| `auth.ldap.display_name_attr` | 显示名属性 | `cn` |
| `auth.ldap.email_attr` | 邮箱属性 | `mail` |
| `auth.ldap.default_role` | 默认角色 | `user` |
| `auth.ldap.admin_group_dn` | Admin 组 DN（可选） | - |

## 开发指南

### 本地开发

```bash
# 启动后端
make run

# 启动前端开发服务器（另一个终端）
cd frontend && npm run dev
```

### 构建 Docker 镜像

```bash
make docker-build
```

## API 概览

所有 API 均以 `/api` 为前缀，需要 JWT Bearer Token 认证。

### 公开接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/login` | 用户登录 |

### 认证用户接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/me` | 获取当前用户信息 |
| PUT | `/api/me` | 更新个人信息 |
| PUT | `/api/me/password` | 修改密码 |
| GET | `/api/metrics_source` | 数据源列表 |
| GET | `/api/feishu_webhook` | WebHook 列表 |
| GET | `/api/push_task` | 推送任务列表 |
| GET | `/api/promqls` | PromQL 查询列表 |
| GET | `/api/send_records` | 发送记录列表 |

### 管理员接口

| 方法 | 路径 | 说明 |
|------|------|------|
| POST/PUT/DELETE | `/api/metrics_source[/:id]` | 数据源管理 |
| POST/PUT/DELETE | `/api/feishu_webhook[/:id]` | WebHook 管理 |
| POST/PUT/DELETE | `/api/push_task[/:id]` | 推送任务管理 |
| POST/PUT/DELETE | `/api/promql[/:id]` | PromQL 管理 |
| GET | `/api/users` | 用户列表 |
| PUT | `/api/users/:id/role` | 修改用户角色 |
| PUT | `/api/users/:id/password` | 重置用户密码 |

## 开源协议

本项目采用 MIT 协议 - 查看 [LICENSE](LICENSE) 文件了解详情。
