# 概述

prometheus alert-manager接入企业微信机器人通知,使用Docker部署;使用前需要添加自己的wecomHookKey->main.go
![](process.png)

### 构建镜像

```shell
docker build --network host -t wecomhook:0.1 .
```

### 清理环境

```shell
docker image prune --filter label=stage=builder
```

### 本地打包镜像

```shell
docker save wecomhook:0.1 > wecomhook.tar
```

### 线上导入镜像

```shell
docker load < wecomhook.tar
```

### 启动容器

```shell
docker run --name wecomhook -p 6666:6666 -d wecomhook:0.1
```

### 最终效果

![](wecom.jpeg)