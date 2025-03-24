# FSVChart Notify

[![Go Report Card](https://goreportcard.com/badge/github.com/jicki/fsvchart-notify)](https://goreportcard.com/report/github.com/jicki/fsvchart-notify)
[![License](https://img.shields.io/github/license/jicki/fsvchart-notify)](https://github.com/jicki/fsvchart-notify/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/jicki/fsvchart-notify)](https://github.com/jicki/fsvchart-notify/releases)

FSVChart Notify 是一个专注于 `PromQL` 监控指标可视化和飞书通知的开源工具。它能够帮助团队监控关键指标绘制成相关图表，并通过飞书机器人推送到指定群组，提升运维效率。

## ✨ 功能特性

- 🔍 实时监控 `PromQL` 指标数据
- 📊 美观的图表可视化界面
- 🔔 灵活的飞书通知配置
- 🚀 支持自定义告警规则和阈值
- 💾 使用 SQLite 持久化存储配置
- 🔄 内置定时任务调度系统

## 🚀 快速开始

### 环境要求

- Go 1.23.3 或更高版本
- Node.js 16+ (用于前端开发)
- Docker (可选，用于容器化部署)

### 安装

1. 克隆仓库

```bash
git clone https://github.com/jicki/fsvchart-notify.git
cd fsvchart-notify
```

2. 安装依赖

```bash
# 后端依赖
go mod download

# 前端依赖
cd frontend
npm install
cd ..
```

3. 编译项目

```bash
make build
```

### 配置

创建 `config.yaml` 文件：

```yaml
server:
  address: "0.0.0.0"
  port: 8080
```

### 运行

```bash
# 直接运行
make run

# 或使用 Docker
make docker
docker run -p 8080:8080 fsvchart-notify
```

访问 `http://localhost:8080` 即可打开管理界面。

## 📚 项目结构

```
fsvchart-notify
├── build                    # Docker 相关配置
├── cmd                      # 程序入口
├── internal                 # 内部包
│   ├── config              # 配置管理
│   ├── database            # 数据库操作
│   ├── models              # 数据模型
│   ├── scheduler           # 定时任务
│   ├── server              # HTTP 服务
│   └── service             # 业务逻辑
├── frontend                # Vue.js 前端项目
├── statik                  # 静态资源
└── web                     # 编译后的前端资源
```

## 🔧 开发指南

### 本地开发

1. 启动后端服务

```bash
make run
```

2. 启动前端开发服务器

```bash
cd frontend
npm run dev
```

### 构建发布

```bash
# 构建完整项目
make build

# 构建 Docker 镜像
make docker
```

## 📝 配置说明

### 服务器配置

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| server.address | 监听地址 | 0.0.0.0 |
| server.port | 监听端口 | 8080 |

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！在贡献代码前，请确保：

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交改动 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 📄 开源协议

本项目采用 MIT 协议 - 查看 [LICENSE](LICENSE) 文件了解详情

## 🙏 致谢

- [Gin Web Framework](https://github.com/gin-gonic/gin)
- [Vue.js](https://vuejs.org/)
- [VictoriaMetrics](https://victoriametrics.com/)