# 从追踪构建任务 — 向导式页面重设计

## 背景

`/tasks/from-trace` 页面当前是一个朴素的单页表单，存在四个问题：

1. **表单交互粗糙** — 模式切换用普通按钮，输入框没有引导感
2. **结果展示单薄** — 构建成功后只平铺信息，缺少后续操作
3. **设计语言不统一** — 没用到 `detail-header`/`detail-chip` 等系统新组件
4. **双入口体验割裂** — 独立页面和 Drawer 版本信息重复（eyebrow/title 出现两次）

## 方案：卡片式三步向导

### 整体结构

```
步骤指示器: ① 选择来源 → ② 确认构建 → ③ 查看结果
内容区:     当前步骤的内容（卡片面板）
操作区:     [上一步] / [下一步] / [开始构建] / [查看任务]
```

- 步骤指示器用 `wizard-steps` CSS 类，当前步骤高亮，已完成步骤带 ✓
- 内容区同一时刻只展示一个步骤
- 独立页面用 `Breadcrumb + PageHeader` 外壳，Drawer 版本共用步骤组件

### 步骤 1：选择来源

- 两个大卡片并排："按执行 ID"（Play 图标）/ "按追踪 ID"（GitBranch 图标）
- 选中卡片边框高亮 `--primary`，未选中 `--line`
- 选中后卡片内部展开输入框 + hint 文字
- 从 Drawer 打开时（已有 `initialExecutionId`），自动选中"按执行 ID"并预填 ID
- "下一步"按钮在输入有效 ID 后可点击

### 步骤 2：确认构建

- 来源摘要：用 `detail-chip` 展示来源类型和 ID
- 说明文字：告知用户系统将自动提取输入参数、预期输出、任务名称和难度等级
- "开始构建"按钮带 Sparkles 图标 + loading 动画

### 步骤 3：查看结果

- 成功提示
- 用 `detail-header` + `detail-chip` 展示任务元数据（名称、difficulty badge、ID、test case 数量）
- 测试用例表格用 `gateway-table` 样式
- 两个操作："继续构建"（回到步骤 1）和 "查看任务"（跳转 `/tasks?highlight=<id>`）

## 涉及文件

| 文件 | 变更 |
|---|---|
| `apps/web/app/tasks/from-trace/page.tsx` | 重写为向导页面外壳 |
| `apps/web/components/TaskFromTraceForm.tsx` | 重写为三步向导组件 `TaskFromTraceWizard` |
| `apps/web/app/components/TaskFromTraceDrawer.tsx` | 改用向导组件，去掉冗余标题 |
| `apps/web/app/globals.css` | 新增 `wizard-steps`、`wizard-step`、`wizard-step-active`、`wizard-step-done`、`wizard-source-card`、`wizard-source-card-active` 样式 |

## CSS 新增

```css
/* ===== Wizard Steps ===== */
.wizard-steps       -- 步骤指示器，水平排列
.wizard-step         -- 单个步骤，灰色文字
.wizard-step-active  -- 当前步骤，primary 高亮
.wizard-step-done    -- 已完成步骤，带 ✓

/* ===== Wizard Source Cards ===== */
.wizard-source-card           -- 来源选择卡片
.wizard-source-card-active    -- 选中状态
```

## 组件拆分

```
TaskFromTraceWizard (新，替换 TaskFromTraceForm)
  ├── WizardStepIndicator (步骤指示器)
  ├── Step1SourceSelect (来源选择 + 输入)
  ├── Step2Confirm (确认构建)
  └── Step3Result (查看结果，复用 detail-header/chip)
```

## Drawer 适配

- Drawer 版本复用 `TaskFromTraceWizard`，传入 `initialExecutionId`
- Drawer 已有标题，步骤组件内不再重复显示标题
- 步骤 3 的"查看任务"按钮在 Drawer 场景下关闭 Drawer 后跳转
