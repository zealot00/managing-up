# Issues & Known Bugs

## Phase 1: sop-to-skill CLI

### 已修复
- P1: Markdown 残留 in descriptions → 添加 `cleanDescription()` 函数
- P2: 无效 tool_ref → 改用 `type=condition`
- P3: SkillSchema 格式不兼容 SEH → 添加顶层 name/version/risk_level

---

## 前端问题

### Login 无法登录
- **状态**: 待修复
- **现象**: 访问 /login 无法登录
- **可能原因**: 
  - API 认证端点未实现
  - 前端 AuthContext 配置问题
  - CORS 问题
- **相关文件**: 
  - `/apps/web/app/login/page.tsx`
  - `/apps/web/context/AuthContext.tsx`
- **建议**: 检查 API `/api/v1/auth/login` 端点是否存在

---

## API 服务问题

### generate-skill / generate-from-extracted LLM 生成失败
- **状态**: 待修复
- **现象**: 使用 deepseek-r1:1.5b 模型生成的 YAML 格式错误
- **原因**: 小模型无法遵循复杂 YAML schema
- **建议**: 配置更强的模型 (deepseek-coder, gpt-4o)
