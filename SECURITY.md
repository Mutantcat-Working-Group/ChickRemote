# 小鸡远程安全说明 / ChickReomte Security

[中文 README](README.md) | [English README](README.en.md)

## 部署注意事项

- 管理面板、Shell、VNC 和 code-server 入口没有独立登录认证，不应直接暴露到公网。绑定回环地址，或使用 VPN / 带认证的反向代理并限制直连入口。
- 更换示例共享密钥，仅向可信参与者分发；共享密钥不能代替细粒度访问控制。
- 跨不可信网络时启用中继 TLS 并验证证书。`ssl.insecure: true` 会跳过证书校验，生产环境应保持 `false`。
- 中继 TLS 不提供浏览器入口的 HTTPS，也不是客户端间端到端加密。中继必须可信，浏览器入口需单独保护。
- 仅以必要权限运行，保护配置、日志和工作目录，并仅连接已获授权的设备。

## 报告安全问题

请勿在公开 Issue 或 PR 中发布有效密钥、可直接利用的攻击细节或第三方主机信息。优先检查仓库 [Security 页面](https://github.com/Mutantcat-Working-Group/ChickRemote/security)是否提供私密漏洞报告入口；若未提供，可先提交不包含敏感细节的 Issue 请求维护者建立私密沟通渠道。

报告应包含受影响的提交或版本、运行平台、影响范围和脱敏复现步骤。当前未承诺固定安全响应时限或历史版本维护周期。

## Deployment Precautions

- The dashboard and Shell, VNC, and code-server endpoints have no independent login authentication. Do not expose them directly to the internet. Bind to loopback, or use a VPN / authenticated reverse proxy and restrict direct access.
- Replace the example shared secret and distribute it only to trusted participants. A shared secret is not fine-grained access control.
- Enable relay TLS and validate certificates across untrusted networks. `ssl.insecure: true` skips certificate verification; keep it `false` in production.
- Relay TLS neither provides HTTPS for browser endpoints nor end-to-end encryption between clients. Trust the relay and protect browser endpoints separately.
- Run with only necessary privileges, protect configuration, logs, and workspace data, and connect only to authorized devices.

## Reporting Vulnerabilities

Do not publish active credentials, actionable exploit details, or third-party host information in public issues or pull requests. Check the repository's [Security page](https://github.com/Mutantcat-Working-Group/ChickRemote/security) for a private reporting option first. If none is available, open an issue without sensitive details to request a private communication channel from the maintainers.

Include the affected commit or version, platform, impact, and sanitized reproduction steps. No fixed security response time or historical-version support period is currently promised.

## 许可证 / License

本项目采用 [MIT License](LICENSE)，允许商业使用。以上为安全使用建议，不为 MIT 增加用途限制；第三方依赖保留各自许可证。

This project uses the [MIT License](LICENSE), which permits commercial use. These security recommendations do not add use restrictions to MIT. Third-party dependencies retain their own licenses.
