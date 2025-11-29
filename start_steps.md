## 完整启动步骤

修复后的项目需要按以下步骤启动:

### 1. 启动基础环境
```bash
make env-up
```

### 2. 推送配置到etcd（关键步骤）
```bash
docker exec etcd etcdctl put /config/config.yaml "$(cat config/config.yaml)"
```

### 3. 构建应用镜像
```bash
make docker 
```

### 4. 启动所有微服务
```bash
bash docker-run.sh
```

### 5. （可选）启动 Nginx 反代 + 演示页
```bash
bash scripts/run-nginx.sh
```

### 注意事项
- 如果遇到MySQL兼容性问题，可能需要清理数据目录：`rm -rf data/mysql/*`
- etcd配置推送是必需步骤，微服务从etcd读取配置而不是本地文件

## 终端代理小贴士

如果终端设置了 HTTP/HTTPS 代理，`curl http://localhost:10001/...` 可能被代理截断或无响应。可以在部署/调试会话开始时关闭代理并允许本地主机直连：

```bash
unset http_proxy https_proxy
export no_proxy=localhost,127.0.0.1,47.243.155.129
```

API测试:
```bash
curl "http://localhost:10001/ping"
curl "http://localhost:10001/douyin/feed/?latest_time=0"
```
