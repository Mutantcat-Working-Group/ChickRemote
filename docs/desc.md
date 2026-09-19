# 小鸡远程架构与实现

[中文 README](../README.md) | [English README](../README.en.md) | [部署指南](startup.md)

小鸡远程（ChickReomte）通过中继转发客户端间的 Protobuf 消息。控制端与受控端都主动连接中继，浏览器只连接控制端的管理面板和规则入口。

```text
浏览器 -- HTTP / WebSocket --> 控制端 chickreomte-cli
                                    |
                              TCP / 可选 TLS
                                    |
                              chickreomte-svr
                                    |
                              TCP / 可选 TLS
                                    |
                              受控端 chickreomte-cli
                                    |
                        Shell / 桌面子进程 / code-server
```

## 模块划分

Go module 为 `org.mutantcat.chickreomte`，以下目录均在此模块下：

| 目录 | 职责 |
| --- | --- |
| `code/server` | 中继服务、客户端握手与消息路由 |
| `code/client/conn` | 客户端连接、消息发送和虚拟链路分发 |
| `code/client/dashboard` | Web 管理面板与统计接口 |
| `code/client/rule` | Shell、VNC、code-server 与 bench 规则 |
| `code/network` | Protobuf 消息、编解码和网络封装 |
| `code/hash` | 基于共享密钥和时间窗口的握手签名 |
| `html` | 构建时嵌入 Go 程序的 Web 资源 |

GitHub 仓库仍为 [Mutantcat-Working-Group/ChickRemote](https://github.com/Mutantcat-Working-Group/ChickRemote)。Go 包名迁移不意味着协议命名空间也被修改：Protobuf 包名保留为 `network` 和 `vncnetwork`。

## 连接与会话流程

1. 两个客户端根据 `server` 连接同一中继；若启用 `ssl.enabled`，先建立 TLS 连接。
2. 客户端发送 ID 和握手签名。当前签名使用共享密钥作为 HMAC-SHA512 密钥，对按 60 秒窗口计算的时间值签名，并非旧文档所述的 MD5。
3. 服务端检查握手与签名。客户端和服务端需要共享相同密钥并同步系统时间。
4. 用户从控制端浏览器打开某条规则，控制端按 `target` 发送 `connect_req`，中继将其路由到受控端。
5. 受控端创建对应会话。例如 Shell 规则启动终端程序，VNC 使用桌面子进程，code-server 规则管理开发环境转发。
6. 受控端通过 `connect_rep` 返回结果。成功后，双方使用链路 ID 分发后续数据。
7. 关闭会话或连接断开时，通过相应断开处理清理链路与会话资源。

## 配置示例

使用仓库的 `conf/server.yaml`、`conf/remote.yaml`、`conf/local.yaml`，并按[部署指南](startup.md)替换共享密钥、配置证书及限制监听地址。开启服务端 TLS 后，两个客户端都必须启用 TLS，不能一侧明文、一侧加密。

例如，受控端 ID 为 `remote` 时，控制端的 Shell 规则可以是：

```yaml
rules:
  - name: shell
    target: remote
    type: shell
    local_addr: 127.0.0.1
    local_port: 8081
    env:
      - TERM=xterm
```

这是客户端配置片段，不是完整配置。`8081` 是控制端 Web Shell 入口，不是 SSH 或 RDP 端口。

## 安全边界

共享密钥握手不是浏览器登录系统，也不提供细粒度主机授权。中继 TLS 保护客户端到中继的链路，不是客户端之间的端到端加密；中继可以处理转发消息，应视为可信基础设施。

管理面板与规则入口不提供独立登录认证，需通过回环绑定、VPN 或外部认证保护。仅在已授权设备上运行，详见[安全说明](../SECURITY.md)。Tauri 2 桌面端托管 Go 客户端并生成回环配置，远程页面在系统浏览器打开，不获得 Tauri 权限。The Tauri desktop controller owns the Go client; remote pages open outside its privileged webview. See [desktop architecture and release guide](desktop.md).

## 历史图示

下列图片及同目录 `.drawio` 源文件保留为历史设计资料，可能包含旧品牌或过时配置。当前行为、命令与安全要求以本页和部署指南为准。

![历史部署示意](imgs/example.jpg)

![历史软件架构](imgs/architecture.jpg)
