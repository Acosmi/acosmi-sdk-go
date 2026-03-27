package acosmi

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// 安全限制常量
const (
	maxDownloadSize   = 50 << 20 // 50MB — 技能 ZIP 包最大下载体积
	maxErrorBodySize  = 1 << 20  // 1MB — 错误响应体最大读取量
	maxSSELineSize    = 1 << 20  // 1MB — SSE 单行最大长度 (大 JSON chunk)
)

// Client Acosmi nexus-v4 统一 API 客户端
// 覆盖全域 API: 模型/权益/商城/钱包/技能/工具/WebSocket
// 自动处理 token 刷新，所有 API 调用线程安全
type Client struct {
	serverURL string
	meta      *ServerMetadata
	tokens    *TokenSet
	store     TokenStore
	http      *http.Client
	mu        sync.RWMutex
	ws        *wsState // WebSocket 长连接状态 (nil = 未连接)
}

// Config 客户端配置
type Config struct {
	// ServerURL nexus-v4 API 根地址，如 http://127.0.0.1:8009 或 http://127.0.0.1:3300/api/v4
	ServerURL string

	// Store token 持久化实现，nil 则使用默认文件存储 (~/.acosmi/tokens.json)
	Store TokenStore

	// HTTPClient 自定义 HTTP 客户端，nil 则使用默认
	HTTPClient *http.Client
}

// NewClient 创建客户端 (自动加载已保存的 token)
func NewClient(cfg Config) (*Client, error) {
	if cfg.ServerURL == "" {
		return nil, fmt.Errorf("ServerURL is required")
	}

	store := cfg.Store
	if store == nil {
		var storeErr error
		store, storeErr = NewFileTokenStore("")
		if storeErr != nil {
			return nil, fmt.Errorf("init token store: %w", storeErr)
		}
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		// [RC-3] 不设全局 Timeout — 全局 Timeout 含 body 读取,
		// 会截断 SSE 流式聊天和大文件下载。改为 per-request context timeout。
		httpClient = &http.Client{}
	}

	c := &Client{
		serverURL: strings.TrimRight(cfg.ServerURL, "/"),
		store:     store,
		http:      httpClient,
	}

	// 尝试加载已保存的 token
	if tokens, err := store.Load(); err == nil && tokens != nil {
		c.tokens = tokens
	}

	return c, nil
}

// ============================================================================
// 授权生命周期
// ============================================================================

// IsAuthorized 是否已授权 (有可用 token)
func (c *Client) IsAuthorized() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tokens != nil
}

// Login 完整授权流程: 发现 → 注册 → 授权 → 换 token → 持久化
// appName: 桌面智能体名称 (如 "CrabClaw Desktop")
// scopes: 请求的权限范围 (参考 AllScopes / ModelScopes / CommerceScopes 等预设)
func (c *Client) Login(ctx context.Context, appName string, scopes []string) error {
	// 1. 发现
	meta, err := Discover(ctx, c.serverURL)
	if err != nil {
		return fmt.Errorf("discovery failed: %w", err)
	}
	// [RC-5] 持锁写入 c.meta, 防止与 ensureToken/forceRefresh 读取产生数据竞争
	c.mu.Lock()
	c.meta = meta
	c.mu.Unlock()

	// 2. 检查是否已有 client_id
	c.mu.RLock()
	var clientID string
	if c.tokens != nil {
		clientID = c.tokens.ClientID
	}
	c.mu.RUnlock()

	if clientID == "" {
		reg, err := Register(ctx, meta, appName)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}
		clientID = reg.ClientID
	}

	// 3. 授权 (打开浏览器)
	result, verifier, err := Authorize(ctx, meta, clientID, scopes)
	if err != nil {
		return fmt.Errorf("authorization failed: %w", err)
	}

	// 4. 换 token
	tokenResp, err := ExchangeCode(ctx, meta, clientID, result.Code, result.RedirectURI, verifier)
	if err != nil {
		return fmt.Errorf("token exchange failed: %w", err)
	}

	// 5. 持久化
	tokens := NewTokenSet(tokenResp, clientID, c.serverURL)
	c.mu.Lock()
	c.tokens = tokens
	c.mu.Unlock()

	if err := c.store.Save(tokens); err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}

	return nil
}

