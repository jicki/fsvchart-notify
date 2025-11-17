# ==============
# Makefile
# ==============

PROJECT_NAME := fsvchart-notify
BIN_DIR := bin
WEB_DIR := web
FRONTEND_DIR := frontend

VERSION := $(shell cat VERSION)
BUILD_FLAGS := -ldflags "-X fsvchart-notify/cmd/main.version=$(VERSION)"

# 默认目标
.PHONY: all
all: clean build run

# -----------
# 前端
# -----------
.PHONY: dev-frontend
dev-frontend:
	@echo "=> Start Vue dev server..."
	cd $(FRONTEND_DIR) && npm install && npm run dev

.PHONY: build-frontend
build-frontend:
	@echo "=> Building Vue frontend..."
	@echo "=> Installing frontend dependencies..."
	rm -rf $(FRONTEND_DIR)/dist
	cd $(FRONTEND_DIR) && \
	npm install && \
	npm run build
	
	@echo "=> Creating web directory..."
	mkdir -p $(WEB_DIR)
	@echo "=> Copying frontend dist to web directory..."
	rm -rf $(WEB_DIR)/*
	cp -r $(FRONTEND_DIR)/dist/* $(WEB_DIR)/
	@echo "=> Frontend build completed"

# -----------
# statik
# -----------
.PHONY: build-statik
build-statik:
	@echo "=> Packing static files with statik..."
	go install github.com/rakyll/statik@latest
	statik -src=$(WEB_DIR) -f

# -----------
# 后端
# -----------
.PHONY: build-backend
build-backend: $(BIN_DIR)/$(PROJECT_NAME)

$(BIN_DIR)/$(PROJECT_NAME):
	@echo "=> Building Go backend..."
	go mod tidy
	mkdir -p $(BIN_DIR)
	go build $(BUILD_FLAGS) -o $(BIN_DIR)/$(PROJECT_NAME) ./cmd/main.go

# -----------
# 整体build
# -----------
.PHONY: build
build: build-frontend build-statik build-backend
	@echo "=> Build finished successfully."

# -----------
# 开发
# -----------
.PHONY: dev-build
dev-build: build-frontend build-statik
	@echo "=> Development build finished."
	go run cmd/main.go

# -----------
# 运行
# -----------
.PHONY: run
run: build
	@echo "=> Running $(PROJECT_NAME)..."
	./$(BIN_DIR)/$(PROJECT_NAME) -config=./config.yaml -db=./data/app.db

# -----------
# Docker
# -----------
.PHONY: docker
docker:
	@echo "=> Reading version from VERSION file..."
	$(eval VERSION := $(shell cat VERSION))
	@echo "=> Building Docker image with tag: $(VERSION)..."
	docker build --build-arg VERSION=$(VERSION) -t reg.deeproute.ai/deeproute-public/$(PROJECT_NAME):$(VERSION) -f build/Dockerfile .
	@echo "=> Tagging as latest as well..."
	docker tag reg.deeproute.ai/deeproute-public/$(PROJECT_NAME):$(VERSION) reg.deeproute.ai/deeproute-public/$(PROJECT_NAME):latest
	docker push reg.deeproute.ai/deeproute-public/$(PROJECT_NAME):$(VERSION)

# -----------
# 清理
# -----------
.PHONY: clean
clean:
	@echo "=> Cleaning..."
	rm -rf $(BIN_DIR)
	rm -rf $(WEB_DIR)/*
