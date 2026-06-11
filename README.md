# grow - 个人能力成长追踪器

把个人成长数据化、游戏化。自定义你的各种能力（身体、编程、乐器...），设置活动（跑步、学习...），每次完成活动获得**复利增长**，不练则会**衰减**。直观看到自己的进步曲线。

## 快速开始

```bash
# 启动服务（需要 Go 1.25+）
make run

# 浏览器打开
open http://localhost:8080
```

## 使用指南

### 1. 创建能力

进入「能力管理」Tab → 点击「+ 新增能力」：

| 字段 | 说明 | 示例 |
|------|------|------|
| 名称 | 能力的名字 | 心肺功能 |
| 基础值 | 初始/最低值（不会衰减至此以下） | 1000 |
| 当前值 | 当前能力值 | 1000 |
| 成长倍率 | 对活动效果的放大倍率 | 1.0 |
| 日衰减率 | 每天衰减百分比（0=不衰减） | 0.5 |

### 2. 创建活动

进入「活动管理」Tab → 点击「+ 新增活动」：

- 设置活动名称（如"跑步30min"）
- **添加效果**：选择影响哪个能力、提升多少百分比
- 一个活动可以影响多个能力（如篮球：身体+1.5%，心肺+1%）

### 3. 完成活动

两种方式：
- **仪表盘**：在「快速完成」下拉选择活动，点「完成！」
- **活动管理**：点击活动旁边的「完成」按钮

每次完成，能力值按复利增长：
```
心肺 = 1000
跑步 +1% → 1010
跑步 +1% → 1020.1
跑步 +1% → 1030.301
...
```

### 4. 查看图表

在「能力管理」中点击能力的「详情」，可以看到**能力增长曲线图**，直观感受进步。

### 5. 衰减机制

如果你一段时间不练习某项能力，它的**有效值**会随时间衰减：
```
有效值 = 当前值 × (1 - 衰减率/100) ^ 距上次天数
```
衰减仅在展示时计算，不影响存储值。有效值不会低于基础值。

## Docker 部署

```bash
# 登录阿里云镜像仓库（首次）
make docker-login

# 本地构建并运行
make docker
make docker-run

# 推送到阿里云
make docker-push TAG=v1.0

# 停止容器
make docker-stop
```

### 手动 Docker 命令

```bash
# 构建
docker build -t grow .

# 运行（挂载当前目录的 data/ 作为数据卷）
docker run -d --name grow -p 8080:8080 -v $(pwd)/data:/data -e TZ=Asia/Shanghai grow

# 停止
docker rm -f grow
```

### 阿里云镜像仓库

```bash
# 完整地址
crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow

# 拉取并运行
docker pull crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:latest
docker run -d --name grow -p 8080:8080 -v $(pwd)/data:/data crpi-51pd4blge4jwd9y0.cn-hangzhou.personal.cr.aliyuncs.com/aliyun3175536781/grow:latest
```

## 配置项

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `-port` | `8080` | HTTP 服务端口 |
| `-db` | `grow.db` | SQLite 数据库文件路径 |

环境变量：
- `TZ=Asia/Shanghai` — 时区设置
- `DB_PATH=/data/grow.db` — Docker 容器中数据库路径

## 技术栈

| 层面 | 技术 |
|------|------|
| 后端 | Go 1.25+（标准库 net/http） |
| 数据库 | SQLite（modernc/sqlite，纯 Go） |
| 前端 | 原生 HTML/CSS/JS + Chart.js |
| 部署 | Docker / 阿里云容器镜像 |

## 开发

```bash
# 项目结构
grow/
  main.go              # 入口
  templates/           # HTML 模板
  static/              # CSS/JS 静态文件
  internal/
    db/                # 数据库初始化 + Schema
    models/            # 数据模型 + CRUD
    handlers/          # API 处理器
    service/           # 业务逻辑（复利/衰减）
  Makefile
  Dockerfile
  README.md
```

## License

MIT