// Logout 吊销 token 并清除本地存储
// [RC-4] meta==nil 时先 Discover 获取 revocation endpoint, 确保服务端 token 也被撤销
func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	tokens := c.tokens
	meta := c.meta
	c.tokens = nil
	c.meta = nil
	c.mu.Unlock()

	if tokens != nil {
		if meta == nil {
			// Lazy discover for revocation endpoint
			discovered, discErr := Discover(ctx, c.serverURL)
			if discErr != nil {
				fmt.Printf("[acosmi-sdk] warning: discover for revocation failed: %v\n", discErr)
			} else {
				meta = discovered
			}
		}
		if meta != nil {
			_ = RevokeToken(ctx, meta, tokens.AccessToken)
			_ = RevokeToken(ctx, meta, tokens.RefreshToken)
		}
	}

	return c.store.Clear()
}

// GetTokenSet 返回当前 token 信息 (用于 CLI whoami 显示)
func (c *Client) GetTokenSet() *TokenSet {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tokens
}

// ============================================================================
// Token 管理
// ============================================================================

// ensureToken 确保有有效的 access_token，过期则自动刷新
func (c *Client) ensureToken(ctx context.Context) (string, error) {
	c.mu.RLock()
	tokens := c.tokens
	c.mu.RUnlock()

	if tokens == nil {
		return "", fmt.Errorf("not authorized, call Login() first")
	}

	if !tokens.IsExpired() {
		return tokens.AccessToken, nil
	}

	// 需要刷新
	c.mu.Lock()
	defer c.mu.Unlock()

	// 根因修复 #2: 双检锁中 c.tokens 可能已被 Logout() 置 nil → panic
	// 必须先检查 nil, 再检查过期
	if c.tokens == nil {
		return "", fmt.Errorf("not authorized, call Login() first")
	}
	if !c.tokens.IsExpired() {
		return c.tokens.AccessToken, nil
	}

	if c.meta == nil {
		meta, err := Discover(ctx, c.serverURL)
		if err != nil {
			return "", fmt.Errorf("discover for refresh: %w", err)
		}
		c.meta = meta
	}

	tokenResp, err := RefreshToken(ctx, c.meta, c.tokens.ClientID, c.tokens.RefreshToken)
	if err != nil {
		return "", fmt.Errorf("refresh token: %w", err)
	}

	c.tokens = NewTokenSet(tokenResp, c.tokens.ClientID, c.serverURL)
	if err := c.store.Save(c.tokens); err != nil {
		fmt.Printf("[acosmi-sdk] warning: save refreshed token failed: %v\n", err)
	}

	return c.tokens.AccessToken, nil
}

// ============================================================================
// API: Managed Models (scope: models / models:chat)
// ============================================================================

// ListModels 获取可用的托管模型列表
func (c *Client) ListModels(ctx context.Context) ([]ManagedModel, error) {
	var resp APIResponse[[]ManagedModel]
	if err := c.doJSON(ctx, http.MethodGet, "/managed-models", nil, &resp, false); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// [RC-2] GetModelUsage 已移除: /managed-models/usage 端点已迁移至 tk-dist 营销系统

// Chat 同步聊天 (适合短回复)
func (c *Client) Chat(ctx context.Context, modelID string, req ChatRequest) (*ChatResponse, error) {
	req.Stream = false
	var resp ChatResponse
	if err := c.doJSON(ctx, http.MethodPost, "/managed-models/"+modelID+"/chat", req, &resp, false); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ChatStream 流式聊天 (SSE)，通过 channel 返回事件
// 调用方应遍历 channel 直到关闭，errCh 报告非 nil 错误
func (c *Client) ChatStream(ctx context.Context, modelID string, req ChatRequest) (<-chan StreamEvent, <-chan error) {
	eventCh := make(chan StreamEvent, 32)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)
		c.chatStreamInternal(ctx, modelID, req, eventCh, errCh, false)
	}()

	return eventCh, errCh
}

