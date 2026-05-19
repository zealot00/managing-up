# 网关密钥管理重构设计

日期：2026-05-18

## 背景

当前网关管理页面的密钥管理存在两个问题：

1. **创建后页面跳转**：点击创建密钥后，`loadData()` 触发整体重新渲染，页面滚动位置重置到顶部，而生成的密钥显示在页面底部表单区域，用户看不到。
2. **密钥名称允许重复**：后端不做唯一性校验，同用户可创建多个同名密钥，造成管理混淆。

## 设计决策

### 密钥创建 → 模态对话框

采用模态对话框完成密钥创建全流程：

**状态机**：
```
idle → (点击创建按钮) → open → (提交) → submitting → (成功) → success → (关闭) → idle
                                                      ↘ (失败) → open (显示错误)
```

**idle/open 阶段**：
- 标题「创建密钥」
- 名称输入框（前端校验：非空 + 不与现有未吊销密钥重名）
- 底部操作栏：取消 / 创建按钮

**success 阶段**：
- 标题「密钥已创建」
- 警告文字「仅显示一次，请立即复制」
- 密钥代码块 + 复制按钮（`navigator.clipboard.writeText()`，复制后按钮文字变为「已复制 ✓」，2 秒后恢复）
- 底部操作栏：「我已保存密钥」关闭按钮

### 密钥列表重构

- 移除左右两栏布局（`gateway-keys-grid`），改为全宽面板
- 面板头部：标题「密钥管理」+ 右侧「创建密钥」按钮
- 已吊销密钥行弱化（`opacity: 0.5`）
- 吊销操作增加 `ConfirmDialog` 二次确认
- 空状态使用 `EmptyState` 组件

### 重名校验

- 前端校验：提交前检查当前已加载的 `keys` 列表，与未吊销密钥对比
- 重名时输入框下方显示 `form-error`

## i18n 新增键

| 键名 | en | zh |
|---|---|---|
| `createKeyTitle` | Create API Key | 创建 API 密钥 |
| `keyCreated` | Key Created | 密钥已创建 |
| `copyKey` | Copy | 复制 |
| `copied` | Copied ✓ | 已复制 ✓ |
| `keySaved` | I've saved the key | 我已保存密钥 |
| `duplicateName` | A key with this name already exists | 已存在同名密钥 |
| `revokeConfirmTitle` | Revoke Key | 吊销密钥 |
| `revokeConfirmDesc` | Are you sure you want to revoke key "{name}"? Requests using this key will be rejected. | 确定吊销密钥「{name}」？吊销后使用该密钥的请求将被拒绝。 |
| `emptyTitle` | No API Keys | 暂无 API 密钥 |
| `emptyDesc` | Create an API key to start using the LLM Gateway. | 创建一个密钥以开始使用 LLM 网关。 |
| `createFirstKey` | Create Key | 创建密钥 |

## CSS 变更

- **移除**：`gateway-keys-grid`、`gateway-keys-form`、`gateway-keys-list`
- **保留**：`gateway-secret`、`gateway-secret-title`、`gateway-secret-code`（模态框内复用）
- **新增**：
  - `.gateway-keys-header` — flex 布局，标题左对齐 + 按钮右对齐
  - `.revoked-row` — `opacity: 0.5` 弱化已吊销行

## 文件变更清单

| 文件 | 变更 |
|---|---|
| `apps/web/app/gateway/page.tsx` | 移除内联表单，添加创建按钮 + CreateKeyDialog + ConfirmDialog 吊销 |
| `apps/web/app/gateway/CreateKeyDialog.tsx` | 新建 — 模态框组件 |
| `apps/web/app/globals.css` | 移除 keys-grid/form/list，新增 keys-header/revoked-row |
| `apps/web/messages/en.json` | 新增 11 个 gateway 键 |
| `apps/web/messages/zh.json` | 新增 11 个 gateway 键 |
