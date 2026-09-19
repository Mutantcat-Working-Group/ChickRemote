# 小鸡远程（ChickReomte）

**简体中文** | [English](README.en.md)

[![Build](https://github.com/Mutantcat-Working-Group/ChickRemote/actions/workflows/desktop.yml/badge.svg)](https://github.com/Mutantcat-Working-Group/ChickRemote/actions/workflows/desktop.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

小鸡远程是一款可私有化部署的远程主机管理工具。通过 Go 中继服务和客户端，在浏览器中访问远程终端、桌面与 code-server 开发环境。

当前版本 **1.0.20260919** 提供 Tauri 2 桌面客户端、命令行客户端和 Web 管理界面。桌面安装包内置 Go 客户端，无需另行安装 Go、Rust 或 Node.js。

## 桌面安装

从 [Releases](https://github.com/Mutantcat-Working-Group/ChickRemote/releases) 下载：Windows x64 使用 NSIS `.exe`，macOS 14+ 分别提供 Intel / Apple Silicon `.dmg`，Linux x64 提供 `.AppImage`。Windows 安装器内置 WebView2 离线安装程序；macOS 应用及 DMG 使用 ad-hoc 签名，不是 Apple 公证，首次启动可能需在系统设置中允许。Windows 也可能出现 SmartScreen 提示。Linux 需可执行权限和 FUSE 2，远程桌面使用 X11 会话。

启动桌面应用，填写本机 ID、中继地址、共享密钥和可选的目标 ID，然后启动客户端并打开会话面板。受控端可将目标 ID 留空。需要先部署中继；桌面端不内置公共中继。默认启用 TLS，必须与中继配置匹配。macOS 受控端需要屏幕录制和辅助功能权限。

配置和日志保存在用户应用数据目录，界面支持中英文切换。桌面配置仅创建 VNC 规则，终端和 code-server 的高级规则仍通过 CLI 配置使用。详见[桌面与发布指南](docs/desktop.md)。

## 功能

| 功能 | 说明 |
| --- | --- |
| Web Shell | Linux、macOS 的 PTY 终端，以及 Windows 的 PowerShell / CMD |
| Web 远程桌面 | 屏幕查看、键鼠控制、滚动与剪贴板操作，受平台能力限制 |
| 远程开发 | 转发受控端的 code-server，需在受控端自行安装 |
| 管理面板 | 查看规则、虚拟链路、会话和流量统计 |
| 私有化部署 | 自行部署中继服务，客户端主动连接中继 |
| 传输协议 | Protobuf 消息、虚拟链路复用、可选 TLS |
| 服务运行 | 支持前台运行及注册系统服务 |

## 工作方式

```text
浏览器 -> 控制端 chickreomte-cli -> 中继 chickreomte-svr <- 受控端 chickreomte-cli
```

- **中继服务端**：转发客户端之间的消息，默认监听 TCP `6154`。
- **控制端**：开启管理面板和规则入口，通过 `target` 指定受控端 ID。
- **受控端**：连接同一中继，响应终端、桌面等请求。

控制端与受控端使用同一个客户端程序，通过配置区分角色。详细设计见[架构与实现](docs/desc.md)。

## 快速开始

以下命令从仓库根目录执行，适用于已有本机可运行二进制的环境；源码构建见下一节。发行包可在 [Releases](https://github.com/Mutantcat-Working-Group/ChickRemote/releases) 查看，旧包的命名和内容可能与当前源码不同。

### 1. 准备配置

| 文件 | 用途 |
| --- | --- |
| `conf/server.yaml` | 中继监听端口与 TLS 证书 |
| `conf/local.yaml` | 控制端 ID、中继地址、管理面板与规则引用 |
| `conf/remote.yaml` | 受控端 ID 与中继地址 |
| `conf/common.yaml` | 共享密钥、超时和日志配置 |
| `conf/rule.d/*.yaml` | Shell、VNC、code-server 规则 |

**启动前必须调整示例配置：**

- 将各机器 `common.yaml` 中的 `secret` 改为同一组强随机密钥，不要使用示例值。可用 `openssl rand -hex 32` 生成，并以带引号的 YAML 字符串保存。
- 将两端的 `server` 指向同一中继地址；每个客户端使用不同的 `id`，规则 `target` 必须匹配受控端 ID。示例分别为 `local` 和 `remote`。
- 将 `local.yaml` 的 `dashboard.listen` 和规则文件的 `local_addr` 改为 `127.0.0.1`。示例目前使用 `0.0.0.0`，会监听所有网卡。
- 跨不可信网络部署时先配置 TLS，不要直接使用明文示例。

配置使用自定义 `#include` 指令。分发配置时请保留 `common.yaml`、`rule.d/` 及相对目录关系；这些指令并非普通 YAML 注释。

### 2. 启动三个角色

在对应机器上分别运行；本机验证时可以使用三个终端：

```sh
# 中继服务端
./bin/chickreomte-svr --conf conf/server.yaml
```

```sh
# 受控端
./bin/chickreomte-cli --conf conf/remote.yaml
```

```sh
# 控制端
./bin/chickreomte-cli --conf conf/local.yaml
```

连接成功后，在控制端浏览器打开 [http://127.0.0.1:8080](http://127.0.0.1:8080)。规则未指定 `local_port` 时会动态分配端口，可从管理面板进入对应服务。

使用具备必要权限的普通用户运行；仅在安装系统服务等确有需要的操作中提升权限。远程桌面还需操作系统授予屏幕录制、辅助功能等权限。

### 3. 可选：注册系统服务

在具备系统服务管理权限的终端中运行，并将配置路径替换为实际绝对路径：

```sh
./bin/chickreomte-cli install --conf /absolute/path/to/conf/remote.yaml
./bin/chickreomte-cli start
./bin/chickreomte-cli status
./bin/chickreomte-cli stop
./bin/chickreomte-cli uninstall
```

服务端使用 `chickreomte-svr` 的同名子命令。`--user` 是 `install` 子命令的选项，不是前台运行参数。升级时先用旧程序停止并卸载旧服务，再安装新服务，保留配置和密钥。

## 从源码构建

需要 Git、Go，以及客户端原生桌面依赖所需的 C/C++ 工具链和平台开发库。`go.mod` 声明 Go 1.18；实际可用性还取决于平台、SDK 和依赖版本，不代表任意平台均可无条件构建。

```sh
git clone https://github.com/Mutantcat-Working-Group/ChickRemote.git
cd ChickRemote
go mod download
sh build
```

`build` 会先生成 Web 资源嵌入文件，再编译 `bin/chickreomte-svr` 和 `bin/chickreomte-cli`。不要在干净检出后跳过资源生成直接构建客户端。

仅构建服务端不需要原生桌面依赖：

```sh
go build -o bin/chickreomte-svr ./code/server
```

Go module 为 `org.mutantcat.chickreomte`，内部包路径统一使用该前缀。GitHub 仓库地址保持不变；自定义模块路径尚未配置远程解析，请克隆后构建，不要使用 `go get org.mutantcat.chickreomte` 安装。

### 平台与已知限制

- 项目包含 Linux、Windows、macOS 实现，但能力不完全一致；现有 VNC 后端不支持 Windows / Linux ARM。
- **macOS**：本地维护的 MIT 截图库补丁使用 ScreenCaptureKit，兼容新 SDK，要求 macOS 14+。支持 Intel 和 Apple Silicon。
- **Windows**：Web 资源目录使用符号链接。检出时需保留链接，或将链接替换为其指向的实际文件 / 目录。
- **code-server**：需在受控端安装并加入 `PATH`，不会随本项目自动安装。
- 相对 `log.dir` 和 `codedir` 基于可执行文件所在目录解析，而非配置文件目录。请保证目录可写，或配置绝对路径。

### 测试

在 `sh build` 完成资源生成且平台依赖满足后运行：

```sh
go test ./...
```

不依赖原生桌面后端的部分可单独验证：

```sh
go test ./code/network/... ./code/server/... ./code/hash ./code/utils
```

## 安全部署

- 管理面板及 Shell、VNC、code-server 入口**没有独立登录认证**。不要直接暴露到公网；使用回环绑定、VPN，或带认证的反向代理，并限制直连入口。
- 中继 TLS 仅保护客户端与中继之间的传输，不会自动为浏览器管理入口提供 HTTPS 或认证。
- 中继配置 `tls.key` / `tls.crt`，客户端设置 `ssl.enabled: true`、`ssl.insecure: false`，并使用可信证书及匹配的服务器名称。
- 共享密钥应视为对主机访问的授权凭据，只向可信参与者分发；限制中继和本地入口的网络访问。
- 仅连接自己拥有或获得明确授权的设备。避免在日志、截图或 Issue 中泄露密钥和主机信息。

## 界面预览

桌面客户端界面预览：

![小鸡远程桌面客户端](docs/imgs/desktop.png)

## 文档与贡献

- [部署指南](docs/startup.md)
- [规则配置](docs/rules.md)
- [架构与实现](docs/desc.md)
- [更新记录](CHANGELOG.md)
- [安全说明](SECURITY.md)
- [提交问题](https://github.com/Mutantcat-Working-Group/ChickRemote/issues)
- [提交 Pull Request](https://github.com/Mutantcat-Working-Group/ChickRemote/pulls)

欢迎提交问题、修复和文档改进。报告问题时请提供操作系统、架构、Go / SDK 版本、复现步骤及脱敏日志；代码变更请附相关测试，并同步更新中英文 README。

## 开源协议

本项目采用 [MIT License](LICENSE)，允许商业使用、修改、分发和再许可，但须保留版权声明及许可声明。软件按原样提供，不附带任何保证，完整条款以 `LICENSE` 为准。

保留原作者李文超的版权声明，后续贡献同样按 MIT 协议发布。第三方依赖和随附资源遵循各自的许可证，项目的 MIT 协议不替代其许可条款。
