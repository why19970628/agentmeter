# AGENTS

[English](AGENTS.md)

本文档用于指导在本仓库工作的 AI 编码代理。

## 项目使命

AgentMeter 是一个离线优先的 Go 工具，用于读取本地 AI agent 用量日志，并统计 Codex、Claude Code、Cursor、Gemini CLI 以及兼容本地记录的 token 用量。

项目必须保持本地优先：

- 不上传用量数据。
- 不读取或打印 API Key。
- 不新增隐藏 telemetry 或后台同步。
- 不在文档或测试中写入真实用户路径、prompt 或私有日志样本。

## 项目结构

- Go module 位于仓库根目录。
- CLI 和 Web server 入口：`main.go`。
- 用量扫描与聚合：`internal/usage`。
- CLI 报告渲染：`internal/report`。
- Web dashboard 服务：`internal/web`。
- Web UI：`templates` 和 `static`。
- 示例数据：`examples`。

## 常用命令

优先使用 Makefile：

```bash
make test
make build
make validate
make run ARGS="summary -period all -group model -paths examples"
```

迭代时优先跑小范围检查：

```bash
go test ./internal/usage
go test ./internal/web
```

## 开发规则

- 解析器要容忍缺失字段、嵌套字段和 provider 特有字段。
- 修改解析逻辑必须补测试，尤其是缺字段、嵌套记录、损坏日志和 session 上下文推断。
- 尽量保持 CLI 输出稳定。如果变更输出列或 JSON 字段，需要同步更新测试和文档。
- Web 修改需要保持默认英文，并支持中文切换。
- Release assets 需要尽量自包含，确保包含运行 Web dashboard 所需的 templates 和 static 文件。
- 除非能明显提升可靠性，否则避免新增依赖，保持离线行为和单二进制简洁性。

## 验证要求

- 解析或聚合变更：运行 `go test ./internal/usage`。
- Web UI/server 变更：运行 `go test ./internal/web`。
- Report/CLI 变更：运行 `go test ./internal/report`，必要时加一次小范围 CLI smoke。
- Release 或 CI 变更：优先运行 `make validate`；如果条件不具备，至少运行 `go test ./...`。

交付或提交前运行：

```bash
go test ./...
```

## CI 和提交流程

- CI 必须运行 `make validate` 和 `make test-harness`（当这些目标存在时）。如果当前仓库还没有 `test-harness`，CI 至少保持 `make validate`，并在新增 harness 目标前记录例外原因。
- PR 必须包含摘要、需求分类、验收标准、变更区域、发布决策、TDD / 测试证据、验证、风险和回滚。
- 功能和错误修复 PR 必须标记合并后需要发版，或写明用户已明确批准延后发版。
- Harness/tooling、docs、CI 和其他工程流程优化 PR 应标记无需发版，除非它们影响已发布行为。
- 提交消息应符合 `{emoji} {type}{scope}: {subject}`，并且 emoji 必须匹配提交类型：
  - `✨ feat`
  - `🐛 fix`
  - `📝 docs`
  - `👷 ci`
  - `💄 style`
  - `♻️ refactor`
  - `🔖 release`
  - `⚡️ perf`
  - `✅ test`
  - `🔧 chore`
  - `🏗️ build`
- 如果仓库存在 commit hooks 或 commitlint，先运行一次 `make setup` 启用 `.githooks/commit-msg`，或直接运行 `make commitlint COMMIT_MSG_FILE=<commit-msg-file>`。当存在 commitlint 时，PR CI 应验证 PR range 内的提交消息。
- 不要在跳过必要本地验证后声称完成。
- Stage/commit 只包含当前迭代文件，不要引入无关脏文件。

## 发版说明

Release description 应描述用户可见变化，不要写入内部本机路径或私有测试命令。README、release notes 和示例中都不要出现真实机器路径。
