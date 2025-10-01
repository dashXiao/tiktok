# TikTok项目部署修复记录

本文档记录了为使TikTok项目能够正常启动和运行而进行的关键修复。

## 修改文件列表

### 1. cmd/api/main.go
**问题**: API服务存在语法错误和流量控制配置问题
**修复**:
- 修复`func Init()`缺少开花括号的语法错误
- 将Sentinel流量控制阈值从0.0改为100.0，解决API被限流问题
- 添加缺失的函数结束花括号

### 2. docker-compose.yml
**问题**: MySQL版本兼容性问题
**修复**:
- 将MySQL镜像从`mysql:latest`改为`mysql:8.0`
- 解决了MySQL 9.x与项目的兼容性问题

### 3. docker-run.sh
**问题**: 容器网络配置不正确
**修复**:
- 从`--net=host`改为`--network=tiktok_tiktok`
- 为API服务添加端口映射`-p 10001:10001`
- 实现了正确的Docker容器间通信和外部API访问

### 4. config/config.yaml (新文件，被.gitignore忽略)
**说明**: 基于config_exmple.yaml创建的运行配置
**创建方法**: `cp config/config_exmple.yaml config/config.yaml`
**主要配置修改**:
- etcd地址: `127.0.0.1:2379` → `etcd:2379`
- 数据库连接: `127.0.0.1:3306` → `mysql:3306`
- Redis地址: `127.0.0.1:6379` → `redis:6379`
- API服务地址: `127.0.0.1:10001` → `0.0.0.0:10001`
- 微服务地址配置为Docker容器名
- OSS配置添加虚拟值避免服务崩溃

**注意**: 此文件被.gitignore忽略，需要手动创建

## 完整启动步骤

修复后的项目需要按以下步骤启动:

### 1. 启动基础环境
```bash
# 启动MySQL、Redis、etcd等基础服务
make env-up
```

### 2. 推送配置到etcd（关键步骤）
```bash
# 等待etcd启动完成，然后推送配置
docker exec etcd etcdctl put /config/config.yaml "$(cat config/config.yaml)"
```

### 3. 构建应用镜像
```bash
# 构建tiktok应用镜像
make docker 
```

### 4. 启动所有微服务
```bash
# 启动API网关和所有微服务
sh docker-run.sh
```

### 注意事项
- 如果遇到MySQL兼容性问题，可能需要清理数据目录：`rm -rf data/mysql/*`
- etcd配置推送是必需步骤，微服务从etcd读取配置而不是本地文件

API测试:
```bash
curl "http://localhost:10001/ping"
curl "http://localhost:10001/douyin/feed/?latest_time=0"
```

## 测试结果

✅ 所有核心功能测试通过:
- 用户注册/登录
- 视频feed流
- 点赞评论功能
- 关注聊天功能

项目现在完全可用！
