# sop-to-skill Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 构建一个 TypeScript CLI 工具，将 SOP 文档转换为可被 AI Agent 执行的 Skill Package。

**Architecture:**
- 核心流程：`Parser → Extractor → LLM Enhancer → Generator → Package`
- CLI 基于 clipanion 构建，提供 `generate`、`extract`、`llm-enhance`、`validate` 命令
- LLM 增强模块支持多种 API（OpenAI、Ollama、本地模型）
- 输出为标准化的 Skill Package，包含人读文档和机器可解析的 Schema

**Tech Stack:**
- TypeScript + Node.js (>=18)
- clipanion + typanion (CLI)
- llamaindex / fetch (LLM API)
- marked (Markdown 解析)
- pdf-parse (PDF 解析)
- mammoth (DOCX 解析)
- js-yaml (YAML 处理)
- adm-zip (ZIP 打包)
- zod (类型验证)
- vitest (测试)

---

## Phase 0: 项目初始化

### Task 0.1: 初始化项目结构

**Files:**
- Create: `sop-to-skill/package.json`
- Create: `sop-to-skill/tsconfig.json`
- Create: `sop-to-skill/.gitignore`
- Create: `sop-to-skill/README.md`

**Step 1: Create package.json**

```json
{
  "name": "sop-to-skill",
  "version": "1.0.0",
  "description": "Convert SOP documents to executable Skills using AI",
  "type": "module",
  "main": "./dist/index.js",
  "bin": {
    "sop-to-skill": "./dist/cli.js"
  },
  "scripts": {
    "build": "tsc",
    "dev": "tsc --watch",
    "test": "vitest run",
    "test:watch": "vitest",
    "lint": "eslint src --ext .ts",
    "typecheck": "tsc --noEmit"
  },
  "keywords": ["sop", "skill", "ai-agent", "cli"],
  "license": "MIT",
  "dependencies": {
    "clipanion": "^4.0.0",
    "typanion": "^3.0.0",
    "llamaindex": "^0.5.0",
    "marked": "^12.0.0",
    "mammoth": "^1.6.0",
    "pdf-parse": "^1.1.1",
    "js-yaml": "^4.1.0",
    "adm-zip": "^0.5.10",
    "zod": "^3.22.0"
  },
  "devDependencies": {
    "@types/node": "^20.11.0",
    "@types/js-yaml": "^4.0.9",
    "@types/adm-zip": "^0.5.5",
    "typescript": "^5.4.0",
    "vitest": "^1.4.0",
    "eslint": "^8.57.0",
    "@typescript-eslint/parser": "^7.0.0",
    "@typescript-eslint/eslint-plugin": "^7.0.0"
  },
  "engines": {
    "node": ">=18.0.0"
  }
}
```

