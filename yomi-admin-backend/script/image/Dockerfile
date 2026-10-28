# yomi-admin-backend 编排层 Dockerfile
# 用法（通过 Makefile）:
#   make docker-build
#   make docker-build TAG=v0.1.0

FROM alpine:3.21

RUN apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone

ARG BIN_PATH
ARG PORT

WORKDIR /app

COPY ${BIN_PATH} /app/server
COPY apps/server/config/ /app/config/

RUN chmod +x /app/server

EXPOSE ${PORT}

ENTRYPOINT ["/app/server"]
