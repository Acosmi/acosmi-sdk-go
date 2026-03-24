package acosmi

// 分组 Scope — 与后端 DesktopOAuthScopes 保持一致 (V2: 10→3 合并)
const (
	ScopeAI      = "ai"      // 模型服务: 模型调用 + 流量包 + 权益
	ScopeSkills  = "skills"  // 技能与工具: 技能商店 + 工具列表 + 执行
	ScopeAccount = "account" // 账户信息: 个人资料 + 钱包余额 + 交易记录
)

// Deprecated: 旧细粒度 scope, 保留向后兼容, 新代码请用分组 scope
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

// AllScopes 全部分组 scope (推荐)
func AllScopes() []string {
	return []string{ScopeAI, ScopeSkills, ScopeAccount}
}

// ModelScopes 模型服务相关 scope
func ModelScopes() []string {
	return []string{ScopeAI}
}

// CommerceScopes 商城/钱包 scope
func CommerceScopes() []string {
	return []string{ScopeAI, ScopeAccount}
}

// SkillScopes 技能/工具 scope
func SkillScopes() []string {
	return []string{ScopeSkills}
}