**Step 2: Create tsconfig.json**

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "lib": ["ES2022"],
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "resolveJsonModule": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist", "**/*.test.ts"]
}
```

**Step 3: Create README.md**

```markdown
# sop-to-skill

Convert SOP documents to executable AI Agent Skills.

## Install

```bash
npm install -g sop-to-skill
```

## Quick Start

```bash
# Generate Skill Package from SOP
sop-to-skill generate ./SOP-DM-002.md --name "数据核查流程" --output ./output

# With LLM enhancement
sop-to-skill generate ./SOP-DM-002.md --name "数据核查流程" --output ./output --llm

# Extract structured data only
sop-to-skill extract ./SOP-DM-002.md --output ./extracted.json
```

## Commands

- `generate` - Generate Skill Package from SOP
- `extract` - Extract structured data from SOP
- `llm-enhance` - Enhance extracted data with LLM
- `validate` - Validate Skill Package structure

## Options

- `--llm` - Enable LLM enhancement
- `--llm-api` - LLM API URL (default: http://localhost:11434)
- `--llm-model` - LLM model name (default: gpt-4o)
```

**Step 4: Commit**

```bash
cd sop-to-skill
git init
git add package.json tsconfig.json .gitignore README.md
git commit -m "chore: initialize sop-to-skill project"
```

---

### Task 0.2: 创建核心类型定义

**Files:**
- Create: `sop-to-skill/src/types/skill-package.ts`
- Create: `sop-to-skill/src/types/action.ts`
- Create: `sop-to-skill/src/types/constraint.ts`
- Create: `sop-to-skill/src/types/test-case.ts`
- Create: `sop-to-skill/src/types/extracted.ts`
- Create: `sop-to-skill/src/types/index.ts`

**Step 1: Create src/types/skill-package.ts**

```typescript
import { z } from 'zod';

// Constraint Level Enum
export const ConstraintLevel = {
  MUST: 'MUST',
  SHOULD: 'SHOULD',
  MAY: 'MAY',
} as const;
export type ConstraintLevel = typeof ConstraintLevel[keyof typeof ConstraintLevel];

// Validation Types
export const ValidationType = {
  ASSERT: 'assert',
  ROLE_CHECK: 'role_check',
  DEADLINE: 'deadline',
  THRESHOLD: 'threshold',
} as const;
export type ValidationType = typeof ValidationType[keyof typeof ValidationType];

// Validation Rule
export interface ValidationRule {
  type: ValidationType;
  condition?: string;
  error_code: string;
  error_message?: string;
  // For deadline type
  normal?: { value: number; unit: string };
  urgent?: { value: number; unit: string };
  urgent_condition?: string;
  // For threshold type
  metric?: string;
  threshold?: number;
  action?: string;
}

// Constraint
export interface Constraint {
  id: string;
  level: ConstraintLevel;
  description: string;
  condition?: string;
  action?: string;
  validation?: ValidationRule;
  roles: string[];
  confidence: number;
  source_quote?: string;
}

// Decision Rule
export interface DecisionRule {
  when: Record<string, string>;
  then: Record<string, string>;
}

// Decision Table
export interface Decision {
  id: string;
  name: string;
  inputVars: string[];
  outputVars: string[];
  rules: DecisionRule[];
}

// Parameter
export interface Parameter {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'enum';
  description: string;
  minValue?: number;
  maxValue?: number;
  defaultValue?: any;
  unit?: string;
}

// Source
export interface Source {
  type: 'sop' | 'policy' | 'guideline';
  fileName: string;
  section?: string;
  url?: string;
}

// Trigger Types
export const TriggerType = {
  EXECUTION: 'execution',
  QUERY: 'query',
  APPROVAL: 'approval',
  EVENT: 'event',
} as const;
export type TriggerType = typeof TriggerType[keyof typeof TriggerType];

// Trigger
export interface Trigger {
  type: TriggerType;
  name: string;
  description?: string;
  input_schema: JSONSchema;
  output_schema?: JSONSchema;
}

// JSON Schema (simplified)
export interface JSONSchema {
  type: string;
  properties?: Record<string, any>;
  required?: string[];
  enum?: any[];
  description?: string;
  items?: any;
  format?: string;
}

// Step
export interface Step {
  id: string;
  name: string;
  description: string;
  action: string;
  condition?: string;
  input: Record<string, any>;
  output: Record<string, any>;
  next_step_on_success?: string;
  next_step_on_failure?: string;
  on_failure?: string;
}

// Error Rule Action
export interface ErrorRuleAction {
  type: string;
  description: string;
  roles?: string[];
}

// Error Rule
export interface ErrorRule {
  condition: string;
  actions: ErrorRuleAction[];
}

// Error Handling
export interface ErrorHandling {
  rules: ErrorRule[];
}

// Skill Meta
export interface SkillMeta {
  name: string;
  version: string;
  description?: string;
  source?: string;
  tags?: string[];
  generated_at?: string;
}

// Skill Schema
export interface SkillSchema {
  meta: SkillMeta;
  triggers: Trigger[];
  steps: Step[];
  constraints: Constraint[];
  decisions?: Decision[];
  parameters?: Parameter[];
  error_handling?: ErrorHandling;
  sources?: Source[];
}

// Skill Manifest
export interface Manifest {
  package: {
    name: string;
    version: string;
    type: 'skill';
  };
  source: {
    type: string;
    file: string;
    section?: string;
    parsed_at: string;
  };
  generation: {
    method: string;
    llm_enhanced: boolean;
    llm_model?: string;
  };
  files: Array<{
    path: string;
    type: string;
    count?: number;
    size?: number;
  }>;
  quality?: {
    constraint_coverage?: number;
    test_case_coverage?: number;
    constraint_confidence_avg?: number;
  };
}

// Complete Skill Package
export interface SkillPackage {
  schema: SkillSchema;
  manifest: Manifest;
}
```

**Step 2: Create src/types/action.ts**

```typescript
import { JSONSchema } from './skill-package.js';

export interface ActionImplementation {
  type: 'http_call' | 'function' | 'script' | 'tool';
  method?: string;
  endpoint?: string;
  timeout?: string;
  retry?: {
    max_attempts: number;
    backoff: 'exponential' | 'linear';
  };
  command?: string;
  interpreter?: string;
}

export interface ActionCondition {
  type: 'role' | 'record_exists' | 'assert' | 'custom';
  required_role?: string;
  record_id?: string;
  condition?: string;
}

export interface Action {
  action: string;
  description: string;
  input: {
    type: 'object';
    properties: Record<string, any>;
    required?: string[];
    description?: string;
    default?: any;
  };
  output: {
    type: 'object';
    properties: Record<string, any>;
    description?: string;
  };
  implementation: ActionImplementation;
  validation?: {
    pre_conditions: ActionCondition[];
    post_conditions: ActionCondition[];
  };
}
```

**Step 3: Create src/types/test-case.ts**

```typescript
export const TestCaseType = {
  HAPPY_PATH: 'happy-path',
  EDGE_CASE: 'edge-case',
  ERROR_CASE: 'error-case',
  COMPLIANCE: 'compliance',
} as const;
export type TestCaseType = typeof TestCaseType[keyof typeof TestCaseType];

export interface ExpectedOutput {
  success: boolean;
  outputPattern?: string;
  validationSteps?: string[];
  errorType?: string;
  handling?: string;
  [key: string]: any;
}

export interface Assertion {
  path: string;
  operator: 'equals' | 'not_equals' | 'contains' | 'length_equals' | 'length_gte' | 'length_lte' | 'gte' | 'lte';
  value: any;
}

export interface TestCaseValidation {
  assertions: Assertion[];
}

export interface TestCase {
  case_id: string;
  type: TestCaseType;
  skill: string;
  description: string;
  input: Record<string, any>;
  expected: ExpectedOutput;
  tags: string[];
  validation?: TestCaseValidation;
}

export interface TestCaseManifest {
  dataset_name: string;
  version: string;
  owner: string;
  description: string;
  generated_at: string;
  skill_name: string;
  summary: {
    total_cases: number;
    happy_path: number;
    edge_cases: number;
    error_cases: number;
  };
  case_files: Array<{
    path: string;
    tags: string[];
  }>;
}
```

**Step 4: Create src/types/extracted.ts**

```typescript
import { Constraint, Decision, Parameter, Source } from './skill-package.js';

export interface ExtractedData {
  constraints: Constraint[];
  decisions: Decision[];
  parameters: Parameter[];
  sources: Source[];
  roles: Record<string, { description: string; mentions: string; source: string }>;
  subjective_judgments: string[];
  ambiguity_notes: string[];
}

export interface ExtractionOptions {
  extractConstraints: boolean;
  extractDecisions: boolean;
  extractRoles: boolean;
  extractBoundaries: boolean;
  confidenceThreshold: number;
}
```

**Step 5: Create src/types/index.ts**

```typescript
export * from './skill-package.js';
export * from './action.js';
export * from './test-case.js';
export * from './extracted.js';
```

**Step 6: Write tests**

Create: `sop-to-skill/src/types/types.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { ConstraintLevel, TriggerType, TestCaseType } from '../types/index.js';

describe('Type Definitions', () => {
  describe('ConstraintLevel', () => {
    it('should have MUST, SHOULD, MAY values', () => {
      expect(ConstraintLevel.MUST).toBe('MUST');
      expect(ConstraintLevel.SHOULD).toBe('SHOULD');
      expect(ConstraintLevel.MAY).toBe('MAY');
    });
  });

  describe('TriggerType', () => {
    it('should have execution, query, approval, event values', () => {
      expect(TriggerType.EXECUTION).toBe('execution');
      expect(TriggerType.QUERY).toBe('query');
      expect(TriggerType.APPROVAL).toBe('approval');
      expect(TriggerType.EVENT).toBe('event');
    });
  });

  describe('TestCaseType', () => {
    it('should have happy-path, edge-case, error-case, compliance values', () => {
      expect(TestCaseType.HAPPY_PATH).toBe('happy-path');
      expect(TestCaseType.EDGE_CASE).toBe('edge-case');
      expect(TestCaseType.ERROR_CASE).toBe('error-case');
      expect(TestCaseType.COMPLIANCE).toBe('compliance');
    });
  });
});
```

**Step 7: Run tests**

```bash
cd sop-to-skill
npm test
```

Expected: PASS

**Step 8: Commit**

```bash
git add src/types/
git commit -m "feat: add core type definitions"
```

---

### Task 0.3: 创建项目规范文件

**Files:**
- Create: `sop-to-skill/SKILL.md` (通用格式规范)
- Create: `sop-to-skill/SKILL.schema.json` (JSON Schema)

**Step 1: Create SKILL.md**

```markdown
# Skill Package Format Specification

**Version**: 1.0
**Purpose**: 定义 Skill Package 的标准格式规范

---

## 概述

Skill 是可被 AI Agent 发现、理解、执行的标准操作单元。
每个 Skill Package 包含完整的**定义**、**执行逻辑**、**测试用例**和**质量保障**。

---

## Package 结构

```
skill-package/
├── SKILL.md                    # 本文件 - Skill 定义
├── skill.schema.json           # JSON Schema 格式
├── skill.manifest.yaml         # Package 元数据
├── actions/                    # 可执行动作定义
├── constraints/                # 约束规则
├── decisions/                  # 决策表
├── test-cases/                 # 测试用例
└── examples/                   # 输入输出示例
```

---

## 核心组件

### 1. SKILL.md

人读版本，包含概述、触发条件、执行步骤、约束规则、错误处理、使用示例。

### 2. skill.schema.json

AI 可解析的 JSON Schema，包含 triggers、steps、constraints、error_handling。

### 3. actions/

每个 action 是一个独立的可执行单元，包含 input/output Schema 和 implementation。

### 4. constraints/

结构化的约束规则：MUST（强制）、SHOULD（推荐）、MAY（可选）。

### 5. test-cases/

测试用例目录：manifest.yaml + happy-path/ + edge-cases/ + error-cases/

---

## 约束词级别

| 级别 | 含义 | AI 行为 |
|------|------|---------|
| **MUST** | 强制要求，不可违反 | 必须验证，违反则失败 |
| **SHOULD** | 推荐做法 | 建议验证，可记录偏差 |
| **MAY** | 可选做法 | 酌情处理 |

---

## 格式版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2026-03-29 | 初始规范 |

---

*本规范兼容 OpenAI Function Calling 格式，支持 JSON Schema Draft-07*
```

**Step 2: Create SKILL.schema.json**

```json
{
  "$schema": "https://skill-schema.github.io/v1/schema.json",
  "$id": "https://skill-schema.github.io/v1/skill",
  "title": "Skill Package Schema",
  "description": "AI Agent Skill Package 的标准格式定义",
  "version": "1.0.0",
  "type": "object",
  "required": ["meta", "triggers", "steps"],
  "properties": {
    "meta": {
      "type": "object",
      "description": "Skill 元数据",
      "required": ["name", "version"],
      "properties": {
        "name": { "type": "string", "description": "Skill 名称" },
        "version": { "type": "string", "pattern": "^\\d+\\.\\d+\\.\\d+$" },
        "description": { "type": "string" },
        "source": { "type": "string", "description": "来源 SOP/文档" },
        "tags": { "type": "array", "items": { "type": "string" } },
        "generated_at": { "type": "string", "format": "date-time" }
      }
    },
    "triggers": {
      "type": "array",
      "description": "触发条件列表",
      "items": {
        "type": "object",
        "required": ["type", "input_schema"],
        "properties": {
          "type": {
            "type": "string",
            "enum": ["execution", "query", "approval", "event"]
          },
          "name": { "type": "string" },
          "description": { "type": "string" },
          "input_schema": { "$ref": "#/definitions/JSONSchema" },
          "output_schema": { "$ref": "#/definitions/JSONSchema" }
        }
      }
    },
    "steps": {
      "type": "array",
      "description": "执行步骤列表",
      "items": {
        "type": "object",
        "required": ["id", "name", "action"],
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "description": { "type": "string" },
          "action": { "type": "string" },
          "condition": { "type": "string" },
          "input": { "type": "object" },
          "output": { "type": "object" },
          "next_step_on_success": { "type": "string" },
          "next_step_on_failure": { "type": "string" },
          "on_failure": { "type": "string" }
        }
      }
    },
    "constraints": {
      "type": "array",
      "description": "约束规则列表",
      "items": {
        "type": "object",
        "required": ["id", "level", "description"],
        "properties": {
          "id": { "type": "string" },
          "level": { "type": "string", "enum": ["MUST", "SHOULD", "MAY"] },
          "description": { "type": "string" },
          "condition": { "type": "string" },
          "action": { "type": "string" },
          "validation": { "$ref": "#/definitions/ValidationRule" },
          "roles": { "type": "array", "items": { "type": "string" } },
          "confidence": { "type": "number", "minimum": 0, "maximum": 1 },
          "source_quote": { "type": "string" }
        }
      }
    },
    "decisions": {
      "type": "array",
      "description": "决策表列表",
      "items": {
        "type": "object",
        "required": ["id", "name", "rules"],
        "properties": {
          "id": { "type": "string" },
          "name": { "type": "string" },
          "inputVars": { "type": "array", "items": { "type": "string" } },
          "outputVars": { "type": "array", "items": { "type": "string" } },
          "rules": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "when": { "type": "object" },
                "then": { "type": "object" }
              }
            }
          }
        }
      }
    },
    "parameters": {
      "type": "array",
      "description": "参数定义列表",
      "items": {
        "type": "object",
        "required": ["name", "type"],
        "properties": {
          "name": { "type": "string" },
          "type": { "type": "string", "enum": ["string", "number", "boolean", "enum"] },
          "description": { "type": "string" },
          "minValue": { "type": "number" },
          "maxValue": { "type": "number" },
          "defaultValue": {},
          "unit": { "type": "string" }
        }
      }
    },
    "error_handling": {
      "type": "object",
      "description": "错误处理规则",
      "properties": {
        "rules": {
          "type": "array",
          "items": {
            "type": "object",
            "required": ["condition", "actions"],
            "properties": {
              "condition": { "type": "string" },
              "actions": {
                "type": "array",
                "items": {
                  "type": "object",
                  "properties": {
                    "type": { "type": "string" },
                    "description": { "type": "string" },
                    "roles": { "type": "array", "items": { "type": "string" } }
                  }
                }
              }
            }
          }
        }
      }
    },
    "sources": {
      "type": "array",
      "description": "来源文档列表",
      "items": {
        "type": "object",
        "properties": {
          "type": { "type": "string", "enum": ["sop", "policy", "guideline"] },
          "fileName": { "type": "string" },
          "section": { "type": "string" },
          "url": { "type": "string", "format": "uri" }
        }
      }
    }
  },
  "definitions": {
    "JSONSchema": {
      "type": "object",
      "properties": {
        "type": { "type": "string" },
        "properties": { "type": "object" },
        "required": { "type": "array", "items": { "type": "string" } },
        "enum": { "type": "array" },
        "description": { "type": "string" },
        "items": {},
        "format": { "type": "string" }
      }
    },
    "ValidationRule": {
      "type": "object",
      "required": ["type", "error_code"],
      "properties": {
        "type": {
          "type": "string",
          "enum": ["assert", "role_check", "deadline", "threshold"]
        },
        "condition": { "type": "string" },
        "error_code": { "type": "string" },
        "error_message": { "type": "string" },
        "normal": {
          "type": "object",
          "properties": {
            "value": { "type": "number" },
            "unit": { "type": "string" }
          }
        },
        "urgent": {
          "type": "object",
          "properties": {
            "value": { "type": "number" },
            "unit": { "type": "string" }
          }
        },
        "urgent_condition": { "type": "string" },
        "metric": { "type": "string" },
        "threshold": { "type": "number" },
        "action": { "type": "string" }
      }
    }
  }
}
```

**Step 3: Commit**

```bash
git add SKILL.md SKILL.schema.json
git commit -m "docs: add skill format specification and JSON schema"
```

---

## Phase 1: CLI 框架

### Task 1.1: 创建 CLI 入口和基础命令

**Files:**
- Create: `sop-to-skill/src/cli.ts`

**Step 1: Create src/cli.ts**

```typescript
import { Cli, Usage } from 'clipanion';
import GenerateCommand from './commands/generate.js';
import ExtractCommand from './commands/extract.js';
import LLMEnhanceCommand from './commands/llm-enhance.js';
import ValidateCommand from './commands/validate.js';
import VersionCommand from './commands/version.js';

const cli = Cli.create({
  binaryLabel: 'sop-to-skill',
  binaryName: 'sop-to-skill',
  binaryVersion: '1.0.0',
  enableHelpCommand: true,
});

cli.register(GenerateCommand);
cli.register(ExtractCommand);
cli.register(LLMEnhanceCommand);
cli.register(ValidateCommand);
cli.register(VersionCommand);

export default cli;
```

**Step 2: Create src/commands/generate.ts**

```typescript
import { Command, Option, UsageError } from 'clipanion';
import path from 'path';
import fs from 'fs/promises';
import { generateSkillPackage } from '../generator/skill-package.js';
import { parseInputFile } from '../parser/factory.js';
import { extractFromText } from '../extractor/index.js';
import { enhanceWithLLM } from '../llm/enhancer.js';
import type { ExtractionOptions, LLMConfig } from '../types/index.js';

export default class GenerateCommand extends Command {
  static paths = [['generate']];
  static usage: Usage = {
    description: 'Generate Skill Package from SOP document',
    details: `
      Converts a SOP document (Markdown, PDF, DOCX) into a complete Skill Package.

      Examples:
        sop-to-skill generate ./SOP-DM-002.md --name "数据核查流程" --output ./output
        sop-to-skill generate ./SOP.md --name "订单处理" --output ./out --llm
    `,
    examples: [
      ['Generate from Markdown', 'sop-to-skill generate ./SOP.md --name "MySkill" --output ./out'],
      ['With LLM enhancement', 'sop-to-skill generate ./SOP.md --name "MySkill" --output ./out --llm'],
    ],
  };

  inputFile = Option.String({ required: true, description: 'Input SOP file path' });
  name = Option.String({ required: true, description: 'Skill name' });
  output = Option.String({ required: true, description: 'Output directory path' });
  framework = Option.String('--framework', { description: 'Output framework (openclaw,gpts,mcp,claude,langchain,all)' });
  llm = Option.Boolean('--llm', { description: 'Enable LLM enhancement' });
  llmApi = Option.String('--llm-api', { description: 'LLM API URL' });
  llmModel = Option.String('--llm-model', { description: 'LLM model name' });

  async execute() {
    const context = this.context;

    context.stdout.write(`Parsing ${this.inputFile}...\n`);

    // Parse input file
    const parsed = await parseInputFile(this.inputFile);

    context.stdout.write(`Extracted ${parsed.constraints.length} constraints, ${parsed.decisions.length} decisions\n`);

    // LLM enhancement if requested
    let enhanced = parsed;
    if (this.llm) {
      context.stdout.write('Enhancing with LLM...\n');
      const llmConfig: LLMConfig = {
        apiUrl: this.llmApi || process.env.LLM_API_URL || 'http://localhost:11434',
        model: this.llmModel || process.env.LLM_MODEL || 'gpt-4o',
        apiKey: process.env.LLM_API_KEY,
      };
      enhanced = await enhanceWithLLM(parsed, llmConfig);
    }

    // Generate Skill Package
    context.stdout.write(`Generating Skill Package '${this.name}'...\n`);
    const skillPackage = await generateSkillPackage(enhanced, this.name, {
      sourceFile: this.inputFile,
      framework: this.framework,
    });

    // Write to output directory
    await fs.mkdir(this.output, { recursive: true });
    await writeSkillPackage(this.output, skillPackage);

    context.stdout.write(`Successfully generated Skill Package at ${this.output}\n`);
  }
}
```

**Step 3: Create src/commands/extract.ts**

```typescript
import { Command, Option } from 'clipanion';
import fs from 'fs/promises';
import { parseInputFile } from '../parser/factory.js';
import { extractFromText } from '../extractor/index.js';

export default class ExtractCommand extends Command {
  static paths = [['extract']];
  static usage = {
    description: 'Extract structured data from SOP document',
    details: 'Extract constraints, decisions, roles without generating Skill Package.',
  };

  inputFile = Option.String({ required: true });
  output = Option.String('--output', { required: true, description: 'Output JSON file' });
  format = Option.String('--format', { description: 'Output format (json|yaml)', default: 'json' });

  async execute() {
    this.context.stdout.write(`Extracting from ${this.inputFile}...\n`);

    const parsed = await parseInputFile(this.inputFile);
    const extracted = extractFromText(parsed.content);

    const outputPath = this.output;
    await fs.mkdir(path.dirname(outputPath), { recursive: true });

    if (this.format === 'yaml') {
      const yaml = await import('js-yaml');
      await fs.writeFile(outputPath, yaml.dump(extracted));
    } else {
      await fs.writeFile(outputPath, JSON.stringify(extracted, null, 2));
    }

    this.context.stdout.write(`Extracted to ${outputPath}\n`);
  }
}
```

**Step 4: Create src/commands/llm-enhance.ts**

```typescript
import { Command, Option } from 'clipanion';
import fs from 'fs/promises';
import { enhanceWithLLM } from '../llm/enhancer.js';
import type { LLMConfig, ExtractedData } from '../types/index.js';

export default class LLMEnhanceCommand extends Command {
  static paths = [['llm-enhance']];
  static usage = {
    description: 'Enhance extracted data with LLM',
    details: 'Use LLM to improve semantic understanding and generate test cases.',
  };

  inputFile = Option.String({ required: true, description: 'Input JSON file from extract' });
  output = Option.String('--output', { required: true, description: 'Output file' });
  llmApi = Option.String('--llm-api', { description: 'LLM API URL' });
  llmModel = Option.String('--llm-model', { description: 'LLM model name' });

  async execute() {
    this.context.stdout.write(`Reading ${this.inputFile}...\n`);

    const content = await fs.readFile(this.inputFile, 'utf-8');
    const extracted: ExtractedData = JSON.parse(content);

    const llmConfig: LLMConfig = {
      apiUrl: this.llmApi || process.env.LLM_API_URL || 'http://localhost:11434',
      model: this.llmModel || process.env.LLM_MODEL || 'gpt-4o',
      apiKey: process.env.LLM_API_KEY,
    };

    this.context.stdout.write('Enhancing with LLM...\n');
    const enhanced = await enhanceWithLLM(extracted, llmConfig);

    await fs.writeFile(this.output, JSON.stringify(enhanced, null, 2));
    this.context.stdout.write(`Enhanced data saved to ${this.output}\n`);
  }
}
```

**Step 5: Create src/commands/validate.ts**

```typescript
import { Command, Option } from 'clipanion';
import fs from 'fs/promises';
import path from 'path';
import { validateSkillPackage } from '../validator/index.js';

export default class ValidateCommand extends Command {
  static paths = [['validate']];
  static usage = {
    description: 'Validate Skill Package structure',
    details: 'Check if a Skill Package has all required files and valid content.',
  };

  packagePath = Option.String({ required: true, description: 'Skill Package directory or file' });

  async execute() {
    this.context.stdout.write(`Validating ${this.packagePath}...\n`);

    const result = await validateSkillPackage(this.packagePath);

    if (result.valid) {
      this.context.stdout.write(`✓ Validation passed (score: ${result.score.toFixed(0%)})\n`);
      process.exit(0);
    } else {
      this.context.stdout.write(`✗ Validation failed:\n`);
      for (const issue of result.issues) {
        this.context.stdout.write(`  [${issue.severity}] ${issue.message}\n`);
      }
      process.exit(1);
    }
  }
}
```

**Step 6: Create src/commands/version.ts**

```typescript
import { Command } from 'clipanion';

export default class VersionCommand extends Command {
  static paths = [['version'], ['--version'], ['-v']];
  static usage = { description: 'Show version' };

  async execute() {
    this.context.stdout.write('sop-to-skill v1.0.0\n');
  }
}
```

**Step 7: Create src/index.ts (library entry)**

```typescript
export * from './types/index.js';
export { parseInputFile } from './parser/factory.js';
export { extractFromText } from './extractor/index.js';
export { enhanceWithLLM } from './llm/enhancer.js';
export { generateSkillPackage } from './generator/skill-package.js';
export { validateSkillPackage } from './validator/index.js';
```

**Step 8: Create src/main.ts (bin entry)**

```typescript
import cli from './cli.js';
import { run } from 'clipanion';

run(cli, process.argv, process.env);
```

**Step 9: Update package.json bin**

```bash
npm pkg set bin.sop-to-skill="./dist/main.js"
```

**Step 10: Commit**

```bash
git add src/cli.ts src/commands/ src/index.ts src/main.ts
git commit -m "feat: add CLI framework with generate/extract/llm-enhance/validate commands"
```

---

## Phase 2: Parser 模块

### Task 2.1: 创建 Parser 工厂和 Markdown Parser

**Files:**
- Create: `sop-to-skill/src/parser/factory.ts`
- Create: `sop-to-skill/src/parser/markdown.ts`
- Create: `sop-to-skill/src/parser/types.ts`

**Step 1: Create src/parser/types.ts**

```typescript
export interface ParsedContent {
  content: string;
  metadata: {
    title?: string;
    version?: string;
    date?: string;
    fileType: 'markdown' | 'pdf' | 'docx' | 'unknown';
  };
  sections: ParsedSection[];
}

export interface ParsedSection {
  title: string;
  level: number;
  content: string;
  children: ParsedSection[];
}
```

**Step 2: Create src/parser/markdown.ts**

```typescript
import { marked } from 'marked';
import type { ParsedContent, ParsedSection } from './types.js';

export interface MarkdownParseOptions {
  extractTables?: boolean;
  extractLists?: boolean;
}

export function parseMarkdown(content: string, options?: MarkdownParseOptions): ParsedContent {
  // Extract metadata from frontmatter
  const frontmatterMatch = content.match(/^---\n([\s\S]*?)\n---\n/);
  let metadata: ParsedContent['metadata'] = { fileType: 'markdown' };
  let mainContent = content;

  if (frontmatterMatch) {
    const frontmatter = frontmatterMatch[1];
    mainContent = content.slice(frontmatterMatch[0].length);
    
    // Parse frontmatter
    const yaml = require('js-yaml');
    try {
      const frontmatterData = yaml.load(frontmatter);
      metadata = { ...metadata, ...frontmatterData };
    } catch {}
  }

  // Extract title from first H1
  const titleMatch = mainContent.match(/^#\s+(.+)$/m);
  if (titleMatch) {
    metadata.title = titleMatch[1];
  }

  // Parse sections using marked
  const tokens = marked.lexer(mainContent);
  const sections = parseTokens(tokens);

  return {
    content: mainContent,
    metadata,
    sections,
  };
}

function parseTokens(tokens: marked.Token[]): ParsedSection[] {
  const sections: ParsedSection[] = [];
  let currentSection: ParsedSection | null = null;
  const sectionStack: ParsedSection[] = [];

  for (const token of tokens) {
    if (token.type === 'heading') {
      const level = token.depth;
      const section: ParsedSection = {
        title: token.text,
        level,
        content: '',
        children: [],
      };

      // Pop sections of same or higher level
      while (sectionStack.length > 0 && sectionStack[sectionStack.length - 1].level >= level) {
        sectionStack.pop();
      }

      if (sectionStack.length === 0) {
        sections.push(section);
      } else {
        sectionStack[sectionStack.length - 1].children.push(section);
      }

      sectionStack.push(section);
      currentSection = section;
    } else if (currentSection) {
      currentSection.content += '\n' + token.text;
    }
  }

  return sections;
}
```

**Step 3: Create src/parser/factory.ts**

```typescript
import fs from 'fs/promises';
import path from 'path';
import { parseMarkdown } from './markdown.js';
import type { ParsedContent } from './types.js';

export async function parseInputFile(filePath: string): Promise<ParsedContent> {
  const ext = path.extname(filePath).toLowerCase();
  const content = await fs.readFile(filePath, 'utf-8');

  switch (ext) {
    case '.md':
    case '.markdown':
      return parseMarkdown(content);

    case '.pdf':
      // Lazy load pdf-parse
      const pdfParse = (await import('pdf-parse')).default;
      const pdfData = await fs.readFile(filePath);
      const pdfResult = await pdfParse(pdfData);
      return parseMarkdown(pdfResult.text);

    case '.docx':
      // Lazy load mammoth
      const mammoth = await import('mammoth');
      const docxResult = await mammoth.extractRawText({ path: filePath });
      return parseMarkdown(docxResult.value);

    default:
      // Try as plain text
      return parseMarkdown(content);
  }
}
```

**Step 4: Write tests**

Create: `sop-to-skill/src/parser/markdown.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { parseMarkdown } from './markdown.js';

describe('Markdown Parser', () => {
  it('should parse basic markdown', () => {
    const content = `# Title

## Section 1

Some content.

## Section 2

More content.
`;
    const result = parseMarkdown(content);
    expect(result.metadata.title).toBe('Title');
    expect(result.sections).toHaveLength(2);
    expect(result.sections[0].title).toBe('Section 1');
    expect(result.sections[1].title).toBe('Section 2');
  });

  it('should parse tables', () => {
    const content = `# Test

| Column 1 | Column 2 |
|----------|----------|
| Value 1 | Value 2 |
`;
    const result = parseMarkdown(content);
    expect(result.sections[0].content).toContain('Column 1');
  });
});
```

**Step 5: Commit**

```bash
git add src/parser/
git commit -m "feat: add parser module with Markdown support"
```

---

## Phase 3: Extractor 模块

### Task 3.1: 创建基础 Extractor

**Files:**
- Create: `sop-to-skill/src/extractor/index.ts`
- Create: `sop-to-skill/src/extractor/constraint.ts`
- Create: `sop-to-skill/src/extractor/decision.ts`
- Create: `sop-to-skill/src/extractor/role.ts`

**Step 1: Create src/extractor/constraint.ts**

```typescript
import type { Constraint, ConstraintLevel } from '../types/index.js';

const CHINESE_MUST_PATTERNS = [
  /必须|应当|需要|不得|严禁|不准|不得超过|不得少于/,
];

const CHINESE_SHOULD_PATTERNS = [
  /建议|推荐|应该|宜|可以考虑|一般应该|通常应当/,
];

const CHINESE_MAY_PATTERNS = [
  /可以|允许|酌情|根据情况|视情况/,
];

const IF_THEN_PATTERNS = [
  /如果(.+?)，则(.+?)$/,
  /若(.+?)，则(.+?)$/,
  /当(.+?)时(.+?)$/,
];

export interface ConstraintExtractorOptions {
  confidenceThreshold: number;
}

export function extractConstraints(
  text: string,
  options?: ConstraintExtractorOptions
): Constraint[] {
  const constraints: Constraint[] = [];
  const sentences = text.split(/[。\n]/);
  
  let counter = 1;

  for (const sentence of sentences) {
    if (sentence.length < 5) continue;

    // Try if-then pattern
    const ifThenMatch = sentence.match(/如果(.+?)，则(.+?)$/);
    if (ifThenMatch) {
      const condition = ifThenMatch[1].trim();
      const action = ifThenMatch[2].trim();
      const level = detectLevel(action);
      
      constraints.push({
        id: `C${counter.toString().padStart(3, '0')}`,
        level,
        description: `如果 ${condition}，则 ${action}`,
        condition,
        action,
        roles: extractRoles(sentence),
        confidence: 0.9,
        source_quote: sentence,
      });
      counter++;
      continue;
    }

    // Try standalone constraint keyword
    const level = detectLevel(sentence);
    if (level) {
      constraints.push({
        id: `C${counter.toString().padStart(3, '0')}`,
        level,
        description: sentence.trim(),
        roles: extractRoles(sentence),
        confidence: 0.85,
        source_quote: sentence,
      });
      counter++;
    }
  }

  return constraints;
}

function detectLevel(text: string): ConstraintLevel | null {
  if (CHINESE_MUST_PATTERNS.some(p => p.test(text))) return 'MUST';
  if (CHINESE_SHOULD_PATTERNS.some(p => p.test(text))) return 'SHOULD';
  if (CHINESE_MAY_PATTERNS.some(p => p.test(text))) return 'MAY';
  return null;
}

const ROLE_PATTERNS = [
  /数据录入员|DMC|QA|DM|数据经理|项目经理|医学监查员|统计师/,
  /质量保证|伦理委员会|总经理|经理|总监|主管|专员/,
  /销售员|采购员|财务|系统管理员|管理员|操作员|审批人/,
];

function extractRoles(text: string): string[] {
  const roles: string[] = [];
  for (const pattern of ROLE_PATTERNS) {
    const match = text.match(pattern);
    if (match) {
      roles.push(match[0]);
    }
  }
  return roles;
}
```

**Step 2: Create src/extractor/decision.ts**

```typescript
import type { Decision, DecisionRule } from '../types/index.js';

export function extractDecisions(text: string): Decision[] {
  const decisions: Decision[] = [];
  
  // Look for decision tables
  const tablePattern = /\|(.+?)\|(.+?)\|(.+?)\|/g;
  const tables = text.matchAll(tablePattern);
  
  // For now, extract simple if-then rules as decisions
  const ifThenPattern = /如果(.+?)，则(.+?)[。\n]/g;
  const matches = text.matchAll(ifThenPattern);
  
  const rules: DecisionRule[] = [];
  for (const match of matches) {
    const whenPart = match[1];
    const thenPart = match[2];
    
    const when: Record<string, string> = {};
    const then: Record<string, string> = {};
    
    // Simple key=value parsing
    const whenPairs = whenPart.split(/[，,]/);
    for (const pair of whenPairs) {
      const [k, v] = pair.split(/[=:]/).map(s => s.trim());
      if (k && v) when[k] = v;
    }
    
    const thenPairs = thenPart.split(/[，,]/);
    for (const pair of thenPairs) {
      const [k, v] = pair.split(/[=:]/).map(s => s.trim());
      if (k && v) then[k] = v;
    }
    
    if (Object.keys(when).length > 0 && Object.keys(then).length > 0) {
      rules.push({ when, then });
    }
  }
  
  if (rules.length > 0) {
    decisions.push({
      id: 'decision_001',
      name: 'If-Then Decision Rules',
      inputVars: ['condition'],
      outputVars: ['action'],
      rules,
    });
  }
  
  return decisions;
}
```

**Step 3: Create src/extractor/role.ts**

```typescript
export interface Role {
  name: string;
  description: string;
  mentions: number;
  source: string;
}

const ROLE_DEFINITIONS: Record<string, { description: string; patterns: RegExp[] }> = {
  'DMC': {
    description: '数据监查员',
    patterns: [/数据监查员|DMC|Data Monitor/],
  },
  '数据经理': {
    description: '数据管理经理',
    patterns: [/数据经理|Data Manager/],
  },
  'QA': {
    description: '质量保证人员',
    patterns: [/质量保证|QA|Quality Assurance/],
  },
  '医学监查员': {
    description: '医学监查员',
    patterns: [/医学监查员|MA|Medical Associate/],
  },
};

export function extractRoles(text: string): Record<string, Role> {
  const roles: Record<string, Role> = {};
  
  for (const [name, def] of Object.entries(ROLE_DEFINITIONS)) {
    for (const pattern of def.patterns) {
      if (pattern.test(text)) {
        if (!roles[name]) {
          roles[name] = {
            name,
            description: def.description,
            mentions: 0,
            source: 'sop',
          };
        }
        roles[name].mentions += (text.match(pattern) || []).length;
      }
    }
  }
  
  return roles;
}
```

**Step 4: Create src/extractor/index.ts**

```typescript
import { parseMarkdown } from '../parser/markdown.js';
import type { ParsedContent } from '../parser/types.js';
import type { ExtractedData } from '../types/index.js';
import { extractConstraints } from './constraint.js';
import { extractDecisions } from './decision.js';
import { extractRoles } from './role.js';

export { extractConstraints } from './constraint.js';
export { extractDecisions } from './decision.js';
export { extractRoles } from './role.js';

export function extractFromText(content: string): ExtractedData {
  const parsed = parseMarkdown(content);
  
  const constraints = extractConstraints(parsed.content);
  const decisions = extractDecisions(parsed.content);
  const roles = extractRoles(parsed.content);
  
  return {
    constraints,
    decisions,
    parameters: [],
    sources: [{
      type: 'sop',
      fileName: parsed.metadata.title || 'unknown',
    }],
    roles,
    subjective_judgments: [],
    ambiguity_notes: [],
  };
}

export function extractFromParsed(parsed: ParsedContent): ExtractedData {
  const constraints = extractConstraints(parsed.content);
  const decisions = extractDecisions(parsed.content);
  const roles = extractRoles(parsed.content);
  
  return {
    constraints,
    decisions,
    parameters: [],
    sources: [{
      type: 'sop',
      fileName: parsed.metadata.title || 'unknown',
      section: parsed.sections[0]?.title,
    }],
    roles,
    subjective_judgments: [],
    ambiguity_notes: [],
  };
}
```

**Step 5: Write tests**

Create: `sop-to-skill/src/extractor/extractor.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { extractConstraints } from './constraint.js';
import { extractFromText } from './index.js';

describe('Extractor', () => {
  describe('extractConstraints', () => {
    it('should extract MUST constraints', () => {
      const text = '必须核查主要疗效终点、安全性数据。';
      const result = extractConstraints(text);
      expect(result).toHaveLength(1);
      expect(result[0].level).toBe('MUST');
    });

    it('should extract SHOULD constraints', () => {
      const text = '应当每周汇总质疑统计报告。';
      const result = extractConstraints(text);
      expect(result).toHaveLength(1);
      expect(result[0].level).toBe('SHOULD');
    });

    it('should extract if-then patterns', () => {
      const text = '如果 SDV 发现错误率超过 5%，则必须扩大核查范围，并向数据经理报告。';
      const result = extractConstraints(text);
      expect(result).toHaveLength(1);
      expect(result[0].condition).toBe('SDV 发现错误率超过 5%');
      expect(result[0].action).toContain('必须扩大核查范围');
    });
  });

  describe('extractFromText', () => {
    it('should extract all components', () => {
      const text = `# SOP Title

必须核查主要疗效终点。

如果 SDV 发现错误率超过 5%，则必须扩大核查范围。
`;
      const result = extractFromText(text);
      expect(result.constraints.length).toBeGreaterThan(0);
    });
  });
});
```

**Step 6: Commit**

```bash
git add src/extractor/
git commit -m "feat: add extractor module for constraint/decision/role extraction"
```

---

## Phase 4: LLM 增强模块

### Task 4.1: 创建 LLM 增强器

**Files:**
- Create: `sop-to-skill/src/llm/enhancer.ts`
- Create: `sop-to-skill/src/llm/prompts.ts`
- Create: `sop-to-skill/src/llm/types.ts`

**Step 1: Create src/llm/types.ts**

```typescript
export interface LLMConfig {
  apiUrl: string;
  model: string;
  apiKey?: string;
}

export interface LLMResponse {
  content: string;
  usage?: {
    prompt_tokens: number;
    completion_tokens: number;
    total_tokens: number;
  };
}

export interface SemanticEnhancement {
  original_constraint: string;
  understood_meaning: string;
  condition?: string;
  action?: string;
  roles: string[];
  edge_cases: string[];
}

export interface EnhancementResult {
  semantic_constraints: SemanticEnhancement[];
  improved_descriptions: Record<string, string>;
  identified_gaps: string[];
  suggested_test_cases: Array<{
    description: string;
    type: 'happy-path' | 'edge-case' | 'error-case';
    input: Record<string, any>;
    expected: Record<string, any>;
  }>;
}
```

**Step 2: Create src/llm/prompts.ts**

```typescript
export const EXTRACT_SEMANTIC_PROMPT = `你是一个企业流程专家，擅长从SOP文档中提取精确的操作规范。

分析以下约束文本，理解其精确含义，并提取结构化信息。

约束文本：
{constraint_text}

请以JSON格式输出：
{
  "understood_meaning": "你理解的精确含义",
  "condition": "触发条件（如果有）",
  "action": "执行动作",
  "roles": ["相关角色列表"],
  "edge_cases": ["可能的边缘情况"]
}`;

export const GENERATE_TEST_CASES_PROMPT = `你是一个测试工程师，擅长从SOP文档生成测试用例。

根据以下Skill定义，生成多样化的测试用例。

Skill名称：{skill_name}
约束规则：
{constraints}

请生成{count}个测试用例，包括：
- happy-path（正常流程）
- edge-case（边界情况）
- error-case（错误情况）

以JSON数组格式输出：
[{
  "case_id": "hp-001",
  "type": "happy-path",
  "description": "测试描述",
  "input": { "key": "value" },
  "expected": { "result": "expected" },
  "tags": ["tag1"]
}]`;

export const IMPROVE_SKILL_PROMPT = `你是一个Skill优化专家。

分析以下Skill定义的不足，并提出改进建议。

Skill定义：
{skill_definition}

检查方面：
1. 完整性 - 是否有遗漏的场景？
2. 清晰性 - 描述是否清晰无歧义？
3. 可执行性 - AI能否准确执行？

以JSON格式输出改进建议：
{
  "gaps": ["不足1", "不足2"],
  "suggestions": ["建议1", "建议2"],
  "clarifications": ["歧义1的澄清"]
}`;
```

**Step 3: Create src/llm/enhancer.ts**

```typescript
import type { ExtractedData, Constraint } from '../types/index.js';
import type { LLMConfig, EnhancementResult } from './types.js';
import { EXTRACT_SEMANTIC_PROMPT, GENERATE_TEST_CASES_PROMPT, IMPROVE_SKILL_PROMPT } from './prompts.js';

export { type LLMConfig, type EnhancementResult } from './types.js';

export async function enhanceWithLLM(
  extracted: ExtractedData,
  config: LLMConfig
): Promise<ExtractedData & { llm_enhanced: true; enhancement_result?: EnhancementResult }> {
  // For now, use OpenAI-compatible API
  const response = await fetch(config.apiUrl + '/v1/chat/completions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(config.apiKey ? { 'Authorization': `Bearer ${config.apiKey}` } : {}),
    },
    body: JSON.stringify({
      model: config.model,
      messages: [
        {
          role: 'system',
          content: 'You are a helpful assistant that outputs valid JSON only.',
        },
        {
          role: 'user',
          content: IMPROVE_SKILL_PROMPT.replace(
            '{skill_definition}',
            JSON.stringify(extracted, null, 2)
          ),
        },
      ],
      temperature: 0.7,
    }),
  });

  if (!response.ok) {
    throw new Error(`LLM API error: ${response.statusText}`);
  }

  const data = await response.json();
  const content = data.choices[0]?.message?.content;

  let enhancementResult: EnhancementResult | undefined;
  try {
    enhancementResult = JSON.parse(content);
  } catch {
    // If parsing fails, continue without enhancement
  }

  // Merge improved constraints back
  const improvedConstraints = enhanceConstraints(extracted.constraints, enhancementResult);

  return {
    ...extracted,
    constraints: improvedConstraints,
    llm_enhanced: true,
    enhancement_result: enhancementResult,
  };
}

function enhanceConstraints(
  constraints: Constraint[],
  result?: EnhancementResult
): Constraint[] {
  if (!result) return constraints;

  const improvedMap = new Map(
    result.semantic_constraints.map(sc => [sc.original_constraint, sc])
  );

  return constraints.map(constraint => {
    const improvement = improvedMap.get(constraint.description);
    if (improvement) {
      return {
        ...constraint,
        description: improvement.understood_meaning,
        condition: improvement.condition || constraint.condition,
        action: improvement.action || constraint.action,
        confidence: 0.95, // Higher confidence with LLM enhancement
      };
    }
    return constraint;
  });
}
```

**Step 4: Create src/llm/test-case-generator.ts**

```typescript
import type { LLMConfig } from './types.js';
import type { TestCase, Constraint } from '../types/index.js';
import { GENERATE_TEST_CASES_PROMPT } from './prompts.js';

export interface TestCaseGeneratorOptions {
  minCases: number;
  maxCases: number;
  types: Array<'happy-path' | 'edge-case' | 'error-case' | 'compliance'>;
}

export async function generateTestCases(
  constraints: Constraint[],
  skillName: string,
  config: LLMConfig,
  options: Partial<TestCaseGeneratorOptions> = {}
): Promise<TestCase[]> {
  const { minCases = 10, maxCases = 20, types = ['happy-path', 'edge-case', 'error-case'] } = options;

  const constraintsText = constraints
    .map(c => `[${c.level}] ${c.description}`)
    .join('\n');

  const prompt = GENERATE_TEST_CASES_PROMPT
    .replace('{skill_name}', skillName)
    .replace('{constraints}', constraintsText)
    .replace('{count}', maxCases.toString());

  const response = await fetch(config.apiUrl + '/v1/chat/completions', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(config.apiKey ? { 'Authorization': `Bearer ${config.apiKey}` } : {}),
    },
    body: JSON.stringify({
      model: config.model,
      messages: [
        { role: 'system', content: 'You are a test engineer. Output valid JSON array only.' },
        { role: 'user', content: prompt },
      ],
      temperature: 0.7,
    }),
  });

  const data = await response.json();
  const content = data.choices[0]?.message?.content;

  try {
    const cases = JSON.parse(content);
    return cases.map((c: any, i: number) => ({
      ...c,
      case_id: c.case_id || `auto-${i.toString().padStart(3, '0')}`,
      skill: skillName,
    }));
  } catch {
    return [];
  }
}
```

**Step 5: Write tests**

Create: `sop-to-skill/src/llm/enhancer.test.ts`

```typescript
import { describe, it, expect, vi } from 'vitest';
import type { ExtractedData } from '../types/index.js';

describe('LLM Enhancer', () => {
  it('should export enhanceWithLLM function', async () => {
    // Mock fetch
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      json: () => Promise.resolve({
        choices: [{ message: { content: '{"gaps":[], "suggestions":[]}' } }],
      }),
    });

    const { enhanceWithLLM } = await import('./enhancer.js');
    const extracted: ExtractedData = {
      constraints: [{
        id: 'C001',
        level: 'MUST',
        description: '必须核查主要疗效终点',
        roles: ['DMC'],
        confidence: 0.85,
      }],
      decisions: [],
      parameters: [],
      sources: [],
      roles: {},
      subjective_judgments: [],
      ambiguity_notes: [],
    };

    const result = await enhanceWithLLM(extracted, {
      apiUrl: 'http://localhost:11434',
      model: 'gpt-4o',
    });

    expect(result.llm_enhanced).toBe(true);
  });
});
```

**Step 6: Commit**

```bash
git add src/llm/
git commit -m "feat: add LLM enhancement module"
```

---

## Phase 5: Generator 模块

### Task 5.1: 创建 Skill Package Generator

**Files:**
- Create: `sop-to-skill/src/generator/skill-package.ts`
- Create: `sop-to-skill/src/generator/skill-md.ts`
- Create: `sop-to-skill/src/generator/manifest.ts`
- Create: `sop-to-skill/src/generator/test-cases.ts`

**Step 1: Create src/generator/skill-package.ts**

```typescript
import type { ExtractedData, SkillSchema, Manifest, SkillPackage as ISkillPackage } from '../types/index.js';
import { generateSkillMarkdown } from './skill-md.js';
import { generateManifest } from './manifest.js';
import { generateTestCases } from './test-cases.js';

export interface GeneratorOptions {
  sourceFile: string;
  framework?: string;
  llmModel?: string;
}

export async function generateSkillPackage(
  extracted: ExtractedData,
  name: string,
  options: GeneratorOptions
): Promise<ISkillPackage> {
  const version = '1.0.0';
  const now = new Date().toISOString();

  // Build schema
  const schema: SkillSchema = {
    meta: {
      name,
      version,
      description: `Generated from ${options.sourceFile}`,
      source: options.sourceFile,
      generated_at: now,
    },
    triggers: buildTriggers(extracted),
    steps: buildSteps(extracted),
    constraints: extracted.constraints,
    decisions: extracted.decisions,
    error_handling: buildErrorHandling(extracted),
  };

  // Build manifest
  const manifest: Manifest = {
    package: {
      name: name.toLowerCase().replace(/\s+/g, '-'),
      version,
      type: 'skill',
    },
    source: {
      type: 'sop',
      file: options.sourceFile,
      parsed_at: now,
    },
    generation: {
      method: 'sop-to-skill v1.0',
      llm_enhanced: 'llm_enhanced' in extracted ? extracted.llm_enhanced : false,
      llm_model: options.llmModel,
    },
    files: [
      { path: 'SKILL.md', type: 'documentation' },
      { path: 'skill.schema.json', type: 'schema' },
      { path: 'skill.manifest.yaml', type: 'manifest' },
      { path: 'test-cases/', type: 'test_cases' },
    ],
    quality: {
      constraint_confidence_avg: calculateAvgConfidence(extracted.constraints),
    },
  };

  return { schema, manifest };
}

function buildTriggers(extracted: ExtractedData) {
  return [
    {
      type: 'execution',
      name: 'execute',
      input_schema: {
        type: 'object',
        properties: {
          action: { type: 'string', enum: ['execute', 'query', 'approve'] },
        },
        required: ['action'],
      },
    },
  ];
}

function buildSteps(extracted: ExtractedData) {
  // Convert constraints to steps
  const mustConstraints = extracted.constraints.filter(c => c.level === 'MUST');
  
  return mustConstraints.slice(0, 5).map((c, i) => ({
    id: `step_${i + 1}`,
    name: c.condition || `执行约束 ${c.id}`,
    description: c.description,
    action: `validate_${c.id}`,
    input: {},
    output: { valid: 'boolean', violations: [] },
    on_failure: i > 0 ? 'abort' : undefined,
  }));
}

function buildErrorHandling(extracted: ExtractedData) {
  const errorRules: any[] = [];
  
  for (const c of extracted.constraints) {
    if (c.condition?.includes('错误率') || c.condition?.includes('error')) {
      errorRules.push({
        condition: c.condition,
        actions: [
          { type: 'notify', description: '报告', roles: c.roles },
        ],
      });
    }
  }
  
  return { rules: errorRules };
}

function calculateAvgConfidence(constraints: any[]): number {
  if (constraints.length === 0) return 0;
  const sum = constraints.reduce((acc, c) => acc + (c.confidence || 0), 0);
  return sum / constraints.length;
}
```

**Step 2: Create src/generator/skill-md.ts**

```typescript
import type { SkillSchema } from '../types/index.js';

export function generateSkillMarkdown(schema: SkillSchema): string {
  const { meta, constraints, steps, triggers } = schema;

  let md = `# ${meta.name}\n\n`;
  md += `**Version**: ${meta.version}\n`;
  md += `**Generated**: ${meta.generated_at || new Date().toISOString().split('T')[0]}\n`;
  if (meta.source) {
    md += `**Source**: ${meta.source}\n`;
  }
  md += '\n---\n\n';

  // Overview
  md += `## 概述\n\n`;
  md += `${meta.description || `${meta.name} 是一个标准操作流程 Skill。`}\n\n`;

  // Triggers
  md += `## 触发条件\n\n`;
  md += `| 类型 | 名称 | 描述 |\n`;
  md += `|------|------|------|\n`;
  for (const trigger of triggers) {
    md += `| ${trigger.type} | ${trigger.name} | ${trigger.description || '-'} |\n`;
  }
  md += '\n';

  // Steps
  if (steps.length > 0) {
    md += `## 执行步骤\n\n`;
    for (let i = 0; i < steps.length; i++) {
      const step = steps[i];
      md += `### Step ${i + 1}: ${step.name}\n\n`;
      md += `${step.description}\n\n`;
    }
  }

  // Constraints
  if (constraints.length > 0) {
    md += `## 约束规则\n\n`;

    const byLevel = {
      MUST: constraints.filter(c => c.level === 'MUST'),
      SHOULD: constraints.filter(c => c.level === 'SHOULD'),
      MAY: constraints.filter(c => c.level === 'MAY'),
    };

    for (const [level, items] of Object.entries(byLevel)) {
      if (items.length === 0) continue;

      const label = {
        MUST: 'MUST（必须遵守）',
        SHOULD: 'SHOULD（应当遵守）',
        MAY: 'MAY（可以遵守）',
      }[level];

      md += `### ${label}\n\n`;
      for (const c of items) {
        md += `- **${c.id}**: ${c.description}\n`;
      }
      md += '\n';
    }
  }

  // Error Handling
  if (schema.error_handling?.rules) {
    md += `## 错误处理\n\n`;
    md += `| 条件 | 处理 |\n`;
    md += `|------|------|\n`;
    for (const rule of schema.error_handling.rules) {
      const actions = rule.actions.map((a: any) => a.description).join(', ');
      md += `| ${rule.condition} | ${actions} |\n`;
    }
    md += '\n';
  }

  return md;
}
```

**Step 3: Create src/generator/manifest.ts**

```typescript
import type { Manifest } from '../types/index.js';
import yaml from 'js-yaml';

export function generateManifestYaml(manifest: Manifest): string {
  return yaml.dump(manifest, { indent: 2, lineWidth: -1 });
}
```

**Step 4: Create src/generator/test-cases.ts**

```typescript
import type { TestCase, TestCaseManifest } from '../types/index.js';
import fs from 'fs/promises';
import path from 'path';
import yaml from 'js-yaml';

export async function generateTestCaseFiles(
  cases: TestCase[],
  outputDir: string
): Promise<void> {
  // Group by type
  const byType: Record<string, TestCase[]> = {
    'happy-path': [],
    'edge-case': [],
    'error-case': [],
    'compliance': [],
  };

  for (const c of cases) {
    if (byType[c.type]) {
      byType[c.type].push(c);
    } else {
      byType['happy-path'].push(c);
    }
  }

  // Create manifest
  const manifest: TestCaseManifest = {
    dataset_name: 'Auto-generated Test Cases',
    version: '1.0.0',
    owner: 'sop-to-skill',
    description: 'Test cases auto-generated from SOP',
    generated_at: new Date().toISOString(),
    skill_name: 'unknown',
    summary: {
      total_cases: cases.length,
      happy_path: byType['happy-path'].length,
      edge_cases: byType['edge-case'].length,
      error_cases: byType['error-case'].length,
    },
    case_files: [],
  };

  // Write manifest
  await fs.mkdir(path.join(outputDir), { recursive: true });
  await fs.writeFile(
    path.join(outputDir, 'manifest.yaml'),
    yaml.dump(manifest, { indent: 2 })
  );

  // Write case files by type
  for (const [type, typeCases] of Object.entries(byType)) {
    if (typeCases.length === 0) continue;

    const typeDir = path.join(outputDir, type.replace('-', '-'));
    await fs.mkdir(typeDir, { recursive: true });

    for (const c of typeCases) {
      const filename = `${c.case_id}.yaml`;
      await fs.writeFile(
        path.join(typeDir, filename),
        yaml.dump(c, { indent: 2 })
      );
      manifest.case_files.push({ path: `${type}/${filename}`, tags: c.tags });
    }
  }
}
```

**Step 5: Create src/generator/index.ts**

```typescript
export { generateSkillPackage, type GeneratorOptions } from './skill-package.js';
export { generateSkillMarkdown } from './skill-md.js';
export { generateManifestYaml } from './manifest.js';
export { generateTestCaseFiles } from './test-cases.js';
```

**Step 6: Write tests**

Create: `sop-to-skill/src/generator/generator.test.ts`

```typescript
import { describe, it, expect } from 'vitest';
import { generateSkillMarkdown } from './skill-md.js';
import type { SkillSchema } from '../types/index.js';

describe('Generator', () => {
  describe('generateSkillMarkdown', () => {
    it('should generate markdown from schema', () => {
      const schema: SkillSchema = {
        meta: {
          name: '测试Skill',
          version: '1.0.0',
          generated_at: '2026-03-29',
        },
        triggers: [
          { type: 'execution', name: 'execute', input_schema: { type: 'object' } },
        ],
        steps: [
          {
            id: 'step_1',
            name: '测试步骤',
            description: '这是一个测试步骤',
            action: 'test_action',
            input: {},
            output: {},
          },
        ],
        constraints: [
          {
            id: 'C001',
            level: 'MUST',
            description: '必须执行某操作',
            roles: ['admin'],
            confidence: 0.9,
          },
        ],
      };

      const md = generateSkillMarkdown(schema);

      expect(md).toContain('# 测试Skill');
      expect(md).toContain('**Version**: 1.0.0');
      expect(md).toContain('C001');
      expect(md).toContain('MUST');
    });
  });
});
```

**Step 7: Commit**

```bash
git add src/generator/
git commit -m "feat: add generator module for skill package generation"
```

---

## Phase 6: Validator 模块

### Task 6.1: 创建 Validator

**Files:**
- Create: `sop-to-skill/src/validator/index.ts`

**Step 1: Create src/validator/index.ts**

```typescript
import fs from 'fs/promises';
import path from 'path';

export interface ValidationIssue {
  type: string;
  severity: 'error' | 'warning';
  message: string;
}

export interface ValidationResult {
  valid: boolean;
  score: number;
  issues: ValidationIssue[];
}

const REQUIRED_FILES = [
  'SKILL.md',
  'skill.schema.json',
  'skill.manifest.yaml',
];

const OPTIONAL_FILES = [
  'test-cases/manifest.yaml',
  'actions/',
  'constraints/',
  'decisions/',
];

export async function validateSkillPackage(
  packagePath: string
): Promise<ValidationResult> {
  const issues: ValidationIssue[] = [];

  // Check if path exists
  try {
    await fs.access(packagePath);
  } catch {
    return {
      valid: false,
      score: 0,
      issues: [{
        type: 'missing_directory',
        severity: 'error',
        message: `Package path does not exist: ${packagePath}`,
      }],
    };
  }

  const stat = await fs.stat(packagePath);
  const isDirectory = stat.isDirectory();

  // If it's a file, check if it's a ZIP
  if (!isDirectory) {
    if (path.extname(packagePath) === '.zip') {
      // TODO: Extract and validate
      return { valid: true, score: 1, issues: [] };
    }
    return {
      valid: false,
      score: 0,
      issues: [{
        type: 'invalid_format',
        severity: 'error',
        message: 'Package must be a directory or ZIP file',
      }],
    };
  }

  // Check required files
  let score = 0;
  let requiredFound = 0;

  for (const file of REQUIRED_FILES) {
    const filePath = path.join(packagePath, file);
    try {
      await fs.access(filePath);
      requiredFound++;
    } catch {
      issues.push({
        type: 'missing_file',
        severity: 'error',
        message: `Required file missing: ${file}`,
      });
    }
  }

  score = requiredFound / REQUIRED_FILES.length;

  // Check optional files
  for (const file of OPTIONAL_FILES) {
    const filePath = path.join(packagePath, file);
    try {
      await fs.access(filePath);
      score = Math.min(score + 0.05, 1);
    } catch {
      issues.push({
        type: 'missing_file',
        severity: 'warning',
        message: `Optional file not found: ${file}`,
      });
    }
  }

  // Validate SKILL.md content
  const skillMdPath = path.join(packagePath, 'SKILL.md');
  try {
    const content = await fs.readFile(skillMdPath, 'utf-8');
    if (content.length < 100) {
      issues.push({
        type: 'empty_content',
        severity: 'warning',
        message: 'SKILL.md seems empty',
      });
    }
  } catch {}

  // Validate schema JSON
  const schemaPath = path.join(packagePath, 'skill.schema.json');
  try {
    const content = await fs.readFile(schemaPath, 'utf-8');
    JSON.parse(content); // Validate JSON
  } catch (e: any) {
    issues.push({
      type: 'invalid_json',
      severity: 'error',
      message: `skill.schema.json is not valid JSON: ${e.message}`,
    });
  }

  const errors = issues.filter(i => i.severity === 'error');
  return {
    valid: errors.length === 0,
    score,
    issues,
  };
}
```

**Step 2: Write tests**

Create: `sop-to-skill/src/validator/validator.test.ts`

```typescript
import { describe, it, expect, beforeEach } from 'vitest';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';
import { validateSkillPackage } from './index.js';

describe('Validator', () => {
  let tempDir: string;

  beforeEach(async () => {
    tempDir = await fs.mkdtemp(path.join(os.tmpdir(), 'sop-test-'));
  });

  it('should validate complete package', async () => {
    // Create minimal package
    await fs.writeFile(path.join(tempDir, 'SKILL.md'), '# Test Skill\n\nContent');
    await fs.writeFile(path.join(tempDir, 'skill.schema.json'), JSON.stringify({}));
    await fs.writeFile(path.join(tempDir, 'skill.manifest.yaml'), 'package:\n  name: test');

    const result = await validateSkillPackage(tempDir);
    expect(result.valid).toBe(true);
    expect(result.score).toBeGreaterThan(0.5);
  });

  it('should fail on missing required files', async () => {
    const result = await validateSkillPackage(tempDir);
    expect(result.valid).toBe(false);
    expect(result.issues.some(i => i.severity === 'error')).toBe(true);
  });
});
```

**Step 3: Commit**

```bash
git add src/validator/
git commit -m "feat: add validator module"
```

---

## Phase 7: 集成测试

### Task 7.1: 创建集成测试

**Files:**
- Create: `sop-to-skill/integration/test-sop.md`
- Create: `sop-to-skill/src/integration.test.ts`

**Step 1: Create integration/test-sop.md**

```markdown
# 测试SOP流程

**版本**: V1.0
**日期**: 2026-03-29

## 1. 目的

本规程用于测试sop-to-skill的转换功能。

## 2. 约束规则

### 2.1 必须遵守

- **C001**: 必须在5个工作日内完成审批
- **C002**: 必须由管理员授权
- **C003**: 如果金额超过10000，则必须经理审批

### 2.2 应当遵守

- **C004**: 应当记录操作日志
- **C005**: 应当在24小时内通知申请人

### 2.3 可以

- **C006**: 可以使用加急通道
- **C007**: 如果系统故障，可以手动处理

## 3. 错误处理

如果审批超时，则通知上级主管。
```

**Step 2: Create integration test**

Create: `sop-to-skill/src/integration.test.ts`

```typescript
import { describe, it, expect, beforeAll } from 'vitest';
import fs from 'fs/promises';
import path from 'path';
import { parseInputFile } from './parser/factory.js';
import { extractFromText } from './extractor/index.js';
import { generateSkillPackage } from './generator/skill-package.js';

const TEST_SOP = path.join(process.cwd(), 'integration/test-sop.md');

describe('Integration', () => {
  it('should convert SOP to Skill Package', async () => {
    // Parse
    const parsed = await parseInputFile(TEST_SOP);
    expect(parsed.content).toBeTruthy();

    // Extract
    const extracted = extractFromText(parsed.content);
    expect(extracted.constraints.length).toBeGreaterThan(0);

    // Generate
    const pkg = await generateSkillPackage(extracted, '测试流程', {
      sourceFile: TEST_SOP,
    });

    expect(pkg.schema.meta.name).toBe('测试流程');
    expect(pkg.schema.constraints.length).toBeGreaterThan(0);
    expect(pkg.manifest.package.type).toBe('skill');
  });
});
```

**Step 3: Run integration test**

```bash
cd sop-to-skill
npm test
```

**Step 4: Commit**

```bash
git add integration/
git add src/integration.test.ts
git commit -m "test: add integration tests"
```

---

## Phase 8: 文档和完善

### Task 8.1: 完善 README 和示例

**Files:**
- Create: `sop-to-skill/examples/data-verification/SKILL.md`
- Update: `sop-to-skill/README.md`

**Step 1: Create examples/data-verification/SKILL.md**

```markdown
# 数据核查流程

**Version**: 1.0.0
**Source**: SOP-DM-002
**Type**: Domain Skill - 临床试验数据管理

---

## 概述

本 Skill 用于执行标准化的临床试验数据核查流程，确保数据质量和合规性。

## 触发条件

| 类型 | 名称 | 输入 | 输出 |
|------|------|------|------|
| execution | execute_verification | record_id, record_type, verification_items | verification_result |
| query | query_status | verification_id | status |
| approval | approve_exception | exception_id, reason | approval_result |

## 执行步骤

### Step 1: 验证约束条件
检查数据是否符合基础规范要求。

### Step 2: 执行 SDV
源数据核查，核对 EDC 与源文件的一致性。

### Step 3: 逻辑检查
验证数据内部逻辑一致性。

### Step 4: 发送质疑
通过 EDC 系统发起质疑。

### Step 5: 跟踪关闭
确认质疑回复并关闭。

## 约束规则

### MUST（必须遵守）
- **C001**: 必须核查主要疗效终点、安全性数据、关键入排标准
- **C002**: 必须由授权 DMC 人员发送质疑
- **C003**: 质疑后 5 个工作日内必须回复（安全性数据 24 小时）

### SHOULD（应当遵守）
- **C004**: SDV 错误率超过 5% 时扩大核查范围
- **C005**: 应当通过系统通知研究者质疑状态

## 错误处理

| 条件 | 动作 |
|------|------|
| 错误率 > 5% | 扩大核查 + 报告数据经理 |
| 系统性问题 | 启动 RCA + CAPA |

---

*此为示例 Skill，完整定义参见 skill.schema.json*
```

**Step 2: Update README**

Add usage examples, API reference, and configuration options.

**Step 3: Commit**

```bash
git add examples/
git add README.md
git commit -m "docs: add examples and improve README"
```

---

## 实现总结

### 任务清单

| Phase | Task | 描述 | 优先级 |
|-------|------|------|--------|
| 0 | 0.1 | 项目初始化 | P0 |
| 0 | 0.2 | 核心类型定义 | P0 |
| 0 | 0.3 | 格式规范文件 | P1 |
| 1 | 1.1 | CLI 框架 | P0 |
| 2 | 2.1 | Parser 模块 | P0 |
| 3 | 3.1 | Extractor 模块 | P0 |
| 4 | 4.1 | LLM 增强模块 | P0 |
| 5 | 5.1 | Generator 模块 | P0 |
| 6 | 6.1 | Validator 模块 | P1 |
| 7 | 7.1 | 集成测试 | P1 |
| 8 | 8.1 | 文档完善 | P2 |

### 预计工时

- Phase 0-1 (基础): 4-6 小时
- Phase 2-5 (核心): 8-12 小时
- Phase 6-8 (完善): 4-6 小时
- **总计**: 16-24 小时

---

## 下一步

**Plan complete and saved to `docs/plans/sop-to-skill-implementation-plan.md`.**

**Two execution options:**

1. **Subagent-Driven (this session)** - I dispatch fresh subagent per task, review between tasks, fast iteration

2. **Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
