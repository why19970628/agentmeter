# 贡献指南

[English](CONTRIBUTING.md)

感谢你帮助改进 AgentMeter。

## 提交前检查

运行相关检查：

```bash
make test
make build
```

完整本地检查：

```bash
make validate
```

如果没有 Make，也可以运行：

```bash
go test ./...
go build .
```

## 变更规则

- 修改解析逻辑需要补充小型脱敏 fixture 测试。
- Web UI 变更需要同时支持英文和中文文案。
- CLI 输出变更如果影响示例，需要同步更新测试和 README。
- Release workflow 变更需要保持 `.goreleaser.yml` 和 `.github/workflows/release.yml` 一致。
- 保持离线优先，不新增网络上传、隐藏 telemetry 或账号连接。

## 隐私规则

不要在 issue、测试、文档或提交中包含真实用户数据：

- API Key
- prompt
- 私有文件路径
- 完整本地日志
- 账号标识

请使用小型合成 fixture。

## Pull Request

请说明：

- 改了什么
- 如何验证
- 已知风险或限制
- 是否需要更新 README 或 release notes
