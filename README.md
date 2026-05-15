# NoteFeed — 笔记推送系统

基于 Obsidian Vault 的 git 提交记录，自动提炼当日学习知识点，搜索相关技术文章，推送到邮箱。

## 工作原理

```
git commit (笔记) → post-commit hook 缓存 diff
         → 每日定时聚合 diffs → DeepSeek 提炼知识点
         → Google 搜索相关文章 → LLM 生成摘要 → 邮件推送
```

## 快速开始

```bash
# 安装
go install github.com/yllmis/note_feed/cmd/note_feed@latest

# 初始化（在项目目录执行，指定 vault 路径）
cd ~/go_projects/note_feed
note_feed init -v /path/to/your/obsidian/vault

# 设置环境变量
export DEEPSEEK_API_KEY=sk-xxx
export TAVILY_API_KEY=tvly-xxx    # 注册: https://tavily.com
export EMAIL_USER=your@email.com
export EMAIL_PASS=smtp_authorization_code
export EMAIL_TO=your@email.com

# 验证各环节
note_feed push test llm
note_feed push test search
note_feed push test email

# 手动触发每日推送
note_feed push daily
```

## 配置

`init` 命令在当前目录生成 `config.yaml`（已加入 `.gitignore`，不提交到仓库）。

| 配置项 | 说明 |
|--------|------|
| `vault_path` | Obsidian Vault 路径 |
| `llm.api_key` | DeepSeek API Key（通过 `DEEPSEEK_API_KEY` 设置） |
| `search.tavily.api_key` | Tavily Search API Key（通过 `TAVILY_API_KEY` 设置） |
| `push.email` | SMTP 邮件配置（QQ 邮箱需使用授权码） |

## 命令

| 命令 | 说明 |
|------|------|
| `init` | 初始化配置、数据库、安装 git hook |
| `cache-diff` | 缓存 commit diff（由 git hook 自动调用） |
| `push daily` | 聚合当日学习内容并推送 |
| `push test email` | 发送测试邮件 |
| `push test llm` | 测试 DeepSeek API 连接 |
| `push test search` | 测试 Google Custom Search |
| `config path` | 显示配置文件路径 |

## 技术栈

- **语言**: Go (单二进制)
- **数据库**: SQLite (modernc.org/sqlite，纯 Go 无 CGO)
- **LLM**: DeepSeek API
- **搜索**: Tavily Search API
- **推送**: SMTP 邮件
- **CLI**: cobra
