# NoteFeed — 笔记推送系统

基于 Obsidian Vault 的 git 提交记录，自动提炼当日学习知识点，搜索相关技术文章，推送到邮箱。

## 工作原理

```
Obsidian 笔记编辑 → git commit
         ↓
post-commit hook 捕获 diff → 存入 SQLite
         ↓
push daily 聚合当日 diffs → DeepSeek 提炼知识点
         ↓
按分类搜索 Tavily 推荐文章
         ↓
DeepSeek 生成文章摘要 → 组装 HTML 邮件 → SMTP 推送
```

## 功能特性

- **自动捕获**：通过 git post-commit hook 自动缓存每次提交的 diff，无需手动操作
- **知识点提炼**：调用 DeepSeek 从 diff 中提取学习知识点及其关键词、分类
- **按需搜索**：按提炼出的知识点分类搜索 Tavily，获取相关技术文章
- **AI 摘要**：DeepSeek 为每篇文章生成 1-2 句中文摘要
- **邮件推送**：HTML 格式邮件，按分类组织，包含文章摘要和链接
- **日期过滤**：支持指定日期或推送前一天的提交记录（配合 crontab 定时任务）
- **去重机制**：已推送的文章 URL 不会重复推送
- **零外部依赖**：纯 Go 编译，CGO_ENABLED=0，单二进制部署

## 快速开始

### 安装

```bash
# 克隆仓库
git clone https://github.com/yllmis/note_feed.git
cd note_feed

# 编译
go build -o note_feed ./cmd/note_feed
```

### 初始化

```bash
# 指定 Obsidian Vault 路径（必须为 git 仓库）
note_feed init -v /path/to/your/obsidian/vault
```

`init` 命令会自动完成以下操作：

1. 检测 Vault 是否为 git 仓库
2. 安装 post-commit hook，每次提交自动缓存 diff
3. 初始化 SQLite 数据库（`<vault>/data/push.db`）
4. 生成配置文件 `config.yaml`

### 配置

编辑 `config.yaml`，填入 API Key 和邮箱配置：

```yaml
llm:
  api_key: "sk-xxx"                    # DeepSeek API Key
  model: "deepseek-chat"               # 模型名称
  timeout: "30s"
  base_url: "https://api.deepseek.com"

search:
  tavily:
    api_key: "tvly-xxx"                # Tavily API Key（注册: https://tavily.com）

push:
  email:
    smtp_host: "smtp.qq.com"           # SMTP 服务器
    smtp_port: 587                     # 端口（587 支持 STARTTLS）
    username: "xxx@qq.com"             # 邮箱账号
    password: "授权码"                  # SMTP 授权码（非登录密码）
    from: "xxx@qq.com"                 # 发件人地址
    to: "xxx@qq.com"                   # 收件人地址
```

也可通过环境变量设置，配置文件中使用 `${VAR_NAME}` 引用：

```bash
export DEEPSEEK_API_KEY=sk-xxx
export TAVILY_API_KEY=tvly-xxx
export EMAIL_USER=xxx@qq.com
export EMAIL_PASS=smtp_authorization_code
export EMAIL_TO=xxx@qq.com
```

### 验证配置

```bash
# 测试 LLM API 连接
note_feed push test llm

# 测试 Tavily 搜索
note_feed push test search

# 发送测试邮件
note_feed push test email
```

### 手动推送

```bash
# 推送今日的学习记录
note_feed push daily

# 推送指定日期的记录
note_feed push daily --date 2024-01-15

# 推送前一天的记录（配合定时任务使用）
note_feed push daily --yesterday
```

## 定时推送

推荐使用系统 crontab 实现每天早上自动推送：

```bash
# 编辑 crontab
crontab -e

# 每天早上 9 点推送前一天的笔记
0 9 * * * /path/to/note_feed --config /path/to/config.yaml push daily --yesterday >> /tmp/note_feed.log 2>&1
```

查看推送日志：

```bash
cat /tmp/note_feed.log
```

## 命令参考

| 命令 | 说明 |
|------|------|
| `init` | 初始化配置、数据库、安装 git hook |
| `cache-diff` | 缓存 commit diff（由 git hook 自动调用） |
| `push daily` | 聚合当日学习内容并推送 |
| `push daily --yesterday` | 推送前一天的记录 |
| `push daily --date YYYY-MM-DD` | 推送指定日期的记录 |
| `push daily -v /path` | 指定 Vault 路径 |
| `push test email` | 发送测试邮件 |
| `push test llm` | 测试 DeepSeek API 连接 |
| `push test search` | 测试 Tavily Search API |
| `config path` | 显示配置文件路径 |
| `--config` | 指定配置文件路径（默认 `./config.yaml`） |

## 邮件效果

推送邮件按分类组织，格式如下：

- **邮件标题**：`笔记推送 — <日期> <分类1> <分类2> ...`
- **正文结构**：按分类分组，每篇文章包含标题（可点击链接）、AI 生成的摘要
- **编码**：中文标题使用 RFC 2047 编码，所有主流邮件客户端均正常显示

## 技术栈

- **语言**: Go
- **数据库**: SQLite（modernc.org/sqlite，纯 Go 无 CGO）
- **LLM**: DeepSeek API（OpenAI 兼容接口）
- **搜索**: Tavily Search API（AI 应用搜索引擎）
- **推送**: SMTP（net/smtp + STARTTLS）
- **CLI**: cobra
