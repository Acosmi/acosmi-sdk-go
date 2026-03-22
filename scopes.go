package acosmi

// Scope 常量 — 与后端 DesktopOAuthScopes 保持一致
const (
	ScopeModels        = "models"
	ScopeModelsChat    = "models:chat"
	ScopeEntitlements  = "entitlements"
	ScopeTokenPackages = "token-packages"
	ScopeSkillStore    = "skill_store"
	ScopeTools         = "tools"
	ScopeToolsExecute  = "tools:execute"
	ScopeWallet        = "wallet"
	ScopeWalletReadonly = "wallet:readonly"
	ScopeProfile       = "profile"
)

// [RC-9] Scope 预设组改为函数, 返回新切片, 防止外部篡改

// AllScopes 全部可用 scope
func AllScopes() []string {
	return []string{
		ScopeModels, ScopeModelsChat,
		ScopeEntitlements, ScopeTokenPackages,
		ScopeSkillStore, ScopeTools, ScopeToolsExecute,
		ScopeWallet, ScopeWalletReadonly,
		ScopeProfile,
	}
}

// ModelScopes 模型商业化相关 scope
func ModelScopes() []string {
	return []string{
		ScopeModels, ScopeModelsChat,
		ScopeEntitlements,
	}
}

// CommerceScopes 商城/钱包 scope
func CommerceScopes() []string {
	return []string{
		ScopeModels,
		ScopeEntitlements,
		ScopeTokenPackages,
		ScopeWalletReadonly,
	}
}

// SkillScopes 技能/工具 scope
func SkillScopes() []string {
	return []string{
		ScopeSkillStore,
		ScopeTools,
	}
}