// chatStreamInternal 流式聊天内部实现
// 根因修复 #7: ChatStream 支持 401 单次重试
// 根因修复 #4: 错误响应体使用 LimitReader 防 OOM
// 根因修复 #13: SSE scanner 使用 1MB 缓冲区, 防 ErrTooLong
func (c *Client) chatStreamInternal(ctx context.Context, modelID string, req ChatRequest,
	eventCh chan<- StreamEvent, errCh chan<- error, retried bool) {

	req.Stream = true
	body, _ := json.Marshal(req)

	token, err := c.ensureToken(ctx)
	if err != nil {
		errCh <- err
		return
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.apiURL("/managed-models/"+modelID+"/chat"),
		bytes.NewReader(body))
	if err != nil {
		errCh <- err
		return
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		errCh <- err
		return
	}
	defer resp.Body.Close()

	// 401 单次重试 (根因修复 #7)
	if resp.StatusCode == http.StatusUnauthorized && !retried {
		resp.Body.Close()
		if refreshErr := c.forceRefresh(ctx); refreshErr != nil {
			errCh <- fmt.Errorf("stream: unauthorized and refresh failed: %w", refreshErr)
			return
		}
		c.chatStreamInternal(ctx, modelID, req, eventCh, errCh, true)
		return
	}

	if resp.StatusCode != http.StatusOK {
		// 根因修复 #4: 限制错误响应体大小
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		errCh <- fmt.Errorf("stream: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
		return
	}

	// 根因修复 #13: 扩大 SSE scanner 缓冲区到 1MB
	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 0, maxSSELineSize), maxSSELineSize)

	var currentEvent string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event:") {
			currentEvent = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
		} else if strings.HasPrefix(line, "data:") {
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				return
			}
			eventCh <- StreamEvent{Event: currentEvent, Data: data}
		}
	}

	if err := scanner.Err(); err != nil {
		errCh <- err
	}
}

// ============================================================================
// API: Entitlements (scope: entitlements)
// ============================================================================

