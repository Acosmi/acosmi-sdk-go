// adapter.go — Provider Adapter 接口与注册表
//
// SDK 层 per-provider 路由：
//   Provider == "anthropic" → AnthropicAdapter → POST /managed-models/:id/anthropic
//   Provider != "anthropic" → OpenAIAdapter    → POST /managed-models/:id/chat
//
// SDK 只负责：格式路由 + 请求结构转换 + 响应结构转换
// 厂商特定参数（如 enable_thinking）由 Nexus Gateway per-provider adapter 处理

package acosmi

// ProviderFormat 标识请求格式
type ProviderFormat int

const (
	FormatAnthropic ProviderFormat = iota // Anthropic 原生格式
	FormatOpenAI                          // OpenAI 兼容格式
)

// ProviderAdapter 定义了将 ChatRequest 转换为特定格式的接口
type ProviderAdapter interface {
	// Format 返回此 adapter 使用的请求格式
	Format() ProviderFormat

	// EndpointSuffix 返回 API 路径后缀
	// Anthropic: "/anthropic", OpenAI: "/chat"
	EndpointSuffix() string

	// BuildRequestBody 将 ChatRequest 转换为 HTTP body (map → json.Marshal)
	// caps 用于条件化字段注入（如 betas）
	BuildRequestBody(caps ModelCapabilities, req *ChatRequest) (map[string]any, error)

	// ParseResponse 解析同步响应 body 为 ChatResponse
	ParseResponse(body []byte) (*ChatResponse, error)

	// ParseStreamLine 解析一行 SSE data 为 StreamEvent
	// 返回 (event, done, error)
	// done=true 表示流结束（遇到 [DONE] 或 message_stop）
	ParseStreamLine(eventType, data string) (StreamEvent, bool, error)
}

// adapterRegistry 按 provider 名称映射 adapter
var adapterRegistry = map[string]ProviderAdapter{
	"anthropic": &AnthropicAdapter{},
	"acosmi":    &AnthropicAdapter{}, // Acosmi 自有模型走 Anthropic 格式
}

// defaultOpenAIAdapter 是非 Anthropic 厂商的默认 adapter
var defaultOpenAIAdapter ProviderAdapter = &OpenAIAdapter{}

// getAdapter 根据 provider 返回对应的 adapter
func getAdapter(provider string) ProviderAdapter {
	if a, ok := adapterRegistry[provider]; ok {
		return a
	}
	return defaultOpenAIAdapter
}
