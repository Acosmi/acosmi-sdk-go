# Acosmi Go SDK 开发手册

> 版本: v0.2.1 | 语言: Go 1.22+ | 许可证: MIT

---

## 目录

- [1. 概述](#1-概述)
- [2. 安装](#2-安装)
  - [2.1 作为 Go 库引入](#21-作为-go-库引入)
  - [2.2 构建 CLI 工具](#22-构建-cli-工具)
  - [2.3 通过 NPM 安装 CLI](#23-通过-npm-安装-cli)
- [3. 认证机制](#3-认证机制)
  - [3.1 OAuth 2.1 PKCE 流程](#31-oauth-21-pkce-流程)
  - [3.2 Scope 权限分组](#32-scope-权限分组)
  - [3.3 Token 存储](#33-token-存储)
  - [3.4 自定义 Token 存储](#34-自定义-token-存储)
- [4. SDK 客户端 API](#4-sdk-客户端-api)
  - [4.1 创建客户端](#41-创建客户端)
  - [4.2 授权生命周期](#42-授权生命周期)
  - [4.3 AI 模型服务](#43-ai-模型服务)
  - [4.3.1 模型能力查询](#431-模型能力查询)
  - [4.3.2 扩展聊天 (CrabCode)](#432-扩展聊天-crabcode)
  - [4.3.3 联网搜索 (Server Tool)](#433-联网搜索-server-tool)
  - [4.3.4 Beta Header 自动组装](#434-beta-header-自动组装)
  - [4.4 权益管理](#44-权益管理)
  - [4.5 流量包商城](#45-流量包商城)
  - [4.6 钱包](#46-钱包)
  - [4.7 技能商店](#47-技能商店)
  - [4.8 技能认证](#48-技能认证)
  - [4.9 AI 技能生成器](#49-ai-技能生成器)
  - [4.10 统一工具列表](#410-统一工具列表)
  - [4.11 WebSocket 实时推送](#411-websocket-实时推送)
- [5. CrabClaw-Skill CLI 命令手册](#5-crabclaw-skill-cli-命令手册)
  - [5.1 全局参数](#51-全局参数)
  - [5.2 login — 登录](#52-login--登录)
  - [5.3 logout — 登出](#53-logout--登出)
  - [5.4 whoami — 查看登录状态](#54-whoami--查看登录状态)
  - [5.5 version — 版本信息](#55-version--版本信息)
  - [5.6 config — 配置管理](#56-config--配置管理)
  - [5.7 search — 搜索技能](#57-search--搜索技能)
  - [5.8 list — 已安装技能](#58-list--已安装技能)
  - [5.9 info — 技能详情](#59-info--技能详情)
  - [5.10 download — 下载技能](#510-download--下载技能)
  - [5.11 install — 安装技能](#511-install--安装技能)
  - [5.12 upload — 上传技能](#512-upload--上传技能)
  - [5.13 generate — AI 生成技能](#513-generate--ai-生成技能)
  - [5.14 certify — 触发认证](#514-certify--触发认证)
- [6. 数据类型参考](#6-数据类型参考)
  - [6.1 OAuth 类型](#61-oauth-类型)
  - [6.2 模型类型](#62-模型类型)
  - [6.2.1 Chat 扩展类型](#621-chat-扩展类型)
  - [6.3 权益类型](#63-权益类型)
  - [6.4 商城/订单类型](#64-商城订单类型)
  - [6.5 钱包类型](#65-钱包类型)
  - [6.6 技能类型](#66-技能类型)
  - [6.7 工具类型](#67-工具类型)
  - [6.8 WebSocket 类型](#68-websocket-类型)
  - [6.9 错误类型](#69-错误类型)
  - [6.10 通用响应包装](#610-通用响应包装)
- [7. 完整示例](#7-完整示例)
- [8. 安全特性](#8-安全特性)
- [9. 项目结构](#9-项目结构)
- [10. 构建与发布](#10-构建与发布)
- [11. 常见问题](#11-常见问题)
- [12. 版本修订记录](#12-版本修订记录)

---

## 1. 概述

Acosmi Go SDK 是 Acosmi 平台的官方 Go 语言客户端库，提供两种使用方式:

1. **Go 库** — 在你的 Go 应用中 `import` 引入，通过类型安全的 API 访问 Acosmi 平台全部功能
2. **CrabClaw-Skill CLI** — 命令行工具，用于技能的搜索、安装、上传、AI 生成和认证管理

### 核心特性

| 特性 | 说明 |
|------|------|
| OAuth 2.1 PKCE | 安全桌面授权流程，自动 token 刷新 |
| 统一客户端 | 一个 `Client` 对象覆盖全域 API |
| 流式聊天 | 基于 SSE (Server-Sent Events) 的实时对话 |
| **Beta 自动组装** | 根据模型能力 + 请求参数自动注入 11 项 beta header (v0.2.0) |
| **Server Tool** | 联网搜索等服务端工具，SDK 自动合入请求体 (v0.2.0) |
| **模型能力查询** | `ModelCapabilities` 矩阵，16 项能力标记 (v0.2.0) |
| WebSocket 长连接 | 实时接收余额/技能/系统推送，自动断线重连 |
| 技能商店 | 浏览/搜索/下载/安装/上传/AI 生成 |
| 线程安全 | 所有 API 调用均通过 `sync.RWMutex` 保护 |
| 跨平台 | 支持 macOS / Linux / Windows (amd64 / arm64) |

### 依赖

| 库 | 版本 | 用途 |
|----|------|------|
| `github.com/fatih/color` | v1.18.0 | CLI 彩色终端输出 |
| `github.com/gorilla/websocket` | v1.5.3 | WebSocket 客户端 |
| `github.com/spf13/cobra` | v1.10.2 | CLI 命令行框架 |

无 CGO 依赖，可纯静态编译。

### 服务地址

| 环境 | 地址 | 说明 |
|------|------|------|
| **生产环境** | `https://acosmi.ai` | SDK 和 CLI 的默认地址，零配置即可使用 |
| 本地开发 | `http://127.0.0.1:3300` | 本地 nginx 统一入口，需手动配置 |

SDK 会自动在地址末尾追加 `/api/v4`，无需手动拼接。

**切换到本地开发环境**（三种方式，优先级从高到低）:

```bash
# 1. 环境变量（推荐）
export ACOSMI_SERVER_URL=http://127.0.0.1:3300

# 2. CLI 参数
crabclaw-skill --server http://127.0.0.1:3300 <命令>

# 3. 配置文件
crabclaw-skill config set server http://127.0.0.1:3300
```

Go 库使用时通过 `Config.ServerURL` 指定:

```go
client, _ := acosmi.NewClient(acosmi.Config{
    ServerURL: "http://127.0.0.1:3300", // 本地开发
})
```

---

## 2. 安装

### 2.1 作为 Go 库引入

```bash
go get github.com/acosmi/acosmi-sdk-go
```

在代码中引入:

```go
import acosmi "github.com/acosmi/acosmi-sdk-go"
```

**要求**: Go 1.22.0 或更高版本。

### 2.2 构建 CLI 工具

```bash
git clone https://github.com/acosmi/acosmi-sdk-go.git
cd acosmi-sdk-go

# 构建当前平台
make build
# 输出: bin/crabclaw-skill

# 构建全平台
make build-all
# 输出: dist/crabclaw-skill-{darwin,linux,windows}-{amd64,arm64}

# 安装到 $GOPATH/bin
make install
```

### 2.3 通过 NPM 安装 CLI

```bash
npm install -g @acosmi/crabclaw-skill
```

NPM 包会在安装后自动下载对应平台的预编译 Go 二进制文件。

可通过环境变量 `CRABCLAW_SKILL_BINARY_PATH` 指定自定义二进制路径。

---

## 3. 认证机制

### 3.1 OAuth 2.1 PKCE 流程

SDK 使用标准 OAuth 2.1 Authorization Code + PKCE 流程进行桌面端安全认证:

```
┌──────────┐    ┌────────────┐    ┌──────────────┐
│  你的应用  │    │  本地 HTTP  │    │  Acosmi 服务  │
│          │    │  回调服务器  │    │              │
└────┬─────┘    └─────┬──────┘    └──────┬───────┘
     │                │                   │
     │ 1. Discover    │                   │
     │ ───────────────────────────────>   │
     │   GET /.well-known/oauth-...       │
     │ <──────────── ServerMetadata ──    │
     │                │                   │
     │ 2. Register    │                   │
     │ ───────────────────────────────>   │
     │   POST /register (动态客户端)      │
     │ <──────────── client_id ────────   │
     │                │                   │
     │ 3. 启动回调服务 │                   │
     │ ─────────────> │                   │
     │   localhost:随机端口               │
     │                │                   │
     │ 4. 打开浏览器   │                   │
     │ ───────────────────────────────>   │
     │   /authorize?code_challenge=...    │
     │                │                   │
     │                │ 5. 用户授权后回调   │
     │                │ <──── ?code=xxx   │
     │                │                   │
     │ 6. ExchangeCode│                   │
     │ ───────────────────────────────>   │
     │   POST /token (code + verifier)    │
     │ <──── access_token + refresh ──    │
     │                │                   │
     │ 7. 保存 token   │                   │
     │ (自动刷新)      │                   │
     └────────────────┴───────────────────┘
```

**关键安全细节**:
- PKCE 使用 S256 方法 (SHA-256 哈希 + Base64URL 编码)
- 本地回调服务器使用 `127.0.0.1` 随机端口 (符合 RFC 8252)
- Token 文件权限 `0600`，目录权限 `0700`
- Access Token 有效期 15 分钟，Refresh Token 有效期 7 天
- Token 过期前 30 秒自动触发刷新

### 3.2 Scope 权限分组

SDK 使用 3 个分组 Scope（从 V1 的 10 个细粒度 scope 合并而来）:

| Scope | 常量 | 权限范围 |
|-------|------|----------|
| `ai` | `acosmi.ScopeAI` | 模型调用 + 流量包 + 权益查询 |
| `skills` | `acosmi.ScopeSkills` | 技能商店 + 工具列表 + 执行 |
| `account` | `acosmi.ScopeAccount` | 个人资料 + 钱包余额 + 交易记录 |

**预设 Scope 组合** (每次调用返回新切片，防止外部篡改):

```go
acosmi.AllScopes()      // ["ai", "skills", "account"] — 推荐
acosmi.ModelScopes()    // ["ai"]
acosmi.CommerceScopes() // ["ai", "account"]
acosmi.SkillScopes()    // ["skills"]
```

> 旧版细粒度 scope (`models`, `models:chat`, `entitlements` 等) 仍可使用，但已标记 `Deprecated`，新代码应使用分组 scope。

### 3.3 Token 存储

默认使用文件存储，token 保存在 `~/.acosmi/tokens.json`:

```json
{
  "access_token": "eyJhbGci...",
  "refresh_token": "dGhpcyBp...",
  "expires_at": "2026-04-04T12:00:00Z",
  "scope": "ai skills account",
  "client_id": "abc123",
  "server_url": "https://acosmi.ai/api/v4"
}
```

**安全须知**: 默认文件存储适用于开发/测试环境。生产桌面应用建议实现系统钥匙串存储。

### 3.4 自定义 Token 存储

实现 `TokenStore` 接口即可替换默认存储:

```go
type TokenStore interface {
    Save(tokens *TokenSet) error
    Load() (*TokenSet, error)
    Clear() error
}
```

**示例: macOS Keychain 存储**

```go
type KeychainStore struct {
    Service string
}

func (k *KeychainStore) Save(tokens *acosmi.TokenSet) error {
    data, _ := json.Marshal(tokens)
    // 调用 macOS Keychain API 保存
    return keychain.Set(k.Service, "acosmi-tokens", data)
}

func (k *KeychainStore) Load() (*acosmi.TokenSet, error) {
    data, err := keychain.Get(k.Service, "acosmi-tokens")
    if err != nil { return nil, nil }
    var tokens acosmi.TokenSet
    json.Unmarshal(data, &tokens)
    return &tokens, nil
}

func (k *KeychainStore) Clear() error {
    return keychain.Delete(k.Service, "acosmi-tokens")
}

// 使用
client, _ := acosmi.NewClient(acosmi.Config{
    ServerURL: "https://acosmi.ai",
    Store:     &KeychainStore{Service: "com.acosmi.desktop"},
})
```

---

## 4. SDK 客户端 API

### 4.1 创建客户端

```go
client, err := acosmi.NewClient(acosmi.Config{
    // 必填: API 服务地址 (生产环境: https://acosmi.ai，SDK 自动追加 /api/v4)
    ServerURL: "https://acosmi.ai",

    // 可选: 自定义 token 存储 (默认: ~/.acosmi/tokens.json)
    Store: nil,

    // 可选: 自定义 HTTP 客户端 (默认: 无全局超时的 http.Client)
    HTTPClient: nil,
})
```

`NewClient` 会自动从 `TokenStore` 加载已保存的 token，无需手动调用。

> **注意**: `HTTPClient` 不设全局 `Timeout`，这是有意为之。全局超时会截断 SSE 流式聊天和大文件下载。所有 API 调用应通过 `context.Context` 控制超时。

### 4.2 授权生命周期

#### 检查授权状态

```go
if client.IsAuthorized() {
    fmt.Println("已登录")
}
```

#### 登录

```go
ctx := context.Background()

err := client.Login(ctx, "我的应用", acosmi.AllScopes())
// 会自动:
// 1. 发现 OAuth 服务端点
// 2. 动态注册客户端
// 3. 打开浏览器进行用户授权
// 4. 接收回调并交换 token
// 5. 保存 token 到 Store
```

参数说明:
- `appName` — 应用名称，显示在授权确认页面
- `scopes` — 请求的权限范围

#### 登出

```go
err := client.Logout(ctx)
// 会自动:
// 1. 吊销服务端 token
// 2. 清除本地存储
```

#### 获取 Token 信息

```go
tokenSet := client.GetTokenSet()
if tokenSet != nil {
    fmt.Printf("过期时间: %v\n", tokenSet.ExpiresAt)
    fmt.Printf("是否过期: %v\n", tokenSet.IsExpired())
    fmt.Printf("授权范围: %s\n", tokenSet.Scope)
}
```

### 4.3 AI 模型服务

> 需要 scope: `ai`

#### 获取可用模型列表

```go
models, err := client.ListModels(ctx)
for _, m := range models {
    fmt.Printf("模型: %s (%s/%s)\n", m.Name, m.Provider, m.ModelID)
    fmt.Printf("  上下文窗口: %d | 最大输出: %d\n", m.ContextWindow, m.MaxTokens)
    fmt.Printf("  价格: %.2f / 百万 Token\n", m.PricePerMTok)
}
```

#### 同步聊天

```go
resp, err := client.Chat(ctx, modelID, acosmi.ChatRequest{
    Messages: []acosmi.ChatMessage{
        {Role: "system", Content: "你是一个有帮助的助手"},
        {Role: "user", Content: "Go 语言的优势是什么？"},
    },
    MaxTokens: 1024,
})
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.Choices[0].Message.Content)
fmt.Printf("Token 消耗: 输入 %d + 输出 %d = 总计 %d\n",
    resp.Usage.PromptTokens,
    resp.Usage.CompletionTokens,
    resp.Usage.TotalTokens)

// 结算后余额 (来自响应 Header, -1 表示服务端未返回)
if resp.TokenRemaining >= 0 {
    fmt.Printf("剩余: %d token / %d 次调用\n",
        resp.TokenRemaining, resp.CallRemaining)
}
```

#### 流式聊天 (SSE) — 推荐使用 ChatStreamWithUsage

`ChatStreamWithUsage` 自动解析控制事件，返回三个 channel:
- `contentCh` — 纯内容增量事件 (已过滤 started/settled/failed)
- `settleCh` — 结算信息 (token 消耗 + 剩余余额)
- `errCh` — 传输错误或服务端 failed 事件

```go
contentCh, settleCh, errCh := client.ChatStreamWithUsage(ctx, modelID, acosmi.ChatRequest{
    Messages: []acosmi.ChatMessage{
        {Role: "user", Content: "写一首关于编程的诗"},
    },
    MaxTokens: 512,
})

// 实时读取内容 (无需手动过滤控制事件)
for event := range contentCh {
    var chunk struct {
        Choices []struct {
            Delta struct {
                Content string `json:"content"`
            } `json:"delta"`
        } `json:"choices"`
    }
    json.Unmarshal([]byte(event.Data), &chunk)
    if len(chunk.Choices) > 0 {
        fmt.Print(chunk.Choices[0].Delta.Content)
    }
}

// 读取结算信息 (token 消耗 + 剩余余额)
if settle, ok := <-settleCh; ok {
    fmt.Printf("\n消耗: %d token (输入 %d + 输出 %d)\n",
        settle.TotalTokens, settle.InputTokens, settle.OutputTokens)
    if settle.TokenRemaining >= 0 {
        fmt.Printf("剩余: %d token / %d 次调用\n",
            settle.TokenRemaining, settle.CallRemaining)
    }
}

// 检查错误 (传输错误或服务端 failed 事件)
if err := <-errCh; err != nil {
    log.Fatal(err)
}
```

#### 流式聊天 (SSE) — 低级 API

`ChatStream` 返回原始事件流，调用方需自行处理控制事件:

```go
eventCh, errCh := client.ChatStream(ctx, modelID, acosmi.ChatRequest{
    Messages: []acosmi.ChatMessage{
        {Role: "user", Content: "写一首关于编程的诗"},
    },
    MaxTokens: 512,
})

for event := range eventCh {
    // event.Event: "started" | "settled" | "pending_settle" | "failed" | "" (数据块)

    // 可用 ParseSettlement 解析结算事件
    if s := acosmi.ParseSettlement(event); s != nil {
        fmt.Printf("消耗: %d token, 剩余: %d\n", s.TotalTokens, s.TokenRemaining)
        continue
    }

    if event.Event == "started" || event.Event == "failed" {
        continue
    }

    var chunk struct {
        Choices []struct {
            Delta struct {
                Content string `json:"content"`
            } `json:"delta"`
        } `json:"choices"`
    }
    json.Unmarshal([]byte(event.Data), &chunk)
    if len(chunk.Choices) > 0 {
        fmt.Print(chunk.Choices[0].Delta.Content)
    }
}

if err := <-errCh; err != nil {
    log.Fatal(err)
}
```

#### 4.3.1 模型能力查询

> v0.2.0 新增

`ListModels()` 返回的 `ManagedModel` 包含 `Capabilities` 字段，描述模型支持的特性矩阵。下游应用可据此决定 UI 功能开关。

```go
models, _ := client.ListModels(ctx)
for _, m := range models {
    caps := m.Capabilities
    fmt.Printf("模型: %s\n", m.Name)
    fmt.Printf("  思考: %v | 1M上下文: %v | 搜索: %v | Fast: %v\n",
        caps.SupportsThinking, caps.Supports1MContext,
        caps.SupportsWebSearch, caps.SupportsFastMode)
    fmt.Printf("  输入上限: %d | 输出上限: %d\n",
        caps.MaxInputTokens, caps.MaxOutputTokens)
}
```

也可以通过便捷方法查询单个模型的能力（内部复用 ListModels 缓存，5 分钟 TTL）:

```go
caps, err := client.GetModelCapabilities(ctx, "claude-opus-4-6")
if caps.SupportsWebSearch {
    // 显示搜索按钮
}
```

> **重要**: `GetModelCapabilities` 和 `Chat/ChatStream` 的 Beta 自动组装均依赖模型缓存。建议在应用启动时调用一次 `ListModels()` 填充缓存。

#### 4.3.2 扩展聊天 (CrabCode)

> v0.2.0 新增

`ChatRequest` 新增了多个扩展字段，供 CrabCode 等高级下游使用。所有扩展字段零值不改变行为，CrabClaw 等基础下游无需修改。

```go
resp, err := client.Chat(ctx, modelID, acosmi.ChatRequest{
    // 基础字段 (同 v0.1.0)
    Messages:  []acosmi.ChatMessage{{Role: "user", Content: "分析这段代码"}},
    MaxTokens: 4096,

    // 扩展: 复杂消息体 (多模态 content blocks)
    // RawMessages 非 nil 时优先于 Messages
    RawMessages: []map[string]interface{}{
        {"role": "user", "content": []map[string]interface{}{
            {"type": "text", "text": "描述这张图片"},
            {"type": "image", "source": map[string]interface{}{
                "type": "base64", "media_type": "image/png", "data": "iVBOR...",
            }},
        }},
    },

    // 扩展: 思考配置
    Thinking: &acosmi.ThinkingConfig{
        Type:         "enabled",
        BudgetTokens: 8192,
    },

    // 扩展: 推理努力级别 (low/medium/high/max)
    Effort: &acosmi.EffortConfig{Level: "high"},

    // 扩展: Fast Mode (Opus 4.6)
    Speed: "fast",

    // 扩展: 结构化输出
    OutputConfig: &acosmi.OutputConfig{
        Format: "json_schema",
        Schema: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "answer": map[string]string{"type": "string"},
            },
        },
    },

    // 扩展: 显式 Beta (SDK 自动合并，通常无需手动指定)
    Betas: []string{"my-custom-beta-2026-01-01"},

    // 扩展: 透传任意字段到请求体
    ExtraBody: map[string]interface{}{
        "custom_field": "custom_value",
    },
})
```

**扩展字段一览**:

| 字段 | 类型 | 说明 |
|------|------|------|
| `RawMessages` | `interface{}` | 复杂消息体 (多模态)，非 nil 时优先于 Messages |
| `System` | `interface{}` | 系统提示 (string 或 content blocks) |
| `Tools` | `interface{}` | 标准工具定义 (与 ServerTools 合并) |
| `Temperature` | `*float64` | 采样温度 |
| `Thinking` | `*ThinkingConfig` | 思考配置 (adaptive/enabled/disabled) |
| `Metadata` | `map[string]string` | 请求元数据 |
| `Betas` | `[]string` | 显式 beta header (SDK 自动合并去重) |
| `ServerTools` | `[]ServerTool` | 服务端工具 (搜索等) |
| `Speed` | `string` | `""` 或 `"fast"` (Fast Mode) |
| `Effort` | `*EffortConfig` | 推理努力级别 |
| `OutputConfig` | `*OutputConfig` | 结构化输出配置 |
| `ExtraBody` | `map[string]interface{}` | 透传任意字段 |

> **设计说明**: 所有扩展字段标记为 `json:"-"`，仅通过内部 `buildChatRequest` 序列化，防止直接 `json.Marshal` 产生不完整输出。

#### 4.3.3 联网搜索 (Server Tool)

> v0.2.0 新增

联网搜索通过 Server Tool 模式工作 — SDK 将搜索工具合入请求体的 `tools` 数组，平台端执行搜索并以 `web_search_tool_result` 流事件返回结果。

```go
// 创建搜索工具 (带配置)
searchTool, err := acosmi.NewWebSearchTool(&acosmi.WebSearchConfig{
    MaxUses:        5,                      // 每请求最多搜索 5 次
    AllowedDomains: []string{"golang.org"}, // 仅搜索 golang.org
    UserLocation:   &acosmi.GeoLoc{Country: "CN"},
})
if err != nil {
    log.Fatal(err) // AllowedDomains 和 BlockedDomains 互斥
}

// 创建搜索工具 (默认配置)
defaultSearch, _ := acosmi.NewWebSearchTool(nil)

// 发起带搜索的聊天
eventCh, errCh := client.ChatStream(ctx, modelID, acosmi.ChatRequest{
    Messages:    []acosmi.ChatMessage{{Role: "user", Content: "Go 1.23 有什么新特性？"}},
    ServerTools: []acosmi.ServerTool{searchTool},
    MaxTokens:   4096,
})

for event := range eventCh {
    // event.Data 中可能包含 web_search_tool_result 类型的 content block
    // 下游自行解析搜索结果
    fmt.Println(event.Data)
}
```

**预定义常量**:

```go
acosmi.ServerToolTypeWebSearch  // "web_search_20250305" — 联网搜索 (Brave Search)
```

#### 4.3.4 Beta Header 自动组装

> v0.2.0 新增

SDK 在每次 `Chat` / `ChatStream` 调用时，根据模型能力矩阵和请求参数自动注入适用的 Beta Header。下游无需手动管理。

**自动注入规则** (11 项真实 beta，经联网验证):

| Beta Header | 注入条件 |
|---|---|
| `claude-code-20250219` | 始终 |
| `interleaved-thinking-2025-05-14` | 模型支持 ISP |
| `context-management-2025-06-27` | 模型支持 ISP |
| `context-1m-2025-08-07` | 模型支持 1M 上下文 |
| `structured-outputs-2025-11-13` | 模型支持 + `OutputConfig` 非 nil |
| `token-efficient-tools-2025-02-19` | 模型支持 + `OutputConfig` 为 nil (与 structured-outputs 互斥) |
| `advanced-tool-use-2025-11-20` | 模型支持 Tool Search |
| `effort-2025-11-24` | 模型支持 + `Effort` 非 nil |
| `fast-mode-2026-02-01` | 模型支持 + `Speed == "fast"` |
| `prompt-caching-scope-2026-01-05` | 模型支持 Prompt Cache |
| `redact-thinking-2026-02-12` | 模型支持 + `Thinking.Display == "summary"` |

**互斥规则**: `structured-outputs` 与 `token-efficient-tools` 互斥，API 拒绝同时存在。SDK 自动处理。

**客户端显式 beta**: 通过 `ChatRequest.Betas` 传入的 beta 会与自动注入的合并去重。

### 4.4 权益管理

> 需要 scope: `ai`

#### 查询余额 (聚合)

```go
balance, err := client.GetBalance(ctx)
fmt.Printf("Token: %d / %d (剩余 %d)\n",
    balance.TotalTokenUsed, balance.TotalTokenQuota, balance.TotalTokenRemaining)
fmt.Printf("调用: %d / %d (剩余 %d)\n",
    balance.TotalCallUsed, balance.TotalCallQuota, balance.TotalCallRemaining)
fmt.Printf("活跃权益: %d 条\n", balance.ActiveEntitlements)
```

#### 查询余额 (含明细)

```go
detail, err := client.GetBalanceDetail(ctx)
for _, e := range detail.Entitlements {
    fmt.Printf("  [%s] %s: Token %d/%d, 到期: %s\n",
        e.Type, e.Status, e.TokenUsed, e.TokenQuota, *e.ExpiresAt)
}
```

#### 列出权益列表

```go
// 列出活跃权益
active, err := client.ListEntitlements(ctx, "active")

// 列出已过期权益
expired, err := client.ListEntitlements(ctx, "expired")

// 列出全部权益 (空字符串)
all, err := client.ListEntitlements(ctx, "")
```

#### 查询消费记录

```go
records, err := client.ListConsumeRecords(ctx, 1, 20) // 第1页，每页20条
fmt.Printf("总计 %d 条消费记录\n", records.Total)
for _, r := range records.Records {
    fmt.Printf("  %s: 模型 %s, 消耗 %d Token, 状态 %s\n",
        r.CreatedAt, r.ModelID, r.TokensConsumed, r.Status)
}
```

### 4.5 流量包商城

> 需要 scope: `ai`

#### 浏览流量包

```go
packages, err := client.ListTokenPackages(ctx)
for _, p := range packages {
    fmt.Printf("%s: %d Token / %s 元 / 有效期 %d 天\n",
        p.Name, p.TokenQuota, p.Price, p.ValidDays)
}
```

#### 查看流量包详情

```go
pkg, err := client.GetTokenPackageDetail(ctx, "package-id-xxx")
fmt.Printf("名称: %s\n描述: %s\n价格: %s 元\n", pkg.Name, pkg.Description, pkg.Price)
```

#### 下单购买

```go
order, err := client.BuyTokenPackage(ctx, "package-id-xxx", &acosmi.PayPayload{
    PayMethod: "alipay", // 或 "wechat"
})
fmt.Printf("订单号: %s\n支付链接: %s\n", order.ID, order.PayURL)
```

#### 查询订单状态

```go
status, err := client.GetOrderStatus(ctx, "order-id-xxx")
fmt.Printf("订单状态: %s\n", status.Status) // pending | paid | expired | cancelled
```

#### 我的订单列表

```go
orders, err := client.ListMyOrders(ctx)
for _, o := range orders {
    fmt.Printf("[%s] %s: %s 元 (%s)\n", o.Status, o.PackageName, o.Amount, o.CreatedAt)
}
```

### 4.6 钱包

> 需要 scope: `account`

#### 钱包统计

```go
stats, err := client.GetWalletStats(ctx)
fmt.Printf("余额: %s 元\n", stats.Balance)
fmt.Printf("本月消费: %s 元\n", stats.MonthlyConsumption)
fmt.Printf("本月充值: %s 元\n", stats.MonthlyRecharge)
fmt.Printf("交易笔数: %d\n", stats.TransactionCount)
```

> 金额字段使用 `json.Number` 类型，避免浮点精度丢失（金融安全）。

#### 交易记录

```go
txns, err := client.GetWalletTransactions(ctx)
for _, tx := range txns {
    fmt.Printf("[%s] %s: %s 元 - %s\n", tx.Type, tx.CreatedAt, tx.Amount, tx.Remark)
}
```

### 4.7 技能商店

技能商店 API 分为**公开端点**（无需登录）和**认证端点**。

#### 快速浏览 (公开)

```go
skills, err := client.BrowseSkillStore(ctx, acosmi.SkillStoreQuery{
    Category: "ACTION",   // 可选: ACTION, PERSONA, WORKFLOW 等
    Keyword:  "翻译",      // 可选: 搜索关键词
    Tag:      "language",  // 可选: 标签过滤
})
for _, s := range skills {
    fmt.Printf("%s [%s] by %s — 下载 %d 次\n", s.Name, s.Category, s.Author, s.DownloadCount)
}
```

#### 分页浏览 (公开)

```go
// 完整浏览（包含全部字段）
resp, err := client.BrowseSkills(ctx, 1, 20, "ACTION", "", "", "")
fmt.Printf("共 %d 个技能，当前第 %d 页\n", resp.Total, resp.Page)

// 轻量浏览（响应体积缩减 90%+，适合列表页）
listResp, err := client.BrowseSkillsList(ctx, 1, 20, "", "搜索", "", "")
for _, s := range listResp.Items {
    fmt.Printf("%s — %s\n", s.Name, s.Description)
}
```

#### 技能详情 (公开)

```go
// 通过 ID 查询
skill, err := client.GetSkillDetail(ctx, "skill-id-xxx")

// 通过 Key 查询
skill, err := client.ResolveSkill(ctx, "translate-api")

fmt.Printf("名称: %s\n", skill.Name)
fmt.Printf("描述: %s\n", skill.Description)
fmt.Printf("版本: %s\n", skill.Version)
fmt.Printf("标签: %v\n", skill.Tags)
fmt.Printf("安全评分: %d (%s)\n", skill.SecurityScore, skill.SecurityLevel)
fmt.Printf("README:\n%s\n", skill.Readme)
```

#### 下载技能 ZIP (公开，有限流)

```go
zipData, filename, err := client.DownloadSkill(ctx, "skill-id-xxx")
if err != nil {
    // 匿名用户: 2次/小时; 登录用户: 无限制
    var rateErr *acosmi.RateLimitError
    if errors.As(err, &rateErr) {
        fmt.Printf("限流中，请 %s 后重试\n", rateErr.RetryAfter)
        return
    }
    log.Fatal(err)
}
os.WriteFile(filename, zipData, 0644) // filename 由服务端返回
```

#### 安装技能 (需登录)

```go
installed, err := client.InstallSkill(ctx, "skill-id-xxx")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("已安装: %s v%s\n", installed.Name, installed.Version)
```

#### 上传技能 (需登录)

```go
zipData, _ := os.ReadFile("my-skill.zip")

uploaded, err := client.UploadSkill(ctx, zipData,
    "PUBLIC",         // scope: "TENANT" (私有) 或 "PUBLIC" (公开)
    "PUBLIC_INTENT",  // intent: "PERSONAL" (私有) 或 "PUBLIC_INTENT" (公开)
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("已上传: %s (%s)\n", uploaded.Name, uploaded.Key)
```

#### 技能统计概览 (需登录)

```go
summary, err := client.GetSkillSummary(ctx)
fmt.Printf("已安装: %d | 已创建: %d | 商店总数: %d\n",
    summary.Installed, summary.Created, summary.StoreAvailable)
```

### 4.8 技能认证

> 需要 scope: `skills`

#### 触发认证流水线

```go
err := client.CertifySkill(ctx, "skill-id-xxx")
// 认证为异步操作，通过 GetCertificationStatus 查询结果
```

#### 查询认证状态

```go
cert, err := client.GetCertificationStatus(ctx, "skill-id-xxx")
fmt.Printf("认证状态: %s\n", cert.CertificationStatus)
// pending | in_progress | certified | rejected
fmt.Printf("安全等级: %s (评分: %d)\n", cert.SecurityLevel, cert.SecurityScore)
```

### 4.9 AI 技能生成器

> 需要 scope: `skills`

#### 生成新技能

```go
result, err := client.GenerateSkill(ctx, acosmi.GenerateSkillRequest{
    Purpose:     "一个将中文翻译成英文的工具",
    Examples:    []string{"输入: 你好世界 → 输出: Hello World"},
    InputHints:  "中文文本",
    OutputHints: "英文翻译结果",
    Category:    "ACTION",
    Language:    "zh",
})

fmt.Printf("生成的技能: %s (%s)\n", result.SkillName, result.SkillKey)
fmt.Printf("描述: %s\n", result.Description)
fmt.Printf("输入 Schema: %s\n", result.InputSchema)
fmt.Printf("测试用例: %v\n", result.TestCases)
```

#### 优化现有技能

```go
optimized, err := client.OptimizeSkill(ctx, acosmi.OptimizeSkillRequest{
    SkillName:    "translate-api",
    Description:  "翻译工具",
    InputSchema:  `{"type":"object","properties":{"text":{"type":"string"}}}`,
    Aspects:      []string{"performance", "accuracy"},
})

fmt.Printf("优化评分: %d\n", optimized.Score)
fmt.Printf("变更: %v\n", optimized.Changes)
```

#### 验证技能定义

```go
err := client.ValidateSkill(ctx, "my-skill-name")
```

### 4.10 统一工具列表

> 需要 scope: `skills`

统一工具接口将**技能 (Skill)** 和**插件 (Plugin)** 合并为一个视图。

#### 列出全部工具

```go
tools, err := client.ListTools(ctx)
for _, t := range tools {
    provider := "未知"
    sourceType := ""
    if t.Provider != nil {
        provider = t.Provider.Name
        sourceType = t.Provider.SourceType
        // SourceType: NATIVE | PROMPT | MCP | WORKFLOW | HTTP | ENGINE
    }
    fmt.Printf("%s [%s] — 来源: %s (%s)\n", t.Name, t.Category, provider, sourceType)

    // MCP 工具会包含端点信息
    if t.Provider != nil && t.Provider.MCPEndpoint != "" {
        fmt.Printf("  MCP 端点: %s\n", t.Provider.MCPEndpoint)
    }
}
```

#### 获取单个工具详情

```go
tool, err := client.GetTool(ctx, "tool-id-xxx")
fmt.Printf("名称: %s\n输入 Schema: %s\n输出 Schema: %s\n超时: %ds\n",
    tool.Name, tool.InputSchema, tool.OutputSchema, tool.Timeout)
```

### 4.11 WebSocket 实时推送

SDK 提供 WebSocket 长连接，用于接收服务端实时推送（余额变化、技能更新、系统通知等）。

#### 建立连接

```go
err := client.Connect(ctx, acosmi.WSConfig{
    // 事件回调 (必须)
    OnEvent: func(e acosmi.WSEvent) {
        fmt.Printf("[事件] type=%s topic=%s\n", e.Type, e.Topic)
        fmt.Printf("  数据: %s\n", string(e.Data))
    },

    // 连接建立回调
    OnConnect: func() {
        fmt.Println("WebSocket 已连接")
    },

    // 断线回调
    OnDisconnect: func(err error) {
        fmt.Printf("WebSocket 断开: %v\n", err)
    },

    // 自动订阅的主题
    Topics: []string{"balance", "skills", "system"},

    // 重连参数 (可选)
    ReconnectMin:  2 * time.Second,  // 最小重连间隔 (默认 2s)
    ReconnectMax:  60 * time.Second, // 最大重连间隔 (默认 60s)
    AutoReconnect: nil,              // 默认 true
})
if err != nil {
    log.Printf("WebSocket 连接失败: %v\n", err)
}
```

#### 检查连接状态

```go
if client.IsConnected() {
    fmt.Println("WebSocket 在线")
}
```

#### 断开连接

```go
client.Disconnect()
```

**WebSocket 特性**:
- 自动断线重连（指数退避: 2s → 4s → 8s → ... → 60s）
- 重连时自动重新订阅主题
- 30 秒握手超时
- 通过 context 取消控制生命周期

#### system 主题通知类型 (13 种)

通过 `"system"` 主题接收的通知 `WSEvent.Data` 包含以下类型:

| 类型常量 | 标题 | 触发场景 |
|---------|------|---------|
| `task_done` | 任务执行完成 | 异步任务/工作流完成 |
| `task_confirm` | 任务计划待确认 | 异步任务需用户确认 |
| `invite_success` | 邀请好友成功 | 邀请的好友注册成功 |
| `commission` | 佣金收入到账 | 代理佣金结算入账 |
| `register` | 欢迎加入 Acosmi | 新用户注册欢迎 |
| `entitlement` | 权益到账通知 | 管理员手动发放权益 |
| `entitlement_exp` | 权益即将到期 | 权益 7 天内到期 (每日 09:00 扫描) |
| `purchase` | 购买成功 | 商城订单支付完成 |
| `tk_alert` | TK 余额不足提醒 | 余额低于 100 TK (每日 10:00 扫描) |
| `withdraw` | 提现成功 | 提现审批通过并打款 |
| `reg_bonus` | 注册奖励到账 | 注册赠送权益发放完成 |
| `claim_monthly` | 免费权益领取成功 | 月度免费 Token 领取成功 |
| `unclaimed_reminder` | 您有免费权益待领取 | 当月未领取免费权益 (每日 11:00 扫描) |

---

## 5. CrabClaw-Skill CLI 命令手册

### 5.1 全局参数

```
crabclaw-skill [全局参数] <命令> [命令参数]
```

| 参数 | 缩写 | 默认值 | 说明 |
|------|------|--------|------|
| `--server` | - | 配置文件中的地址 | 服务器地址 |
| `--json` | - | false | JSON 格式输出（便于脚本处理） |
| `--help` | `-h` | - | 显示帮助 |

### 5.2 login — 登录

```bash
crabclaw-skill login [--force]
```

| 参数 | 缩写 | 说明 |
|------|------|------|
| `--force` | `-f` | 强制重新登录（即使已有有效 token） |

打开浏览器进行 OAuth 授权。授权成功后 token 保存到 `~/.acosmi/tokens.json`。

**示例**:
```bash
# 首次登录
crabclaw-skill login

# 强制刷新 token
crabclaw-skill login --force
```

### 5.3 logout — 登出

```bash
crabclaw-skill logout
```

吊销服务端 token 并清除本地存储的凭证。

### 5.4 whoami — 查看登录状态

```bash
crabclaw-skill whoami
```

显示当前登录状态、token 过期时间、授权范围等信息。

**示例输出**:
```
已登录
  服务器: https://acosmi.ai/api/v4
  授权范围: ai skills account
  过期时间: 2026-04-04 12:30:00
  状态: 有效
```

### 5.5 version — 版本信息

```bash
crabclaw-skill version
```

显示版本号和构建时间。

### 5.6 config — 配置管理

配置文件路径: `~/.acosmi/cli-config.json`

#### 查看配置

```bash
crabclaw-skill config show
```

#### 修改配置

```bash
crabclaw-skill config set <键> <值>
```

可配置项:

| 键 | 默认值 | 说明 |
|----|--------|------|
| `server` | `https://acosmi.ai` | API 服务地址 (生产环境) |
| `skilldir` | `~/.acosmi/skills` | 技能本地安装目录 |

**示例**:
```bash
crabclaw-skill config set server https://acosmi.ai
crabclaw-skill config set skilldir ~/my-skills
```

#### 重置配置

```bash
crabclaw-skill config reset
```

**环境变量覆盖**: `ACOSMI_SERVER_URL` 优先级高于配置文件。

### 5.7 search — 搜索技能

```bash
crabclaw-skill search <关键词> [参数]
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--category` | 全部 | 技能类别: ACTION / PERSONA / WORKFLOW 等 |
| `--tag` | 全部 | 标签过滤 |
| `--source` | 全部 | 来源过滤 |
| `--page` | 1 | 页码 |
| `--page-size` | 20 | 每页数量 |

**无需登录**。

**示例**:
```bash
# 搜索翻译相关技能
crabclaw-skill search "翻译"

# 按类别搜索
crabclaw-skill search "" --category ACTION --page-size 10

# JSON 格式输出
crabclaw-skill --json search "translate"
```

### 5.8 list — 已安装技能

```bash
crabclaw-skill list
```

列出当前用户已安装的技能。**需要登录**。

### 5.9 info — 技能详情

```bash
crabclaw-skill info <技能key>
```

查看技能的完整信息（描述、版本、Schema、README 等）。**无需登录**。

**示例**:
```bash
crabclaw-skill info translate-api

# JSON 格式
crabclaw-skill --json info translate-api
```

### 5.10 download — 下载技能

```bash
crabclaw-skill download <技能key> [--output <路径>]
```

| 参数 | 缩写 | 默认值 | 说明 |
|------|------|--------|------|
| `--output` | `-o` | `skill-{key}-v{version}.zip` | 输出文件路径 |

**无需登录**，但匿名下载有限流（2 次/小时）。登录后无限制。

**示例**:
```bash
crabclaw-skill download translate-api
crabclaw-skill download translate-api -o ~/skills/translate.zip
```

### 5.11 install — 安装技能

```bash
crabclaw-skill install <技能key> [参数]
```

| 参数 | 缩写 | 默认值 | 说明 |
|------|------|--------|------|
| `--local-only` | - | false | 仅安装到本地，不同步到服务器 |
| `--dir` | - | 配置的 skilldir | 本地安装目录 |
| `--force` | `-f` | false | 覆盖已安装的技能 |

如果指定了 `--local-only`，则无需登录；否则需要登录以将安装记录同步到服务端。

**示例**:
```bash
# 安装到服务器 + 本地
crabclaw-skill install translate-api

# 仅本地安装
crabclaw-skill install translate-api --local-only --dir ~/my-skills

# 强制覆盖
crabclaw-skill install translate-api --force
```

### 5.12 upload — 上传技能

```bash
crabclaw-skill upload <ZIP路径> [参数]
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--public` | false | 发布到公共商店 |
| `--certify` | false | 上传后自动触发认证 |

**需要登录**。

**示例**:
```bash
# 上传私有技能
crabclaw-skill upload ./my-skill.zip

# 上传公开技能并自动认证
crabclaw-skill upload ./my-skill.zip --public --certify
```

### 5.13 generate — AI 生成技能

```bash
crabclaw-skill generate "<描述>" [参数]
```

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--category` | 自动 | 技能类别 |
| `--language` | 自动 | 语言 |
| `--save` | 不保存 | 保存生成的 ZIP 到指定路径 |

**需要登录**。使用 AI 根据自然语言描述自动生成技能定义。

**示例**:
```bash
# 生成翻译技能
crabclaw-skill generate "一个将中文翻译成英文的工具"

# 生成并保存
crabclaw-skill generate "网页截图工具" --category ACTION --save screenshot.zip

# JSON 格式查看结果
crabclaw-skill --json generate "天气查询工具"
```

### 5.14 certify — 触发认证

```bash
crabclaw-skill certify <技能key或ID>
```

**需要登录**。触发异步认证流水线，通过 `info` 命令查看认证结果。

**示例**:
```bash
crabclaw-skill certify my-custom-skill
# 稍后查看结果
crabclaw-skill info my-custom-skill
```

---

## 6. 数据类型参考

### 6.1 OAuth 类型

#### ServerMetadata

OAuth 授权服务器元数据 (RFC 8414)。

```go
type ServerMetadata struct {
    Issuer                string   `json:"issuer"`
    AuthorizationEndpoint string   `json:"authorization_endpoint"`
    TokenEndpoint         string   `json:"token_endpoint"`
    RevocationEndpoint    string   `json:"revocation_endpoint"`
    RegistrationEndpoint  string   `json:"registration_endpoint"`
    ScopesSupported       []string `json:"scopes_supported"`
}
```

#### TokenSet

持久化的 token 对。

```go
type TokenSet struct {
    AccessToken  string    `json:"access_token"`
    RefreshToken string    `json:"refresh_token"`
    ExpiresAt    time.Time `json:"expires_at"`
    Scope        string    `json:"scope"`
    ClientID     string    `json:"client_id"`
    ServerURL    string    `json:"server_url"`
}

// IsExpired 在过期前 30 秒即返回 true
func (t *TokenSet) IsExpired() bool
```

### 6.2 模型类型

#### ManagedModel

```go
type ManagedModel struct {
    ID            string            `json:"id"`
    Name          string            `json:"name"`
    Provider      string            `json:"provider"`        // 如 "openai", "anthropic"
    ModelID       string            `json:"modelId"`          // 如 "gpt-4o", "claude-opus-4-6"
    MaxTokens     int               `json:"maxTokens"`
    IsEnabled     bool              `json:"isEnabled"`
    PricePerMTok  float64           `json:"pricePerMTok"`     // 每百万 Token 价格
    IsDefault     bool              `json:"isDefault"`
    ContextWindow int               `json:"contextWindow"`
    Capabilities  ModelCapabilities `json:"capabilities"`     // v0.2.0: 模型能力矩阵
}
```

#### ModelCapabilities (v0.2.0)

```go
type ModelCapabilities struct {
    SupportsThinking         bool `json:"supports_thinking"`
    SupportsAdaptiveThinking bool `json:"supports_adaptive_thinking"`
    SupportsISP              bool `json:"supports_isp"`              // 交错思考
    SupportsWebSearch        bool `json:"supports_web_search"`
    SupportsToolSearch       bool `json:"supports_tool_search"`
    SupportsStructuredOutput bool `json:"supports_structured_output"`
    SupportsEffort           bool `json:"supports_effort"`
    SupportsMaxEffort        bool `json:"supports_max_effort"`       // Opus 4.6 独有
    SupportsFastMode         bool `json:"supports_fast_mode"`        // Opus 4.6 独有
    Supports1MContext        bool `json:"supports_1m_context"`
    SupportsPromptCache      bool `json:"supports_prompt_cache"`
    SupportsCacheEditing     bool `json:"supports_cache_editing"`
    SupportsTokenEfficient   bool `json:"supports_token_efficient"`
    SupportsRedactThinking   bool `json:"supports_redact_thinking"`
    MaxInputTokens           int  `json:"max_input_tokens"`
    MaxOutputTokens          int  `json:"max_output_tokens"`
}
```

#### ChatMessage / ChatRequest / ChatResponse

```go
type ChatMessage struct {
    Role    string `json:"role"`    // "system" | "user" | "assistant"
    Content string `json:"content"`
}

type ChatRequest struct {
    // ── 基础字段 (v0.1.0, CrabClaw 兼容) ──
    Messages  []ChatMessage `json:"messages"`
    Stream    bool          `json:"stream,omitempty"`
    MaxTokens int           `json:"max_tokens,omitempty"`

    // ── 扩展字段 (v0.2.0, CrabCode, 全部 json:"-") ──
    RawMessages  interface{}            // 复杂消息体, 非 nil 时优先于 Messages
    System       interface{}            // 系统提示 (string 或 content blocks)
    Tools        interface{}            // 标准工具定义
    Temperature  *float64
    Thinking     *ThinkingConfig
    Metadata     map[string]string
    Betas        []string               // 显式 beta (SDK 自动合并去重)
    ServerTools  []ServerTool           // 服务端工具 (合入 tools 数组)
    Speed        string                 // "" | "fast"
    Effort       *EffortConfig
    OutputConfig *OutputConfig
    ExtraBody    map[string]interface{} // 透传任意字段
}

type ChatResponse struct {
    ID      string
    Choices []struct {
        Index   int
        Message ChatMessage
    }
    Usage struct {
        PromptTokens     int
        CompletionTokens int
        TotalTokens      int
    }
    // 结算后余额 (从响应 Header 填充, json:"-" 不参与 JSON 序列化)
    // -1 表示服务端未返回
    TokenRemaining int64
    CallRemaining  int
}
```

#### StreamEvent

SSE 流式事件。Server Tool 执行结果 (`web_search_tool_result` / `server_tool_use` 等) 以 JSON 内容透传在 `Data` 中，下游自行解析。

```go
type StreamEvent struct {
    Event string `json:"event"` // "started" | "settled" | "pending_settle" | "failed" | "" (数据块)
    Data  string `json:"data"`  // JSON 字符串
}
```

#### StreamSettlement

流式结算事件，包含本次请求的 token 消耗及结算后的剩余余额。通过 `ParseSettlement()` 从 `StreamEvent` 解析:

```go
type StreamSettlement struct {
    RequestID      string `json:"requestId"`
    ConsumeStatus  string `json:"consumeStatus"`  // "SETTLED" | "PENDING_SETTLE"
    InputTokens    int    `json:"inputTokens"`
    OutputTokens   int    `json:"outputTokens"`
    TotalTokens    int    `json:"totalTokens"`
    TokenRemaining int64  `json:"tokenRemaining"`  // -1 表示服务端未返回
    CallRemaining  int    `json:"callRemaining"`   // -1 表示服务端未返回
}

// ParseSettlement 从 settled/pending_settle 类型的 StreamEvent 中解析结算信息
// 非结算事件返回 nil
func ParseSettlement(ev StreamEvent) *StreamSettlement
```

**便捷方法**: `ChatStreamWithUsage()` 内部已自动调用 `ParseSettlement`，推荐直接使用高级 API。

### 6.2.1 Chat 扩展类型

> v0.2.0 新增

```go
// ThinkingConfig 思考配置
type ThinkingConfig struct {
    Type         string // "adaptive" | "enabled" | "disabled"
    BudgetTokens int    // type="enabled" 时的 token 预算
    Display      string // "none" | "summary" | "" (完整)
}

// ServerTool 服务端工具
type ServerTool struct {
    Type   string                 // 如 "web_search_20250305"
    Name   string                 // 如 "web_search"
    Config map[string]interface{} // 工具特定配置
}

// WebSearchConfig 搜索工具配置
type WebSearchConfig struct {
    MaxUses        int      // 每请求最大搜索次数 (默认 8)
    AllowedDomains []string // 域名白名单 (与 BlockedDomains 互斥)
    BlockedDomains []string // 域名黑名单
    UserLocation   *GeoLoc  // 搜索地域偏好
}

// GeoLoc 地理位置
type GeoLoc struct {
    Country string // ISO 3166-1 alpha-2
    City    string
}

// EffortConfig 推理努力级别
type EffortConfig struct {
    Level string // "low" | "medium" | "high" | "max"
}

// OutputConfig 结构化输出配置
type OutputConfig struct {
    Format string      // "json_schema" | ""
    Schema interface{} // JSON Schema 定义
}
```

**预定义常量**:

```go
const ServerToolTypeWebSearch = "web_search_20250305"
```

**便捷方法**:

```go
// NewWebSearchTool 创建搜索工具 (校验 AllowedDomains ⊕ BlockedDomains 互斥)
func NewWebSearchTool(cfg *WebSearchConfig) (ServerTool, error)

// GetModelCapabilities 查询单个模型能力 (复用 ListModels 缓存, 5min TTL)
func (c *Client) GetModelCapabilities(ctx context.Context, modelID string) (*ModelCapabilities, error)
```

### 6.3 权益类型

#### EntitlementBalance

```go
type EntitlementBalance struct {
    TotalTokenQuota     int64 `json:"totalTokenQuota"`
    TotalTokenUsed      int64 `json:"totalTokenUsed"`
    TotalTokenRemaining int64 `json:"totalTokenRemaining"`
    TotalCallQuota      int   `json:"totalCallQuota"`
    TotalCallUsed       int   `json:"totalCallUsed"`
    TotalCallRemaining  int   `json:"totalCallRemaining"`
    ActiveEntitlements  int   `json:"activeEntitlements"`
}
```

#### EntitlementItem

```go
type EntitlementItem struct {
    ID             string  `json:"id"`
    Type           string  `json:"type"`           // REG_BONUS | FREE_TRIAL | TOKEN_PKG | MONTHLY
    Status         string  `json:"status"`         // active | expired | exhausted
    TokenQuota     int64   `json:"tokenQuota"`
    TokenUsed      int64   `json:"tokenUsed"`
    TokenRemaining int64   `json:"tokenRemaining"`
    CallQuota      int     `json:"callQuota"`
    CallUsed       int     `json:"callUsed"`
    CallRemaining  int     `json:"callRemaining"`
    ExpiresAt      *string `json:"expiresAt,omitempty"`
    SourceID       string  `json:"sourceId,omitempty"`
    SourceType     string  `json:"sourceType,omitempty"`
    Remark         string  `json:"remark,omitempty"`
    CreatedAt      string  `json:"createdAt"`
}
```

#### ConsumeRecord / ConsumeRecordPage

```go
type ConsumeRecord struct {
    ID              string `json:"id"`
    EntitlementID   string `json:"entitlementId"`
    RequestID       string `json:"requestId"`
    ModelID         string `json:"modelId,omitempty"`
    TokensConsumed  int64  `json:"tokensConsumed"`
    Status          string `json:"status"`
    CreatedAt       string `json:"createdAt"`
}

type ConsumeRecordPage struct {
    Records  []ConsumeRecord `json:"records"`
    Total    int64           `json:"total"`
    Page     int             `json:"page"`
    PageSize int             `json:"pageSize"`
}
```

### 6.4 商城/订单类型

#### TokenPackage

```go
type TokenPackage struct {
    ID          string      `json:"id"`
    Name        string      `json:"name"`
    Description string      `json:"description,omitempty"`
    TokenQuota  int64       `json:"tokenQuota"`
    CallQuota   int         `json:"callQuota,omitempty"`
    Price       json.Number `json:"price"`       // 使用 json.Number 防止精度丢失
    ValidDays   int         `json:"validDays"`
    IsEnabled   bool        `json:"isEnabled"`
    SortOrder   int         `json:"sortOrder,omitempty"`
}
```

#### Order / OrderStatus

```go
type Order struct {
    ID          string      `json:"id"`
    PackageID   string      `json:"packageId"`
    PackageName string      `json:"packageName,omitempty"`
    Amount      json.Number `json:"amount"`
    Status      string      `json:"status"`     // pending | paid | expired | cancelled
    PayURL      string      `json:"payUrl,omitempty"`
    CreatedAt   string      `json:"createdAt"`
}

type OrderStatus struct {
    OrderID string `json:"orderId"`
    Status  string `json:"status"`
}
```

### 6.5 钱包类型

```go
type WalletStats struct {
    Balance            json.Number `json:"balance"`
    MonthlyConsumption json.Number `json:"monthlyConsumption"`
    MonthlyRecharge    json.Number `json:"monthlyRecharge"`
    TransactionCount   int         `json:"transactionCount"`
}

type Transaction struct {
    ID        string      `json:"id"`
    Type      string      `json:"type"`      // recharge | consume | refund | commission
    Amount    json.Number `json:"amount"`
    Remark    string      `json:"remark,omitempty"`
    CreatedAt string      `json:"createdAt"`
}
```

### 6.6 技能类型

#### SkillStoreItem (完整)

```go
type SkillStoreItem struct {
    ID                  string   `json:"id"`
    PluginID            string   `json:"pluginId"`
    Key                 string   `json:"key"`
    Name                string   `json:"name"`
    Description         string   `json:"description"`
    Icon                string   `json:"icon"`
    Category            string   `json:"category"`
    InputSchema         string   `json:"inputSchema"`
    OutputSchema        string   `json:"outputSchema"`
    Timeout             int      `json:"timeout"`
    RetryCount          int      `json:"retryCount"`
    RetryDelay          int      `json:"retryDelay"`
    Version             string   `json:"version"`
    TotalCalls          int64    `json:"totalCalls"`
    AvgDurationMs       int64    `json:"avgDurationMs"`
    SuccessRate         float64  `json:"successRate"`
    IsEnabled           bool     `json:"isEnabled"`
    SecurityLevel       string   `json:"securityLevel"`
    SecurityScore       int      `json:"securityScore"`
    Scope               string   `json:"scope"`
    Status              string   `json:"status"`
    DownloadCount       int64    `json:"downloadCount"`
    Readme              string   `json:"readme"`
    Tags                []string `json:"tags"`
    Author              string   `json:"author"`
    PublisherID         string   `json:"publisherId"`
    IsPublished         bool     `json:"isPublished"`
    PluginName          string   `json:"pluginName"`
    PluginIcon          string   `json:"pluginIcon"`
    UpdatedAt           string   `json:"updatedAt"`
    Visibility          string   `json:"visibility,omitempty"`
    CertificationStatus string   `json:"certificationStatus,omitempty"`
    Source              string   `json:"source,omitempty"`
}
```

#### SkillStoreListItem (轻量)

```go
type SkillStoreListItem struct {
    ID                  string   `json:"id"`
    Key                 string   `json:"key"`
    Name                string   `json:"name"`
    Description         string   `json:"description"`
    Icon                string   `json:"icon"`
    Category            string   `json:"category"`
    Version             string   `json:"version"`
    Author              string   `json:"author"`
    DownloadCount       int64    `json:"downloadCount"`
    Tags                []string `json:"tags"`
    CertificationStatus string   `json:"certificationStatus,omitempty"`
    Visibility          string   `json:"visibility,omitempty"`
    Source              string   `json:"source,omitempty"`
    UpdatedAt           string   `json:"updatedAt"`
}
```

#### 技能生成/优化

```go
type GenerateSkillRequest struct {
    Purpose     string   `json:"purpose"`              // 技能目的描述
    Examples    []string `json:"examples,omitempty"`    // 输入输出示例
    InputHints  string   `json:"inputHints,omitempty"`  // 输入提示
    OutputHints string   `json:"outputHints,omitempty"` // 输出提示
    Category    string   `json:"category,omitempty"`    // 类别
    Language    string   `json:"language,omitempty"`    // 语言
}

type GenerateSkillResult struct {
    SkillName    string   `json:"skillName"`
    SkillKey     string   `json:"skillKey"`
    Description  string   `json:"description"`
    SkillMd      string   `json:"skillMd"`
    InputSchema  string   `json:"inputSchema"`
    OutputSchema string   `json:"outputSchema"`
    TestCases    []string `json:"testCases"`
    Readme       string   `json:"readme"`
    Category     string   `json:"category"`
    Tags         []string `json:"tags"`
    Timeout      int      `json:"timeout"`
}

type OptimizeSkillRequest struct {
    SkillName    string   `json:"skillName"`
    Description  string   `json:"description,omitempty"`
    InputSchema  string   `json:"inputSchema,omitempty"`
    OutputSchema string   `json:"outputSchema,omitempty"`
    Readme       string   `json:"readme,omitempty"`
    Aspects      []string `json:"aspects,omitempty"`    // 优化方向
}

type OptimizeSkillResult struct {
    OptimizedSkill GenerateSkillResult `json:"optimizedSkill"`
    Changes        []string            `json:"changes"`
    Score          int                 `json:"score"`
}

type CertificationStatus struct {
    SkillID             string      `json:"skillId"`
    CertificationStatus string      `json:"certificationStatus"` // pending | in_progress | certified | rejected
    CertifiedAt         *int64      `json:"certifiedAt,omitempty"`
    SecurityLevel       string      `json:"securityLevel,omitempty"`
    SecurityScore       int         `json:"securityScore"`
    Report              interface{} `json:"report,omitempty"`
}
```

### 6.7 工具类型

```go
type ToolView struct {
    ID           string        `json:"id"`
    Key          string        `json:"key"`
    Name         string        `json:"name"`
    Description  string        `json:"description"`
    Icon         string        `json:"icon"`
    Category     string        `json:"category"`
    InputSchema  string        `json:"inputSchema"`
    OutputSchema string        `json:"outputSchema"`
    Timeout      int           `json:"timeout"`
    IsEnabled    bool          `json:"isEnabled"`
    Provider     *ToolProvider `json:"provider,omitempty"`
}

type ToolProvider struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Icon        string `json:"icon"`
    SourceType  string `json:"sourceType"`         // NATIVE | PROMPT | MCP | WORKFLOW | HTTP | ENGINE
    MCPEndpoint string `json:"mcpEndpoint,omitempty"` // MCP 协议端点
    IsEnabled   bool   `json:"isEnabled"`
}
```

### 6.8 WebSocket 类型

```go
type WSConfig struct {
    OnEvent       func(WSEvent)     // 事件回调（必须）
    OnConnect     func()            // 连接建立回调
    OnDisconnect  func(error)       // 断线回调
    Topics        []string          // 自动订阅主题
    ReconnectMin  time.Duration     // 最小重连间隔（默认 2s）
    ReconnectMax  time.Duration     // 最大重连间隔（默认 60s）
    AutoReconnect *bool             // 自动重连（默认 true）
}

type WSEvent struct {
    Type      string          `json:"type"`
    Topic     string          `json:"topic,omitempty"`
    Data      json.RawMessage `json:"data,omitempty"`
    ConnID    string          `json:"connId,omitempty"`
    Timestamp string          `json:"timestamp,omitempty"`
    Message   string          `json:"message,omitempty"`
}
```

### 6.9 错误类型

#### RateLimitError

下载技能时触发 429 限流返回此错误。

```go
type RateLimitError struct {
    Message    string // 错误消息
    RetryAfter string // 建议重试时间
    Raw        string // 原始响应体
}

// 使用 errors.As 判断
var rateErr *acosmi.RateLimitError
if errors.As(err, &rateErr) {
    fmt.Printf("请 %s 后重试\n", rateErr.RetryAfter)
}
```

### 6.10 通用响应包装

```go
// 标准 API 响应（兼容 yudao 和 nexus-v4 格式）
type APIResponse[T any] struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
    Msg     string `json:"msg"`      // yudao 兼容
    Data    T      `json:"data"`
}

// 获取消息（优先 message，降级到 msg）
func (r *APIResponse[T]) GetMessage() string

// yudao 分页格式
type YudaoPageResult[T any] struct {
    List  []T   `json:"list"`
    Total int64 `json:"total"`
}
```

---

## 7. 完整示例

以下是一个完整的 SDK 使用示例，展示了从登录到调用各种 API 的全流程:

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "time"

    acosmi "github.com/acosmi/acosmi-sdk-go"
)

func main() {
    // 服务地址（支持环境变量覆盖）
    // 默认连接生产环境 https://acosmi.ai; 本地开发设置环境变量覆盖:
    //   export ACOSMI_SERVER_URL=http://127.0.0.1:3300
    serverURL := os.Getenv("ACOSMI_SERVER_URL")
    if serverURL == "" {
        serverURL = "https://acosmi.ai"
    }

    // 1. 创建客户端
    client, err := acosmi.NewClient(acosmi.Config{
        ServerURL: serverURL,
    })
    if err != nil {
        log.Fatal(err)
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
    defer cancel()

    // 2. 登录
    if !client.IsAuthorized() {
        fmt.Println("首次使用，将打开浏览器进行授权...")
        if err := client.Login(ctx, "CrabClaw-Skill Desktop Agent", acosmi.AllScopes()); err != nil {
            log.Fatalf("授权失败: %v", err)
        }
        fmt.Println("授权成功!")
    }

    // 3. 查询权益余额
    balance, err := client.GetBalance(ctx)
    if err != nil {
        log.Fatalf("查询余额失败: %v", err)
    }
    fmt.Printf("Token 余额: %d / %d (剩余 %d)\n",
        balance.TotalTokenUsed, balance.TotalTokenQuota, balance.TotalTokenRemaining)

    // 4. 查询钱包
    if wallet, err := client.GetWalletStats(ctx); err == nil {
        fmt.Printf("钱包余额: %s 元\n", wallet.Balance)
    }

    // 5. 浏览流量包
    if packages, err := client.ListTokenPackages(ctx); err == nil {
        fmt.Printf("\n流量包商城 (%d 个):\n", len(packages))
        for _, p := range packages {
            fmt.Printf("  - %s: %d Token / %s 元\n", p.Name, p.TokenQuota, p.Price)
        }
    }

    // 6. 浏览技能商店
    if skills, err := client.BrowseSkillStore(ctx, acosmi.SkillStoreQuery{}); err == nil {
        fmt.Printf("\n技能商店 (%d 个):\n", len(skills))
        for _, s := range skills {
            fmt.Printf("  - %s [%s] 下载: %d\n", s.Name, s.Category, s.DownloadCount)
        }
    }

    // 7. 获取模型列表
    models, err := client.ListModels(ctx)
    if err != nil {
        log.Fatalf("获取模型列表失败: %v", err)
    }
    fmt.Printf("\n可用模型 (%d 个):\n", len(models))
    for _, m := range models {
        fmt.Printf("  - %s (%s)\n", m.Name, m.ModelID)
    }

    if len(models) == 0 {
        fmt.Println("没有可用模型")
        return
    }

    // 8. WebSocket 实时推送
    err = client.Connect(ctx, acosmi.WSConfig{
        Topics: []string{"balance", "skills"},
        OnEvent: func(e acosmi.WSEvent) {
            fmt.Printf("[WS] %s: %s\n", e.Type, string(e.Data))
        },
        OnConnect:    func() { fmt.Println("[WS] 已连接") },
        OnDisconnect: func(err error) { fmt.Printf("[WS] 断开: %v\n", err) },
    })
    if err == nil {
        defer client.Disconnect()
    }

    // 9. 流式聊天
    fmt.Printf("\n使用 %s 进行对话:\n", models[0].Name)
    eventCh, errCh := client.ChatStream(ctx, models[0].ID, acosmi.ChatRequest{
        Messages: []acosmi.ChatMessage{
            {Role: "user", Content: "用一句话介绍你自己"},
        },
        MaxTokens: 256,
    })

    fmt.Print("AI: ")
    for event := range eventCh {
        if event.Event == "settled" || event.Event == "started" {
            continue
        }
        var chunk struct {
            Choices []struct {
                Delta struct {
                    Content string `json:"content"`
                } `json:"delta"`
            } `json:"choices"`
        }
        if json.Unmarshal([]byte(event.Data), &chunk) == nil && len(chunk.Choices) > 0 {
            fmt.Print(chunk.Choices[0].Delta.Content)
        }
    }
    fmt.Println()

    if err := <-errCh; err != nil {
        log.Fatalf("流式聊天错误: %v", err)
    }
}
```

---

## 8. 安全特性

SDK 内置了多项安全防护措施，无需额外配置:

| 编号 | 风险 | 防护措施 |
|------|------|----------|
| RC-1 | appName JSON 注入 | 使用 `json.Marshal` 安全编码 |
| RC-2 | Token 无限刷新循环 | 强制最低 60 秒有效期 |
| RC-3 | 全局超时截断 SSE 流 | 不设全局 Timeout，使用 per-request Context |
| RC-4 | 401 无限递归 | 单次重试限制 |
| RC-5 | 元数据并发竞争 | `sync.RWMutex` 保护 |
| RC-6 | ZIP 边界碰撞 | 随机边界生成 |
| RC-7 | Token 刷新竞争 | 单次重试限制 |
| RC-8 | 错误消息 XSS | HTML 转义 |
| RC-9 | Scope 数组外部篡改 | 函数返回新切片 |
| RC-10 | Home 目录为空 | 提前返回错误 |
| RC-11 | 域名后缀匹配 | 使用 `HasSuffix` 替代 `Contains` |
| RC-12 | OrderPage 死代码 | 移除未使用类型，防止 API 混淆 |
| RC-13 | SSE 超大行 | 1MB 缓冲区限制 |
| RC-14 | ExpiresIn=0 | 强制 60 秒最低值 |
| RC-15 | WebSocket 握手阻塞 | 30 秒显式超时 |

**ZIP 安全防护** (CLI 解压时):
- Zip Slip 路径穿越检查
- 去除 setuid/setgid/sticky 位
- 下载体积限制: 50MB

**Token 文件安全**:
- 目录权限: `0700`
- 文件权限: `0600`

---

## 9. 项目结构

```
acosmi-sdk-go/
├── auth.go              # OAuth 2.1 PKCE 认证流程
├── client.go            # 统一 API 客户端 + buildChatRequest
├── types.go             # 所有数据类型定义 (含 v0.2.0 扩展类型)
├── betas.go             # Beta Header 自动组装引擎 (v0.2.0)
├── store.go             # Token 持久化接口 + 文件存储实现
├── scopes.go            # OAuth Scope 常量和分组函数
├── ws.go                # WebSocket 长连接（自动重连）
├── go.mod               # Go 模块定义
├── go.sum               # 依赖校验
├── Makefile             # 构建脚本
├── .gitignore           # Git 忽略规则
├── README.md            # 项目简介
├── docs/
│   └── guide.md         # 本开发手册
├── example/
│   └── main.go          # 完整使用示例
├── cmd/
│   └── crabclawskill/   # CrabClaw-Skill CLI
│       ├── main.go      # CLI 入口 + 根命令
│       ├── config.go    # 配置管理
│       ├── zip.go       # ZIP 安全解压/打包
│       ├── cmd_login.go
│       ├── cmd_logout.go
│       ├── cmd_whoami.go
│       ├── cmd_version.go
│       ├── cmd_config.go
│       ├── cmd_search.go
│       ├── cmd_list.go
│       ├── cmd_info.go
│       ├── cmd_download.go
│       ├── cmd_install.go
│       ├── cmd_upload.go
│       ├── cmd_generate.go
│       └── cmd_certify.go
└── npm/                  # NPM 发布包装
    ├── package.json
    ├── bin/
    │   └── crabclaw-skill.js
    └── scripts/
        └── postinstall.js
```

---

## 10. 构建与发布

### 本地构建

```bash
# 构建当前平台
make build          # → bin/crabclaw-skill

# 交叉编译全平台
make build-all      # → dist/crabclaw-skill-{os}-{arch}

# 安装到 GOPATH
make install

# 清理构建产物
make clean
```

### 支持的平台

| 操作系统 | 架构 | 二进制名 |
|----------|------|----------|
| macOS | arm64 (Apple Silicon) | `crabclaw-skill-darwin-arm64` |
| macOS | amd64 (Intel) | `crabclaw-skill-darwin-amd64` |
| Linux | amd64 | `crabclaw-skill-linux-amd64` |
| Linux | arm64 | `crabclaw-skill-linux-arm64` |
| Windows | amd64 | `crabclaw-skill-windows-amd64.exe` |

### 版本信息注入

构建时通过 `-ldflags` 自动注入版本和构建时间:

```bash
# 使用 git tag 作为版本
git tag v0.1.0
make build
./bin/crabclaw-skill version
# crabclaw-skill v0.1.0 (built 2026-04-04T10:00:00Z)
```

### NPM 发布

```bash
cd npm
npm publish --access public
```

NPM 包在 `postinstall` 阶段自动从 GitHub Releases 下载对应平台的预编译二进制。

---

## 11. 常见问题

### Q: 如何切换服务器地址？

三种方式（优先级从高到低）:

1. **CLI 参数**: `crabclaw-skill --server https://acosmi.ai <命令>`
2. **环境变量**: `export ACOSMI_SERVER_URL=https://acosmi.ai`
3. **配置文件**: `crabclaw-skill config set server https://acosmi.ai`

### Q: Token 过期了怎么办？

SDK 会在 token 过期前 30 秒自动刷新，通常无需手动处理。如果 Refresh Token 也过期（7 天不活动），需要重新登录:

```bash
crabclaw-skill login --force
```

### Q: 匿名下载被限流了？

匿名用户每小时只能下载 2 次。登录后无限制:

```bash
crabclaw-skill login
crabclaw-skill download my-skill
```

### Q: 如何在 CI/CD 中使用？

CI 环境无法打开浏览器进行 OAuth 授权。建议:
1. 在本地登录获取 token
2. 将 `~/.acosmi/tokens.json` 作为 CI Secret
3. CI 中将 secret 写入同路径

### Q: WebSocket 一直重连怎么办？

检查:
1. 服务器地址是否正确
2. token 是否有效 (`crabclaw-skill whoami`)
3. 防火墙是否阻止了 WebSocket 连接

可设置 `AutoReconnect` 为 `false` 来禁用自动重连:

```go
autoReconnect := false
client.Connect(ctx, acosmi.WSConfig{
    AutoReconnect: &autoReconnect,
    // ...
})
```

### Q: 如何处理 API 错误？

所有 API 方法返回标准的 `error` 接口。HTTP 状态码信息包含在错误消息中:

```go
balance, err := client.GetBalance(ctx)
if err != nil {
    // err.Error() 包含 HTTP 状态码和服务端错误消息
    // 例: "HTTP 401: not authorized"
    fmt.Println(err)
}
```

对于限流错误，可使用 `errors.As` 进行类型判断:

```go
var rateErr *acosmi.RateLimitError
if errors.As(err, &rateErr) {
    fmt.Printf("限流中，请 %s 后重试\n", rateErr.RetryAfter)
}
```

### Q: SDK 线程安全吗？

是的。`Client` 内部使用 `sync.RWMutex` 保护所有共享状态（token、元数据），所有 API 调用都可以在多个 goroutine 中并发使用。

---

## 12. 版本修订记录

### v0.2.1 (2026-04-06) — 实时余额推送

**types.go 变更**:
- `ChatResponse` 新增 `TokenRemaining int64` / `CallRemaining int` (`json:"-"`, 从响应 Header 填充, -1 = 未返回)
- 新增 `StreamSettlement` 结构体: 7 字段 (requestId + consumeStatus + 3 token 消耗 + 2 余额)
- 新增 `ParseSettlement(ev StreamEvent) *StreamSettlement` — 从 settled/pending_settle 事件解析结算信息

**client.go 变更**:
- `Chat()` 改用 `doJSONFull` — 读取 `X-Token-Remaining` / `X-Call-Remaining` 响应 Header 填充余额字段
- 新增 `ChatStreamWithUsage(ctx, modelID, req)` — 返回 contentCh + settleCh + errCh 三 channel:
  - contentCh: 纯内容事件 (过滤 started/settled/failed)
  - settleCh: 结算信息 (token 消耗 + 剩余余额)
  - errCh: 传输错误 + 服务端 failed 事件
- 新增 `doJSONFull` / `doJSONFullInternal` — 返回 `(http.Header, error)`, 复用全部 doJSON 逻辑
- `doJSON` 重构为 `doJSONFullInternal` 的 wrapper (消除 60 行代码重复)
- 新增 `parseStreamError(data)` — 从 failed 事件 JSON 提取 `"stage: error"` 格式化错误

**审计修复**:
- P1: `ChatStreamWithUsage` 全部 channel send 用 `select + ctx.Done()` 防 goroutine 泄漏
- P1: `failed` 事件不再泄漏到 contentCh, 解析为 error 发送到 errCh

**后端变更** (nexus-v4 managed_model.go):
- 流式 settled 事件新增 `tokenRemaining` / `callRemaining` (settle 后查询余额, 失败静默降级)
- 同步响应新增 `X-Token-Remaining` / `X-Call-Remaining` Header

### v0.2.0 (2026-04-06) — CrabCode 增值能力扩展

**新增文件**:
- `betas.go` — Beta Header 自动组装引擎 (11 项真实 beta + 互斥规则)

**types.go 变更**:
- `ChatRequest` 新增 12 个扩展字段: `RawMessages`, `System`, `Tools`, `Temperature`, `Thinking`, `Metadata`, `Betas`, `ServerTools`, `Speed`, `Effort`, `OutputConfig`, `ExtraBody`
- 所有扩展字段标记 `json:"-"`，仅通过 `buildChatRequest` 序列化 (CrabClaw 零影响)
- 新增 `ThinkingConfig` 结构体: `Type` (adaptive/enabled/disabled) + `BudgetTokens` + `Display`
- 新增 `ServerTool` 结构体: `Type` + `Name` + `Config` (服务端工具, 如联网搜索)
- 新增 `ServerToolTypeWebSearch` 常量: `"web_search_20250305"` (Brave Search)
- 新增 `WebSearchConfig` 结构体: `MaxUses` + `AllowedDomains` + `BlockedDomains` + `UserLocation`
- 新增 `GeoLoc` 结构体: `Country` (ISO 3166-1) + `City`
- 新增 `EffortConfig` 结构体: `Level` (low/medium/high/max)
- 新增 `OutputConfig` 结构体: `Format` (json_schema) + `Schema`
- 新增 `NewWebSearchTool(cfg)` 便捷方法 (返回 `(ServerTool, error)`, 校验 AllowedDomains ⊕ BlockedDomains 互斥)
- `ManagedModel` 新增 `Capabilities ModelCapabilities` 字段

**新增 `ModelCapabilities` 结构体** (16 字段):
- 思考: `SupportsThinking`, `SupportsAdaptiveThinking`, `SupportsISP`
- 工具: `SupportsWebSearch`, `SupportsToolSearch`, `SupportsStructuredOutput`
- 推理: `SupportsEffort`, `SupportsMaxEffort`, `SupportsFastMode`
- 上下文: `Supports1MContext`, `SupportsPromptCache`, `SupportsCacheEditing`
- 输出: `SupportsTokenEfficient`, `SupportsRedactThinking`
- 数值: `MaxInputTokens`, `MaxOutputTokens`

**client.go 变更**:
- `Client` 新增 `modelCache []ManagedModel` + `modelCacheTime time.Time` (5min TTL)
- 新增 `buildChatRequest(modelID, req)` 内部方法: ServerTools 合入 tools 数组 + ExtraBody 透传 + Beta 自动组装
- 新增 `GetModelCapabilities(ctx, modelID)` 公开方法: 从 ListModels 缓存查询单个模型能力
- 新增 `getCachedCapabilities(modelID)` 内部方法: 线程安全缓存查询
- `ListModels()` 增加缓存写入逻辑
- `Chat()` 重构: 从 `json.Marshal(req)` 改为 `buildChatRequest` → `json.RawMessage`
- `chatStreamInternal()` 重构: 同上

**auth.go 变更**:
- 新增 `LoginWithHandler(ctx, appName, scopes, handler, opts...)` — 带事件回调的登录流程 (CrabCode 适配)
- 新增 `LoginEvent` / `LoginErrCode` / `LoginEventType` 事件类型
- 新增 `LoginOption` 函数式选项: `WithSkipBrowser()`, `WithLoginHint()`, `WithLoginMethod()`, `WithOrgUUID()`, `WithExpiresIn()`
- `Login()` 内部委托 `loginInternal()`，签名不变

**betas.go 新增** (11 项 beta 常量 + 逻辑):
- `buildBetas(caps, req)` 根据 ModelCapabilities + 请求参数自动组装
- 互斥规则: `structured-outputs` ⊕ `token-efficient-tools`
- `uniqueMerge(base, extra)` 去重合并
- 经联网验证剔除 3 项虚构 beta (task-budgets/advisor-tool/cache-editing)、修正 2 项日期 (structured-outputs → 2025-11-13, token-efficient-tools → 2025-02-19)

**审计修复**:
- P1: 删除 `hasServerTool` 死代码
- P1: `NewWebSearchTool` 添加 `AllowedDomains ⊕ BlockedDomains` 互斥校验 (返回 error)
- P2: 全部扩展字段统一 `json:"-"` (修正 System/Tools/Temperature/Thinking/Metadata 的 JSON tag 不一致)
- P2: `buildChatRequest` 文档补充 ListModels 前置要求说明

### v0.2.0 后端安全加固 (2026-04-06)

**Gateway 安全**:
- `sanitizeBetas()`: beta header 字符白名单 `[a-zA-Z0-9_-]` + 单值长度 ≤64 + 数量 ≤20, 防止 CRLF/header 注入
- `safePositiveInt()`: thinking_budget 类型校验 + 正数 + 上限 100000, 非法值静默降级
- `adaptPassthrough` Speed 仅接受 `"fast"`, Effort level 仅接受 `low/medium/high/max`, 重建 map 丢弃未知 key
- Effort 未知 level 不触发 `enable_thinking` (DashScope/VolcEngine switch 改为仅匹配已知值)

**API 响应安全**:
- 新增 `ManagedModelPublicResponse` — 非管理员 `GET /managed-models` 隐藏 providerConfig/endpoint/pricing/sortOrder/hasApiKey
- 管理员和 S2S 端点保持完整 `ManagedModelResponse`

**Chat 输入校验**:
- `betas` ≤20, `tools` ≤50, `messages` ≤500, 超限返回 400

### v0.1.0 (2026-03-22) — 初始发布

- 统一 acosmi-sdk-go (合并 desktop-sdk-go + jineng-sdk-go)
- OAuth 2.1 PKCE 授权流程
- 34 个公开 API 方法覆盖: 模型/权益/商城/钱包/技能/工具/WebSocket
- CrabClaw-Skill CLI (13 命令)
- 18 项根因修复 (3P0+5P1+8P2+6P3)
- 中文开发手册

---

## 许可证

MIT License

Copyright (c) 2026 Acosmi
