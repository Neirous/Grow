# grow 操作手册

## 目录
- [一、本地开发运行](#一本地开发运行)
- [二、新电脑快速部署（推荐）](#二新电脑快速部署推荐)
- [三、Docker 镜像发布（阿里云 CR）](#三docker-镜像发布阿里云-cr)
- [四、常用命令速查](#四常用命令速查)
- [五、数据备份与恢复](#五数据备份与恢复)
- [六、常见问题](#六常见问题)

---

## 一、本地开发运行

**前置条件**：Go 1.25+

```bash
cd /Users/linyuhui/Go/src/grow

# 启动服务
make run
# 或者
go run main.go

# 自定义端口
make run PORT=3000
```

浏览器打开 **http://localhost:8080**

---

## 二、新电脑快速部署（推荐）

**前置条件**：仅需安装 Docker

### 2.1 拉取并运行

```bash
# 拉取镜像
docker pull crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.0

# 创建数据目录
mkdir -p ~/grow-data

# 启动容器
docker run -d \
  --name grow \
  -p 8080:8080 \
  -v ~/grow-data:/data \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.0

# 确认运行状态
docker ps | grep grow
```

浏览器打开 **http://localhost:8080**

### 2.2 参数说明

| 参数 | 说明 | 示例 |
|------|------|------|
| `--name grow` | 容器名称 | grow |
| `-p 8080:8080` | 端口映射 | 宿主机:容器 |
| `-v ~/grow-data:/data` | 数据持久化 | 数据存宿主机不丢失 |
| `-e TZ=Asia/Shanghai` | 时区 | 中国标准时间 |
| `--restart unless-stopped` | 自动重启 | 开机自启（手动停除外） |

### 2.3 容器管理

```bash
# 查看日志
docker logs -f grow

# 停止
docker stop grow

# 启动
docker start grow

# 删除容器（数据在 ~/grow-data 里不会丢）
docker rm -f grow
```

---

## 三、Docker 镜像发布（阿里云 CR）

### 3.1 首次发布

```bash
cd /Users/linyuhui/Go/src/grow

# 1. 登录阿里云（需输入密码）
sudo docker login --username=aliyun3175536781 \
  crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com

# 2. 构建镜像
make docker TAG=v1.0

# 3. 推送
docker push crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.0
```

### 3.2 发布新版本

```bash
# 构建并推送（直接 push 不重新 build）
docker tag crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.0 \
           crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.1
docker push crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.1

# 或者走 Makefile
make docker TAG=v1.1
docker push crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.1
```

### 3.3 新电脑升级

```bash
# 拉取新版本
docker pull crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.1

# 停旧容器
docker stop grow && docker rm grow

# 用新版本启动（数据目录不变，数据不会丢失）
docker run -d \
  --name grow \
  -p 8080:8080 \
  -v ~/grow-data:/data \
  -e TZ=Asia/Shanghai \
  --restart unless-stopped \
  crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.1
```

---

## 四、常用命令速查

### Makefile

| 命令 | 说明 |
|------|------|
| `make run` | 本地启动（Go 开发模式） |
| `make build` | 编译二进制 |
| `make docker TAG=v1.0` | 构建 Docker 镜像 |
| `make docker-push TAG=v1.0` | 构建 + 推送到阿里云 |
| `make docker-run` | 本地 Docker 运行 |
| `make docker-stop` | 停止并删除容器 |
| `make docker-login` | 登录阿里云镜像仓库 |
| `make clean` | 清理编译产物和容器 |

### API 测试

```bash
# 创建能力
curl -X POST http://localhost:8080/api/abilities \
  -H 'Content-Type: application/json' \
  -d '{"name":"心肺功能","base_value":1000,"current_value":1000}'

# 创建活动
curl -X POST http://localhost:8080/api/activities \
  -H 'Content-Type: application/json' \
  -d '{"name":"跑步30min","effects":[{"ability_id":1,"boost_percentage":1.0}]}'

# 完成活动
curl -X POST http://localhost:8080/api/activities/1/complete \
  -H 'Content-Type: application/json' \
  -d '{"note":"今天跑得很爽"}'

# 查看能力
curl http://localhost:8080/api/abilities/1 | python3 -m json.tool
```

---

## 五、数据备份与恢复

所有数据存储在 SQLite 数据库文件中。

### 本地运行
数据库文件：`grow.db`（项目根目录）

```bash
# 备份
cp grow.db grow.db.backup.$(date +%Y%m%d)

# 恢复
cp grow.db.backup.20260611 grow.db
```

### Docker 运行
数据库文件：`~/grow-data/grow.db`

```bash
# 备份
cp ~/grow-data/grow.db ~/grow-data/grow.db.backup.$(date +%Y%m%d)

# 恢复
docker stop grow
cp ~/grow-data/grow.db.backup.20260611 ~/grow-data/grow.db
docker start grow
```

---

## 六、常见问题

### Q: 页面打开是空白的/卡住了？
刷新页面即可。如果仍有问题，查看浏览器控制台（F12 → Console）是否有报错。

### Q: 数据会丢失吗？
- **本地运行**：数据库文件 `grow.db` 在项目目录下，只要不删除就不会丢。
- **Docker 运行**：数据通过 `-v` 挂载到宿主机 `~/grow-data/`，删除容器不会丢数据。

### Q: 如何修改端口？
```bash
# 本地
make run PORT=3000

# Docker
docker run -d --name grow -p 3000:8080 -v ~/grow-data:/data crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:v1.0
```

### Q: 如何彻底重置数据？
```bash
# 本地
rm grow.db

# Docker
docker stop grow && docker rm grow
rm -rf ~/grow-data
mkdir -p ~/grow-data
# 重新 docker run...
```

### Q: Docker 拉取镜像很慢？
阿里云 CR 在国内网络下速度很快，如果还慢可以开 Docker 代理。

### Q: 如何查看容器是否正常运行？
```bash
docker logs grow
# 看到 "grow is running at http://localhost:8080" 就说明正常
```
