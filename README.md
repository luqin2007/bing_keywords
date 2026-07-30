# Bing Keywords

每日关键字服务 - 自动从多个数据源采集关键字，提供去重和不重复的 API 服务。

## 功能

- 每日返回不重复的关键字（10 天内不重复）
- 自动从以下数据源采集关键字，每条记录详细来源参数：
  - GitHub Trending 仓库名（随机语言过滤）
  - GitHub Trending 开发者名（随机语言过滤）
  - SourceForge 开源应用名（随机 OS + 分类过滤）
  - OpenRouter AI 模型名（随机 category / input_modalities / supported_parameters）
- 关键字 30 天后自动清理
- 支持 Webhook 通知（采集失败时告警，HMAC-SHA256 签名）
- 请求日志记录（文件 + stdout）
- 后台管理页面（仪表盘、关键字管理、日志查看、设置）
- 低内存占用（~15MB）

## 快速开始

```bash
# 构建
docker build -t bing-keywords .

# 运行
docker run -d -p 8080:8080 -v keywords-data:/data -v /path/to/config:/app/config bing-keywords

# 请求关键字
curl "http://localhost:8080/api/keywords?count=5"

# 打开管理页面
# 浏览器访问 http://localhost:8080/admin
```

## API

### GET /api/keywords

返回随机关键字列表。

参数：
- `count` (可选, 默认 5, 范围 1-100): 返回的关键字数量

响应示例：
```json
{
  "code": 200,
  "msg": "请求成功",
  "data": [
    { "title": "GPT-4" },
    { "title": "Rust" },
    { "title": "VSCode" }
  ]
}
```

## 管理页面

访问 `http://localhost:8080/admin` 打开管理后台，包含：

| 标签页 | 功能 |
|--------|------|
| **仪表盘** | KPI 卡片（总/可用/今日使用）、来源分布图、一键采集 |
| **关键字** | 关键字列表分页查看、按来源/状态筛选、搜索 |
| **日志** | 请求日志分页查看、按状态筛选 |
| **设置** | 在线编辑 config.json |

## 配置

配置文件位于 `/app/config/config.json`（首次启动自动生成）：

```json
{
  "port": 8080,
  "db_path": "/data/keywords.db",
  "log_file": "/data/requests.log",
  "webhook_url": "",
  "webhook_hmac": "",
  "admin_token": ""
}
```

- `webhook_url` / `webhook_hmac`：Webhook 通知地址和 HMAC 密钥
  - 请求头携带 `X-Signature-256: sha256=<HMAC-SHA256 hex>`
- `admin_token`：管理页面的访问令牌（空值表示无需认证）

## 技术栈

- Go 1.23
- SQLite（纯 Go 实现，无需 CGO）
- Docker（多阶段构建，alpine 运行时）
- 管理页面：Tailwind CSS + 原生 JS