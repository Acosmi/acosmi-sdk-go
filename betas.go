package acosmi

// Beta Header 常量 — 经联网验证的真实 Anthropic API beta 值
// 虚构/错误日期的 header 已剔除 (task-budgets / advisor-tool / cache-editing / web-search beta)
const (
	betaClaudeCode          = "claude-code-20250219"              // 始终: 代理标识
	betaInterleavedThinking = "interleaved-thinking-2025-05-14"   // ISP: 交错思考
	betaContext1M           = "context-1m-2025-08-07"             // 1M 上下文 (retiring 2026-04-30)
	betaContextManagement   = "context-management-2025-06-27"     // 上下文编辑
	betaStructuredOutputs   = "structured-outputs-2025-11-13"     // 结构化输出
	betaAdvancedToolUse     = "advanced-tool-use-2025-11-20"      // Tool Search
	betaEffort              = "effort-2025-11-24"                 // Effort 控制 (Opus 4.5 需要, 4.6 stable)
	betaPromptCachingScope  = "prompt-caching-scope-2026-01-05"   // 缓存作用域隔离
	betaFastMode            = "fast-mode-2026-02-01"              // 快速推理 (Opus 4.6)
	betaRedactThinking      = "redact-thinking-2026-02-12"        // 思考脱敏
	betaTokenEfficientTools = "token-efficient-tools-2025-02-19"  // 高效工具 (Claude 3.7, Claude 4 内置)
)

// buildBetas 根据模型能力和请求参数自动组装 beta header 列表
// 内部方法, 在 buildChatRequest 中调用
func buildBetas(caps ModelCapabilities, req *ChatRequest) []string {
	betas := []string{betaClaudeCode} // 始终包含

	// ── 思考相关 ──
	if caps.SupportsISP {
		betas = append(betas, betaInterleavedThinking)
		betas = append(betas, betaContextManagement)
	}
	if caps.SupportsRedactThinking && req.Thinking != nil && req.Thinking.Display == "summary" {
		betas = append(betas, betaRedactThinking)
	}

	// ── 上下文 ──
	if caps.Supports1MContext {
		betas = append(betas, betaContext1M)
	}

	// ── 输出控制 (互斥: structured-outputs ⊕ token-efficient-tools) ──
	hasStructuredOutput := caps.SupportsStructuredOutput && req.OutputConfig != nil
	if hasStructuredOutput {
		betas = append(betas, betaStructuredOutputs)
	} else if caps.SupportsTokenEfficient {
		// 仅在不使用 structured-outputs 时注入
		betas = append(betas, betaTokenEfficientTools)
	}

	// ── Tool Search ──
	if caps.SupportsToolSearch {
		betas = append(betas, betaAdvancedToolUse)
	}

	// ── 推理控制 ──
	if caps.SupportsEffort && req.Effort != nil {
		betas = append(betas, betaEffort)
	}
	if caps.SupportsFastMode && req.Speed == "fast" {
		betas = append(betas, betaFastMode)
	}

	// ── 缓存 ──
	if caps.SupportsPromptCache {
		betas = append(betas, betaPromptCachingScope)
	}

	// ── 合并客户端显式传入 (去重) ──
	return uniqueMerge(betas, req.Betas)
}

// uniqueMerge 合并两个字符串切片并去重, 保留顺序
func uniqueMerge(base, extra []string) []string {
	if len(extra) == 0 {
		return base
	}
	seen := make(map[string]struct{}, len(base)+len(extra))
	for _, s := range base {
		seen[s] = struct{}{}
	}
	for _, s := range extra {
		if _, ok := seen[s]; !ok {
			base = append(base, s)
			seen[s] = struct{}{}
		}
	}
	return base
}
