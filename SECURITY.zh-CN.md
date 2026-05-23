# 安全策略

[English](SECURITY.md)

## 支持版本

安全修复面向最新发布版本和 `main` 分支。

## 报告漏洞

请避免在维护者有时间调查前公开披露漏洞。

如果 GitHub 支持 private vulnerability reporting，请优先使用；否则可以创建 issue，并提供最小脱敏复现。

## 数据和隐私边界

AgentMeter 是离线优先工具。安全相关变更必须遵守：

- 默认不上传本地日志、prompt、API Key、token 用量或项目路径。
- 不读取、打印、哈希、指纹化或持久化原始 API Key。
- 不新增隐藏 telemetry、后台同步或账号连接。
- 优先使用小型合成 fixture，不使用真实用户日志。
- 报告问题时移除私有路径、prompt 和个人数据。

## 报告内容

有效报告建议包含：

- 受影响版本或 commit
- 操作系统
- 复现步骤
- 预期行为
- 实际行为
- 影响范围

分享日志或 fixture 前，请先移除密钥和个人数据。
