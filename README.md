# Bing Keywords

每日关键字服务 - 自动从多个数据源采集关键字，提供去重和不重复的 API 服务。

## 功能

- 每日返回不重复的关键字（10 天内不重复）
- 自动从以下数据源采集关键字：
  - GitHub Trending 仓库名
  - GitHub Trending 开发者名
  - SourceForge 开源应用名
  - OpenRouter AI 模型名
- 关键字 30 天后自动清理
- 支持 Webhook 通知（采集失败时告警）
- 请求日志记录（文件 + stdout）
- 低内存占用（~15MB）

## 快速开始

```bash
# 构建
docker build -t bing-keywords .

# 运行
docker run -d -p 8080:8080 -v keywords-data:/data -v /path/to/config:/app/config bing-keywords

# 请求
curl "http://localhost:8080/api/keywords?count=5"
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

## 配置

配置文件位于 `/app/config/config.json`（首次启动自动生成）：

```json
{
  "port": 8080,
  "db_path": "/data/keywords.db",
  "log_file": "/data/requests.log",
  "webhook_url": "",
  "webhook_hmac": ""
}
```

webhook 请求会携带 `X-Signature-256` 头，值为 `sha256=<HMAC-SHA256 hex>`，接收方可用相同密钥验证签名。
```

## 技术栈

- Go 1.23
- SQLite（纯 Go 实现，无需 CGO）
- Docker（多阶段构建，alpine 运行时）