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

// AllScopes 全部可用 scope
var AllScopes = []string{
	ScopeModels, ScopeModelsChat,
	ScopeEntitlements, ScopeTokenPackages,
	ScopeSkillStore, ScopeTools, ScopeToolsExecute,
	ScopeWallet, ScopeWalletReadonly,
	ScopeProfile,
}

// ModelScopes 模型商业化相关 scope
var ModelScopes = []string{
	ScopeModels, ScopeModelsChat,
	ScopeEntitlements,
}

// CommerceScopes 商城/钱包 scope
var CommerceScopes = []string{
	ScopeModels,
	ScopeEntitlements,
	ScopeTokenPackages,
	ScopeWalletReadonly,
}

// SkillScopes 技能/工具 scope
var SkillScopes = []string{
	ScopeSkillStore,
	ScopeTools,
}