// GetBalance 查询当前用户的权益余额 (聚合)
func (c *Client) GetBalance(ctx context.Context) (*EntitlementBalance, error) {
	var resp APIResponse[EntitlementBalance]
	if err := c.doJSON(ctx, http.MethodGet, "/entitlements/balance", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetBalanceDetail 查询详细余额 (含每条权益明细)
func (c *Client) GetBalanceDetail(ctx context.Context) (*BalanceDetail, error) {
	var resp APIResponse[BalanceDetail]
	if err := c.doJSON(ctx, http.MethodGet, "/entitlements/balance-detail", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ListEntitlements 查询当前用户权益列表
// status: "ACTIVE" / "EXPIRED" / "" (全部)
func (c *Client) ListEntitlements(ctx context.Context, status string) ([]EntitlementItem, error) {
	path := "/entitlements"
	if status != "" {
		path += "?status=" + url.QueryEscape(status)
	}
	var resp APIResponse[[]EntitlementItem]
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp, false); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ListConsumeRecords 查询核销记录 (分页)
func (c *Client) ListConsumeRecords(ctx context.Context, page, pageSize int) (*ConsumeRecordPage, error) {
	path := fmt.Sprintf("/entitlements/consume-records?page=%d&pageSize=%d", page, pageSize)
	var resp APIResponse[ConsumeRecordPage]
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ============================================================================
// API: Token Packages / 商城 (scope: token-packages)
// ============================================================================

// ListTokenPackages 获取商城流量包列表
// 兼容 yudao 分页格式和直接数组格式 (tk-dist 代理透传)
func (c *Client) ListTokenPackages(ctx context.Context) ([]TokenPackage, error) {
	var raw APIResponse[json.RawMessage]
	if err := c.doJSON(ctx, http.MethodGet, "/token-packages", nil, &raw, false); err != nil {
		return nil, err
	}
	var page YudaoPageResult[TokenPackage]
	if json.Unmarshal(raw.Data, &page) == nil && page.List != nil {
		return page.List, nil
	}
	var packages []TokenPackage
	if err := json.Unmarshal(raw.Data, &packages); err != nil {
		return nil, fmt.Errorf("decode token packages: %w", err)
	}
	return packages, nil
}

// GetTokenPackageDetail 获取流量包详情
func (c *Client) GetTokenPackageDetail(ctx context.Context, packageID string) (*TokenPackage, error) {
	var resp APIResponse[TokenPackage]
	if err := c.doJSON(ctx, http.MethodGet, "/token-packages/"+packageID, nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// BuyTokenPackage 购买流量包 (创建订单)
func (c *Client) BuyTokenPackage(ctx context.Context, packageID string, payload *PayPayload) (*Order, error) {
	var body interface{}
	if payload != nil {
		body = payload
	}
	var resp APIResponse[Order]
	if err := c.doJSON(ctx, http.MethodPost, "/token-packages/"+packageID+"/buy", body, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetOrderStatus 查询订单支付状态
func (c *Client) GetOrderStatus(ctx context.Context, orderID string) (*OrderStatus, error) {
	var resp APIResponse[OrderStatus]
	if err := c.doJSON(ctx, http.MethodGet, "/token-packages/orders/"+orderID+"/status", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ListMyOrders 查询我的订单列表
// 兼容 yudao 分页格式 {"data":{"list":[...],"total":N}} 和直接数组 {"data":[...]}
func (c *Client) ListMyOrders(ctx context.Context) ([]Order, error) {
	var raw APIResponse[json.RawMessage]
	if err := c.doJSON(ctx, http.MethodGet, "/token-packages/my", nil, &raw, false); err != nil {
		return nil, err
	}
	// 尝试 yudao 分页格式
	var page YudaoPageResult[Order]
	if json.Unmarshal(raw.Data, &page) == nil && page.List != nil {
		return page.List, nil
	}
	// 降级: 直接数组
	var orders []Order
	if err := json.Unmarshal(raw.Data, &orders); err != nil {
		return nil, fmt.Errorf("decode orders: %w", err)
	}
	return orders, nil
}

// ============================================================================
// API: Wallet (scope: wallet:readonly)
// ============================================================================

// GetWalletStats 获取钱包统计 (余额/月消费/月充值)
func (c *Client) GetWalletStats(ctx context.Context) (*WalletStats, error) {
	var resp APIResponse[WalletStats]
	if err := c.doJSON(ctx, http.MethodGet, "/wallet/stats", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetWalletTransactions 获取最近交易记录
func (c *Client) GetWalletTransactions(ctx context.Context) ([]Transaction, error) {
	var resp APIResponse[[]Transaction]
	if err := c.doJSON(ctx, http.MethodGet, "/wallet/transactions", nil, &resp, false); err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// ============================================================================
// API: Skill Store (scope: skill_store / 公共端点)
// ============================================================================

// BrowseSkillStore 浏览技能商店 (公共端点, 无需认证)
// 便捷方法: 等价于 BrowseSkills(ctx, 1, 50, query.Category, query.Keyword, query.Tag, "")
func (c *Client) BrowseSkillStore(ctx context.Context, query SkillStoreQuery) ([]SkillStoreItem, error) {
	resp, err := c.BrowseSkills(ctx, 1, 50, query.Category, query.Keyword, query.Tag, "")
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// BrowseSkills 浏览公共技能商店 (V3 分页接口, 公共端点, 无需认证)
func (c *Client) BrowseSkills(ctx context.Context, page, pageSize int, category, keyword, tag, source string) (*SkillBrowseResponse, error) {
	qv := url.Values{}
	qv.Set("page", fmt.Sprintf("%d", page))
	qv.Set("pageSize", fmt.Sprintf("%d", pageSize))
	if category != "" {
		qv.Set("category", category)
	}
	if keyword != "" {
		qv.Set("keyword", keyword)
	}
	if tag != "" {
		qv.Set("tag", tag)
	}
	if source != "" {
		qv.Set("source", source)
	}

	var resp APIResponse[SkillBrowseResponse]
	if err := c.doPublicJSON(ctx, http.MethodGet, "/skill-store?"+qv.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// BrowseSkillsList 轻量浏览公共技能商店（仅返回标题、简介等展示字段）
// 等价于 BrowseSkills + fields=minimal，响应体积缩减 90%+
func (c *Client) BrowseSkillsList(ctx context.Context, page, pageSize int,
	category, keyword, tag, source string) (*SkillBrowseListResponse, error) {
	qv := url.Values{}
	qv.Set("page", fmt.Sprintf("%d", page))
	qv.Set("pageSize", fmt.Sprintf("%d", pageSize))
	qv.Set("fields", "minimal")
	if category != "" {
		qv.Set("category", category)
	}
	if keyword != "" {
		qv.Set("keyword", keyword)
	}
	if tag != "" {
		qv.Set("tag", tag)
	}
	if source != "" {
		qv.Set("source", source)
	}

	var resp APIResponse[SkillBrowseListResponse]
	if err := c.doPublicJSON(ctx, http.MethodGet, "/skill-store?"+qv.Encode(), nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// GetSkillDetail 获取技能商店中某个技能的详情 (公共端点)
func (c *Client) GetSkillDetail(ctx context.Context, skillID string) (*SkillStoreItem, error) {
	var resp APIResponse[SkillStoreItem]
	if err := c.doPublicJSON(ctx, http.MethodGet, "/skill-store/"+skillID, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ResolveSkill 按 key 精确查找公共技能 (公共端点)
func (c *Client) ResolveSkill(ctx context.Context, key string) (*SkillStoreItem, error) {
	var resp APIResponse[SkillStoreItem]
	if err := c.doPublicJSON(ctx, http.MethodGet, "/skill-store/resolve/"+key, nil, &resp); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// InstallSkill 安装技能到当前用户的租户空间 (需 OAuth scope: skill_store)
func (c *Client) InstallSkill(ctx context.Context, skillID string) (*SkillStoreItem, error) {
	var resp APIResponse[SkillStoreItem]
	if err := c.doJSON(ctx, http.MethodPost, "/skill-store/"+skillID+"/install", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// DownloadSkill 下载技能 ZIP 包 (公共端点, 双模式)
// 有 token 时自动附带 (享受无限流), 无 token 时匿名 (受限流)
// 返回 *RateLimitError 表示 429 限流
// 根因修复 #5: 使用 LimitReader 限制下载体积为 50MB
// [RC-3] 5 分钟超时 (大文件下载)
func (c *Client) DownloadSkill(ctx context.Context, skillID string) ([]byte, string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	token, _ := c.ensureToken(ctx)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.apiURL("/skill-store/"+skillID+"/download"), nil)
	if err != nil {
		return nil, "", err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("download skill: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return nil, "", &RateLimitError{
			Message:    "匿名下载已达限制",
			RetryAfter: resp.Header.Get("Retry-After"),
			Raw:        string(bodyBytes),
		}
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return nil, "", fmt.Errorf("download skill: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// 根因修复 #5: 限制最大下载体积
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxDownloadSize+1))
	if err != nil {
		return nil, "", fmt.Errorf("read download body: %w", err)
	}
	if int64(len(data)) > maxDownloadSize {
		return nil, "", fmt.Errorf("download skill: response exceeds %dMB limit", maxDownloadSize>>20)
	}

	filename := "skill.zip"
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if idx := strings.Index(cd, "filename"); idx != -1 {
			parts := strings.SplitN(cd[idx:], "=", 2)
			if len(parts) == 2 {
				filename = strings.Trim(parts[1], "\"' ")
			}
		}
	}

	return data, filename, nil
}

// UploadSkill 上传技能 ZIP 包
// scope: "TENANT"
// intent: "PERSONAL" (仅自己用) 或 "PUBLIC_INTENT" (走认证→公开)
// 根因修复 #10: retried 参数防止无限递归
func (c *Client) UploadSkill(ctx context.Context, zipData []byte, scope, intent string) (*SkillStoreItem, error) {
	return c.uploadSkillInternal(ctx, zipData, scope, intent, false)
}

// [RC-6] 使用 mime/multipart.Writer 生成随机 boundary, 防止 ZIP 内容碰撞
// [RC-3] 5 分钟超时 (大文件上传)
func (c *Client) uploadSkillInternal(ctx context.Context, zipData []byte, scope, intent string, retried bool) (*SkillStoreItem, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	token, err := c.ensureToken(ctx)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	_ = writer.WriteField("scope", scope)
	_ = writer.WriteField("intent", intent)
	part, err := writer.CreateFormFile("file", "skill.zip")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(zipData); err != nil {
		return nil, fmt.Errorf("write zip data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.apiURL("/skill-store/upload"), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 根因修复 #10: 401 时刷新 token 并重试一次, retried 防无限递归
	if resp.StatusCode == http.StatusUnauthorized && !retried {
		resp.Body.Close()
		if refreshErr := c.forceRefresh(ctx); refreshErr != nil {
			return nil, fmt.Errorf("upload: unauthorized and refresh failed: %w", refreshErr)
		}
		return c.uploadSkillInternal(ctx, zipData, scope, intent, true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return nil, fmt.Errorf("upload: HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data struct {
			Skill SkillStoreItem `json:"skill"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return &result.Data.Skill, nil
}

// GetSkillSummary 获取技能统计概览 (installed/created/total/storeAvailable)
func (c *Client) GetSkillSummary(ctx context.Context) (*SkillSummary, error) {
	var resp APIResponse[SkillSummary]
	if err := c.doJSON(ctx, http.MethodGet, "/skills/summary", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// CertifySkill 触发技能认证管线 (异步)
func (c *Client) CertifySkill(ctx context.Context, skillID string) error {
	return c.doJSON(ctx, http.MethodPost, "/skill-store/"+skillID+"/certify", nil, nil, false)
}

// GetCertificationStatus 查询技能认证状态
func (c *Client) GetCertificationStatus(ctx context.Context, skillID string) (*CertificationStatus, error) {
	var resp APIResponse[CertificationStatus]
	if err := c.doJSON(ctx, http.MethodGet, "/skill-store/"+skillID+"/certification", nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ============================================================================
// API: Skill Generator (scope: skill_store)
// ============================================================================

// GenerateSkill 根据自然语言描述生成技能定义 (基于独立 LLM)
func (c *Client) GenerateSkill(ctx context.Context, req GenerateSkillRequest) (*GenerateSkillResult, error) {
	var resp APIResponse[GenerateSkillResult]
	if err := c.doJSON(ctx, http.MethodPost, "/skill-generator/generate", req, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// OptimizeSkill 优化已有技能定义
func (c *Client) OptimizeSkill(ctx context.Context, req OptimizeSkillRequest) (*OptimizeSkillResult, error) {
	var resp APIResponse[OptimizeSkillResult]
	if err := c.doJSON(ctx, http.MethodPost, "/skill-generator/optimize", req, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ValidateSkill 校验技能定义正确性
func (c *Client) ValidateSkill(ctx context.Context, skillName string) error {
	body := map[string]string{"skillName": skillName}
	return c.doJSON(ctx, http.MethodPost, "/skill-generator/validate", body, nil, false)
}

// ============================================================================
// API: Unified Tools (scope: tools)
// ============================================================================

// ListTools 获取当前用户租户下的所有工具 (Skill 优先 + Plugin 兜底)
func (c *Client) ListTools(ctx context.Context) ([]ToolView, error) {
	var resp APIResponse[ToolListResponse]
	if err := c.doJSON(ctx, http.MethodGet, "/tools", nil, &resp, false); err != nil {
		return nil, err
	}
	return resp.Data.Skills, nil
}

// GetTool 获取单个工具详情
func (c *Client) GetTool(ctx context.Context, toolID string) (*ToolView, error) {
	var resp APIResponse[ToolView]
	if err := c.doJSON(ctx, http.MethodGet, "/tools/"+toolID, nil, &resp, false); err != nil {
		return nil, err
	}
	return &resp.Data, nil
}

// ============================================================================
// Internal HTTP
// ============================================================================

// 根因修复 #11: 使用 strings.HasSuffix 精确匹配, 防止域名含 /api/v4 子串时误判
func (c *Client) apiURL(path string) string {
	base := c.serverURL
	if !strings.HasSuffix(base, "/api/v4") {
		base += "/api/v4"
	}
	return base + path
}

// 根因修复 #3: doJSON 增加 retried 参数, 401 重试只允许一次, 防无限递归栈溢出
// [RC-3] per-request 30s 超时 (替代全局 http.Client.Timeout)
func (c *Client) doJSON(ctx context.Context, method, path string, body interface{}, result interface{}, retried bool) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	token, err := c.ensureToken(ctx)
	if err != nil {
		return err
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), bodyReader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	// 根因修复 #3: retried 防止无限递归 (refresh 成功但服务端仍返 401 → 栈溢出)
	if resp.StatusCode == http.StatusUnauthorized && !retried {
		resp.Body.Close()
		if refreshErr := c.forceRefresh(ctx); refreshErr != nil {
			return fmt.Errorf("unauthorized and refresh failed: %w", refreshErr)
		}
		return c.doJSON(ctx, method, path, body, result, true)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// doPublicJSON 公共端点请求
// 有 token 时自动附带 (享受认证用户待遇), 无 token 时匿名请求
// 不做 401 重试 (公共端点不应要求认证)
// [RC-3] per-request 30s 超时
func (c *Client) doPublicJSON(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	token, _ := c.ensureToken(ctx)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.apiURL(path), bodyReader)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("request %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBodySize))
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

// 根因修复 #6: forceRefresh 不再静默吞掉 store.Save 错误, 打印警告
func (c *Client) forceRefresh(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tokens == nil {
		return fmt.Errorf("no tokens to refresh")
	}
	if c.meta == nil {
		meta, err := Discover(ctx, c.serverURL)
		if err != nil {
			return err
		}
		c.meta = meta
	}

	tokenResp, err := RefreshToken(ctx, c.meta, c.tokens.ClientID, c.tokens.RefreshToken)
	if err != nil {
		return err
	}

	c.tokens = NewTokenSet(tokenResp, c.tokens.ClientID, c.serverURL)
	if saveErr := c.store.Save(c.tokens); saveErr != nil {
		fmt.Printf("[acosmi-sdk] warning: save refreshed token failed: %v\n", saveErr)
	}
	return nil
}
