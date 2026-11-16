# Beszel Agent for QNAP QTS (`BeszelAgent`)

一个与具体 Beszel 版本解耦的 QNAP QTS 运行环境与 Web 管理套件，参考上游 [`henrygd/beszel`](https://github.com/henrygd/beszel)。

<img width="1972" height="1168" alt="图片" src="https://github.com/user-attachments/assets/98854211-5738-4e03-9ae6-2f8b73e57823" />

## 特性

- **不绑定 Beszel 版本**：QPKG 仅提供运行环境、进程生命周期管理与 Web 管理界面，不内置硬编码版本的 `beszel-agent`。
- **QTS 桌面原生集成**：通过 `/cgi-bin/qpkg/BeszelAgent/` 嵌入 QTS 桌面，跟随 QTS 管理员会话 (`NAS_SID`) 校验权限，无需额外开放管理端口。
- **在线下载与多源更新**：支持从 GitHub 官方 Releases、`gh.beszel.dev` 镜像、自定义 GitHub 代理前缀或任意直链下载任意版本 `beszel-agent`，自动校验架构与 SHA-256。
- **手动上传与回滚**：支持在网页端直接上传 `.tar.gz`、`.zip` 或裸 `beszel-agent` Linux ELF 二进制文件，自动备份旧版为 `.bak` 并支持一键回滚。
- **完整配置管理**：图形化配置所有常用 `beszel-agent` 环境变量（`KEY`、`HUB_URL`、`TOKEN`、`LISTEN`、`FILESYSTEM`、`EXTRA_FILESYSTEMS`、`DOCKER_HOST`、`SMART_DEVICES` 等），并提供自定义环境变量列表以兼容未来新版本。
- **QNAP 专属优化**：
  - 将 `DATA_DIR` 持久化到 `<QPKG_ROOT>/data/agent-data`（避免 QNAP `/` 内存盘重启丢失指纹）。
  - 自动探测 QNAP 存储卷（如 `/share/CACHEDEV*_DATA`）、Container Station Docker Socket (`/var/run/docker.sock`)、`hwmon` 传感器，并在 CI / Release 构建的 `x86_64`、`arm_64`、`arm-x41` 包内均内置静态编译的 `smartctl`（smartmontools 7.5，启动时自动维护 `/usr/local/sbin/smartctl` 软链接并注入 `PATH`；若本地手动打包未构建 ARM 版 `smartctl`，也可运行 `bash ./scripts/build-smartctl.sh` 或配合 Entware `smartmontools`）。
  - 中英文双语界面切换。

## 构建 QPKG

前置依赖：Go 1.24+（纯 Go 跨平台打包，无需 Docker 或 Linux 环境）。

```powershell
.\scripts\build.ps1
```

构建产物输出在 `dist\` 目录（包含 `x86_64`、`arm_64`、`arm-x41` 架构包）。

### 架构包对照表

| QPKG 文件后缀 | 适用架构 / 平台 | 常见 QNAP 机型示例 |
| --- | --- | --- |
| `x86_64` | Intel / AMD 64 位（`x86_64`） | TS-x51 / TS-x53 / TS-x64 / TS-x73 / TVS / QuTS hero x86 等 |
| `arm_64` | ARM 64 位（`aarch64`） | TS-x33 / TS-x28A / TS-x32X / TS-x35A 等 |
| `arm-x41` | ARM 32 位 Annapurna Labs（`armv7l`、`Platform = ARM_AL`） | **TS-x31P / TS-x31P2 / TS-x31P3 / TS-x31X / TS-x31+ / TS-x41** 等 |

> [!TIP]
> 不确定机型对应哪个包时，可在 NAS SSH 中执行以下命令查看：
> ```sh
> uname -m && getcfg "" Platform -f /etc/platform.conf
> ```
> 其中 `armv7l` + `ARM_AL`（例如 **TS-231P**）在 QNAP 官方体系中归属 **`arm-x41`**（而非早期的 Mindspeed `arm-x31` / `ARM_MS`）。

## 安装与测试

1. **安装 QPKG（支持 QTS 4.5 – 6.1+，已在 QTS 6.1 实测通过）**：
   - 在 QTS **App Center** 手动安装对应架构的 `.qpkg`（若开启了代码签名限制，需在 App Center 设置中勾选允许安装无有效数字签名的应用程序）。
   - 如需通过 SSH 安装（或环境开启了严格签名校验），可执行 `qpkg_cli -m <file>.qpkg -A` 或以 root 直接运行安装包：
     ```sh
     sudo sh BeszelAgent_1.0.0_x86_64.qpkg
     ```
2. 在 QTS 桌面点击 **Beszel Agent** 图标打开管理界面。
3. 在 **二进制与更新** 页在线下载最新版或手动上传 `beszel-agent` 二进制。
4. 在 **配置信息** 页填写 `KEY`（或 `HUB_URL` + `TOKEN`）及可选存储卷 `FILESYSTEM`（如需采集 SATA 硬盘 SMART，建议设置 `SMART_DEVICES` 如 `/dev/sda:sat,/dev/sdb:sat,/dev/nvme0`），点击 **保存并重启 Agent**。
