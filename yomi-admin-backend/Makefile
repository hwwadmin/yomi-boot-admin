# yomi-admin-backend Makefile
# 请在 yomi-admin-backend 根目录执行 make 命令
#
# 原则：日常开发统一通过 make 调用；script/ 为实现细节，不直接执行。
# 说明：本仓为 go.work 多 module，tidy/fmt/vet/lint/test 会遍历 script/dev/workspace-modules.sh 里的 WORKSPACE_MODULES。

GO          ?= go
BIN_DIR     ?= bin
TARGET      ?= linux_amd64,linux_arm64
TAG         ?= latest
SERVICE     ?= yomi-admin-backend
PORT        ?= 8080
comma       := ,

# docker-build 产物名
IMAGE_NAME  := yomi-admin-backend

.DEFAULT_GOAL := help

ifeq ($(OS),Windows_NT)
SHELL_DEV   := powershell -NoProfile -ExecutionPolicy Bypass -File
SCRIPT      = $(SHELL_DEV) script/dev/Invoke-GitBash.ps1 script/$(1)
CLEAN_BIN   = powershell -NoProfile -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue '$(BIN_DIR)'"
else
SCRIPT      = bash script/$(1)
CLEAN_BIN   = rm -rf $(BIN_DIR)
endif

# check
# 作用: 本地提交前门禁
# 用法: make check
.PHONY: check
check: fmt vet lint

# deps
# 作用: 同步 go.work，并整理各 module 的 go.mod / go.sum
# 用法: make deps
.PHONY: deps
deps:
	$(GO) work sync
	$(call SCRIPT,dev/tidy.sh)

# tidy
# 作用: 同步 go.work，并整理各 module 的 go.mod / go.sum
# 用法: make tidy
.PHONY: tidy
tidy:
	$(GO) work sync
	$(call SCRIPT,dev/tidy.sh)

# cross-build
# 作用: 交叉编译编排层（apps/server/cmd）到 bin/（文件名带平台后缀）；默认 linux_amd64 + linux_arm64
# 用法: make cross-build
#       make cross-build TARGET=linux_amd64
#       make cross-build TARGET=windows_amd64,linux_amd64
.PHONY: cross-build
cross-build:
	$(call SCRIPT,_build/build.sh) $(subst $(comma), ,$(TARGET))

# docker-build
# 作用: 交叉编译并构建 Docker 镜像，导出制品到 bin/yomi-admin-backend-<tag>.tar
# 用法: make docker-build
#       make docker-build TAG=v0.1.0
# 加载: docker load -i bin/yomi-admin-backend-latest.tar
.PHONY: docker-build
docker-build:
	@$(MAKE) cross-build TARGET=linux_amd64
	@echo ">>> Building image: $(IMAGE_NAME):$(or $(TAG),latest)"
	docker build -f script/image/Dockerfile \
		--build-arg BIN_PATH=$(BIN_DIR)/$(SERVICE)-linux-amd64 \
		--build-arg PORT=$(PORT) \
		-t $(IMAGE_NAME):$(or $(TAG),latest) .
	@echo ">>> Saving image artifact: $(BIN_DIR)/$(IMAGE_NAME)-$(or $(TAG),latest).tar"
	docker save -o $(BIN_DIR)/$(IMAGE_NAME)-$(or $(TAG),latest).tar $(IMAGE_NAME):$(or $(TAG),latest)

# clean
# 作用: 删除 bin/
# 用法: make clean
.PHONY: clean
clean:
	$(CLEAN_BIN)

# test
# 作用: 对各 module 执行单元测试（排除 apps/server/test 集成测试目录）
# 用法: make test
.PHONY: test
test:
	$(call SCRIPT,dev/test.sh)

# fmt
# 作用: 对各 module 执行 go fmt + gofumpt
# 用法: make fmt
.PHONY: fmt
fmt:
	$(call SCRIPT,dev/fmt.sh)

# vet
# 作用: 对各 module 执行 go vet
# 用法: make vet
.PHONY: vet
vet:
	$(call SCRIPT,dev/vet.sh)

# lint
# 作用: 对各 module 执行 golangci-lint（读取根目录 .golangci.yml）
# 用法: make lint
.PHONY: lint
lint:
	$(call SCRIPT,dev/lint.sh)

# help
# 作用: 显示可用命令
# 用法: make help
.PHONY: help
help:
	@echo "yomi-admin-backend targets:"
	@echo "  make tidy        同步 go.work 并整理各 module 依赖"
	@echo "  make fmt         go fmt + gofumpt"
	@echo "  make vet         go vet"
	@echo "  make lint        golangci-lint"
	@echo "  make test        单元测试（排除集成测试目录）"
	@echo "  make check       fmt + vet + lint"
	@echo "  make cross-build 交叉编译编排层"
	@echo "  make docker-build 构建并导出 Docker 镜像"
	@echo "  make clean       删除 bin/"
