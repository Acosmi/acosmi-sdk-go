// adapter_anthropic.go — Anthropic 原生格式 adapter
//
// 重构自 buildChatRequest 现有逻辑，功能完全等同。
// 包含 buildBetas() 调用、ServerTools 合入、ExtraBody 透传。

package acosmi

import (
	"encoding/json"
	"fmt"
)

// AnthropicAdapter 实现 ProviderAdapter，用于 Anthropic 原生模型
type AnthropicAdapter struct{}

func (a *AnthropicAdapter) Format() ProviderFormat { return FormatAnthropic }
func (a *AnthropicAdapter) EndpointSuffix() string  { return "/anthropic" }

// BuildRequestBody 构建 Anthropic 格式请求体
// 逻辑等同于原 buildChatRequest，包含完整的 betas/tools/ServerTools/ExtraBody 处理
func (a *AnthropicAdapter) BuildRequestBody(caps ModelCapabilities, req *ChatRequest) (map[string]any, error) {
	body := make(map[string]any)

	// 消息: RawMessages 优先于 Messages
	if req.RawMessages != nil {
		body["messages"] = req.RawMessages
	} else if len(req.Messages) > 0 {
		body["messages"] = req.Messages
	}

	body["stream"] = req.Stream

	if req.MaxTokens > 0 {
		body["max_tokens"] = req.MaxTokens
	}
	if req.System != nil {
		body["system"] = req.System
	}
	if req.Temperature != nil {
		body["temperature"] = *req.Temperature
	}
	if req.Thinking != nil {
		body["thinking"] = req.Thinking
	}
	if req.Metadata != nil {
		body["metadata"] = req.Metadata
	}

	// ── 合入 Tools + ServerTools ──
	var allTools []any
	if req.Tools != nil {
		if toolsJSON, err := json.Marshal(req.Tools); err == nil {
			var parsed []any
			if json.Unmarshal(toolsJSON, &parsed) == nil {
				allTools = append(allTools, parsed...)
			}
		}
	}
	for _, st := range req.ServerTools {
		schema := map[string]any{
			"type": st.Type,
			"name": st.Name,
		}
		for k, v := range st.Config {
			schema[k] = v
		}
		allTools = append(allTools, schema)
	}
	if len(allTools) > 0 {
		body["tools"] = allTools
	}

	// ── 推理控制 ──
	if req.Speed != "" {
		body["speed"] = req.Speed
	}
	if req.Effort != nil {
		body["effort"] = req.Effort
	}
	if req.OutputConfig != nil {
		body["output_config"] = req.OutputConfig
	}

	// ── Beta 自动组装 ──
	betas := buildBetas(caps, req)
	if len(betas) > 0 {
		body["betas"] = betas
	}

	// ── 透传 ExtraBody ──
	for k, v := range req.ExtraBody {
		body[k] = v
	}

	return body, nil
}

// ParseResponse 解析 Anthropic 格式同步响应
// 兼容 APIResponse 包装 {"code":0,"data":{...}} 和裸 Anthropic JSON 两种格式
func (a *AnthropicAdapter) ParseResponse(body []byte) (*ChatResponse, error) {
	// 尝试 APIResponse 包装
	var wrapper struct {
		Code    int             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	raw := body
	if json.Unmarshal(body, &wrapper) == nil && len(wrapper.Data) > 0 && string(wrapper.Data) != "null" {
		if wrapper.Code != 0 {
			return nil, &BusinessError{Code: wrapper.Code, Message: wrapper.Message}
		}
		raw = wrapper.Data
	}

	var resp ChatResponse
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decode anthropic response: %w", err)
	}
	resp.TokenRemaining = -1
	resp.CallRemaining = -1
	return &resp, nil
}

// ParseStreamLine 解析 Anthropic SSE 行
// Anthropic 原生协议无 [DONE]（message_stop 后上游关闭连接），
// 但 Nexus Gateway 的 ChatStream 路径会追加 [DONE] 哨兵，此处一并处理。
func (a *AnthropicAdapter) ParseStreamLine(eventType, data string) (StreamEvent, bool, error) {
	if data == "[DONE]" {
		return StreamEvent{}, true, nil
	}
	return StreamEvent{Event: eventType, Data: data}, false, nil
}
