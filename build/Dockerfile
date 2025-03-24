# ===================
# Dockerfile
# ===================

# 1. Node 构建阶段 (编译前端)
FROM reg.deeproute.ai/deeproute-public/node:18-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/ ./
RUN npm install
RUN npm run build
# 上一步会生成 dist 目录，在 /app/frontend/dist 下

# 2. Go 构建阶段
FROM reg.deeproute.ai/deeproute-public/golang:1.23-alpine AS go-builder
WORKDIR /app

# 添加版本参数
ARG VERSION

# 拷贝 go.mod / go.sum 先
COPY go.mod go.sum ./
RUN go mod download

# 拷贝后端源码
COPY . .

# 在多阶段构建里，需要先把第一个阶段编译好的 web 文件复制过来
COPY --from=frontend-builder /app/frontend/dist ./web

# 运行 statik 命令打包
RUN go install github.com/rakyll/statik@latest
RUN statik -src=./web -f

# Install gcc
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add --no-cache gcc g++ make musl-dev

# 编译 go
RUN mkdir -p bin
RUN CGO_ENABLED=1 go build -ldflags "-X main.version=${VERSION}" -o bin/fsvchart-notify ./cmd/main.go

# 3. 运行阶段 (最小化镜像)
FROM reg.deeproute.ai/deeproute-public/alpine:latest
WORKDIR /app
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata

# 拷贝可执行文件
COPY --from=go-builder /app/bin/fsvchart-notify /app/

# 如果需要 config.yaml
COPY --from=go-builder /app/config.yaml /app/config.yaml

# # 如果需要 data/ 目录
# COPY --from=go-builder /app/data /app/data

EXPOSE 8080
CMD ["./fsvchart-notify", "-config=./config.yaml", "-db=./data/app.db"]
