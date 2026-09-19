# 小鸡远程规则配置

[中文 README](../README.md) | [English README](../README.en.md) | [部署指南](startup.md)

规则由连接发起方配置，放在客户端配置的 `rules` 列表中，或通过示例的 `#include rule.d/*.yaml` 加载。下列代码块均为规则列表片段，不能单独作为完整客户端配置。

## 公共字段

| 字段 | 含义 |
| --- | --- |
| `name` | 规则名称；在同一客户端配置中保持唯一 |
| `target` | 受控端客户端 ID，示例为 `remote` |
| `type` | `shell`、`vnc`、`code-server` 或用于测试的 `bench` |
| `local_addr` | 控制端 Web 入口监听地址；建议 `127.0.0.1` |
| `local_port` | 可选的 Web 入口端口；省略或为 `0` 时动态分配 |

规则端口提供的是本地 HTTP / WebSocket 服务，并非 SSH、RDP 或标准 VNC 协议端口。不要与管理面板或其他规则的端口冲突。管理页面和规则入口没有独立登录认证，不应直接对公网开放。

## Shell 规则

```yaml
- name: shell
  target: remote
  type: shell
  local_addr: 127.0.0.1
  # local_port: 8081
  # exec: /bin/bash
  env:
    - TERM=xterm
```

`exec` 为受控端要启动的可执行程序，不是带参数的整段 Shell 命令。省略时 Linux / macOS 优先选择 `bash`，其次为 `sh`；Windows 优先选择 `powershell`，其次为 `cmd`。`env` 为终端进程的环境变量列表。

## VNC 规则

```yaml
- name: vnc
  target: remote
  type: vnc
  local_addr: 127.0.0.1
  # local_port: 8082
  fps: 10
```

`fps` 为请求的每秒截屏次数，省略或设为 `0` 时使用 `10`，超过 `50` 时按 `50` 处理。实际帧率受屏幕尺寸、网络与系统性能影响。

- 受控端创建子进程执行截图及键鼠操作，主进程在本机 `127.0.0.1:6155` 至 `127.0.0.1:6955` 中选择端口通信，不需要向公网开放该范围。
- 现有后端不支持 Windows / Linux ARM。新版 macOS SDK 移除的截图 API 会导致客户端编译失败，见 [README](../README.md#平台与已知限制)。
- macOS 需授予屏幕录制、辅助功能等权限；其他系统也需具备可访问的图形会话。
- Windows RDP 最小化或断开后的捕获行为取决于系统与会话状态，部署时应验证[系统服务模式](startup.md#注册系统服务)下的行为。
- 旧版 Windows 的 Ctrl+Alt+Del 模拟还可能需要配置软件安全注意序列（SAS）策略，应由管理员按实际系统版本评估。

## code-server 规则

```yaml
- name: code-server
  target: remote
  type: code-server
  local_addr: 127.0.0.1
  # local_port: 8083
```

受控端需自行安装 [code-server](https://github.com/coder/code-server) 并加入运行客户端的用户或系统服务的 `PATH`。它不会随小鸡远程自动安装。工作目录数据位置由客户端 `codedir` 控制；相对路径基于客户端可执行文件目录，而非配置文件目录。

配置变更后需重启客户端。跨机器访问 Web 入口时，应使用 VPN 或带认证的反向代理，同时限制对原始规则端口的直接访问。
