# 小鸡远程部署指南

[中文 README](../README.md) | [English README](../README.en.md) | [规则配置](rules.md)

小鸡远程（ChickReomte）由中继服务端、控制端和受控端组成。两种客户端角色使用相同的 `chickreomte-cli`，通过配置区分。本文命令均从仓库根目录执行；发行包用户需按实际目录调整路径。

## 准备程序

按 [README 构建说明](../README.md#从源码构建)生成 `bin/chickreomte-svr` 和 `bin/chickreomte-cli`，或从 [Releases](https://github.com/Mutantcat-Working-Group/ChickRemote/releases) 获取适合平台的包。旧发行包不一定包含当前源码的命名和修复。

Tauri 2 桌面端安装、配置和发布说明见[桌面指南](desktop.md)。macOS 截图已迁移至 ScreenCaptureKit，要求 macOS 14+。Desktop installation and release instructions are in the [desktop guide](desktop.md); macOS capture requires macOS 14+.

## 配置与安全边界

部署时保留 `conf/` 的目录结构。`#include common.yaml` 和 `#include rule.d/*.yaml` 是配置加载器处理的指令，不是可随意删除的注释。

1. 使用 `openssl rand -hex 32` 生成强随机密钥，将各端 `conf/common.yaml` 的 `secret` 改成同一带引号的字符串，替换示例值。
2. 将 `conf/local.yaml` 和 `conf/remote.yaml` 的 `server` 设置为同一中继地址。客户端 `id` 必须不同，规则 `target` 与受控端 ID 一致。
3. 将控制端 `dashboard.listen` 和每条规则的 `local_addr` 设置为 `127.0.0.1`。示例文件当前为 `0.0.0.0`，会监听所有网卡。
4. 管理面板、Shell、VNC 和 code-server 入口不提供独立登录认证。跨机器访问时使用 VPN 或带认证的反向代理，并阻断未经认证的直连访问；不要直接暴露到公网。
5. 跨不可信网络运行前启用并验证 TLS。中继 TLS 不会自动保护浏览器访问的 HTTP / WebSocket 入口。

共享密钥应只交给可信参与者；它不是按用户或按主机配置的细粒度访问控制。各机器需同步系统时间，因为握手签名依赖时间窗口。

`log.dir` 与 `codedir` 的相对路径基于可执行文件目录解析。目录必须可写；系统服务部署建议使用绝对路径，并限制配置文件的读取权限。

## 服务器端部署

修改 `conf/server.yaml` 的监听端口及 TLS 配置后运行：

```sh
./bin/chickreomte-svr --conf conf/server.yaml
```

默认端口为 TCP `6154`，防火墙应仅允许需要连接的客户端访问。

## 受控端部署

配置 `conf/remote.yaml` 的 `id`、`server` 和 `ssl` 后运行：

```sh
./bin/chickreomte-cli --conf conf/remote.yaml
```

示例 ID 为 `remote`，管理面板默认在此示例中关闭。使用具备必要权限的普通用户运行。VNC 需要可访问的桌面会话及对应系统权限；code-server 需要预先安装到受控端的 `PATH`。

## 控制端部署

配置 `conf/local.yaml` 和 `conf/rule.d/` 中的规则后运行：

```sh
./bin/chickreomte-cli --conf conf/local.yaml
```

连接成功后，在控制端打开 [http://127.0.0.1:8080](http://127.0.0.1:8080)。不要将规则端口与管理面板端口混为一谈；未配置 `local_port` 的规则会动态分配端口，可从管理面板进入。

## TLS 配置

在服务端的 `conf/server.yaml` 中配置证书和私钥的实际绝对路径：

```yaml
tls:
  key: /absolute/path/to/server.key
  crt: /absolute/path/to/server.crt
```

在两个客户端的配置中均设置：

```yaml
server: relay.example.com:6154
ssl:
  enabled: true
  insecure: false
```

将 `relay.example.com` 替换为真实域名，证书名称必须与连接地址匹配。私有 CA 证书应导入客户端系统信任库。`insecure: true` 会跳过证书校验，不只是关闭 SNI，不应作为生产环境自签证书的解决办法。

修改后重启相应进程。TLS 加密和共享密钥握手是不同机制，两者都需要正确配置。

## 注册系统服务

以下操作在具备系统服务管理权限的终端执行。将配置路径替换为实际绝对路径：

```sh
./bin/chickreomte-cli install --conf /absolute/path/to/conf/remote.yaml
./bin/chickreomte-cli start
./bin/chickreomte-cli status
./bin/chickreomte-cli restart
./bin/chickreomte-cli stop
./bin/chickreomte-cli uninstall
```

服务端用 `chickreomte-svr` 执行同名子命令。`install --user <用户名>` 可设置服务身份，`--user` 不适用于前台运行。Linux 使用 systemd，Windows 可通过 `services.msc` 管理服务。

## 旧版本升级

程序和系统服务名称为 `chickreomte-svr`、`chickreomte-cli`。先用旧程序停止并卸载旧服务，再安装新服务，并更新部署脚本中的程序名。备份并保留配置和密钥，重新建立浏览器会话。

历史版本的命令与兼容性信息见 [CHANGELOG](../CHANGELOG.md)，不能直接用作当前安装步骤。项目采用 [MIT License](../LICENSE)，允许商业使用；仅在已获得授权的设备上部署。
