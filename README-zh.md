# TPM — 终端临时密码管理器

> English version: [README.md](README.md)

一款轻量、跨平台的命令行工具，专为 **临时 / 动态密码**（TOTP、Token）而生，面向
AI 时代：对终端和脚本友好，输出干净，AI Agent 可以直接捕获使用。

TPM **不是** 长期静态密码保险箱（如 1Password），它专注于自动化脚本和 AI Agent
在执行任务过程中所需的短时验证码、动态 Token 和 OTP。

## 核心亮点

- **单一静态二进制** —— Go 编写，无运行时、无依赖，丢进 `PATH` 即可使用。
- **全平台支持** —— macOS、Windows、Linux（amd64 + arm64）。
- **静态加密存储** —— 每个条目都是独立的 `AES-256-GCM` 加密 `.enc` 文件，云盘里
  没有任何可读明文。
- **云盘同步无锁定** —— 数据目录指向 iCloud / OneDrive / Google Drive，由它们的
  官方客户端负责同步。一个条目一个文件，多设备不会产生合并冲突。
- **主密钥托管于系统密钥库** —— macOS Keychain / Windows 凭据管理器 / Linux
  Secret Service。主密钥永不落地、永不上云。
- **AI / 管道友好** —— `--silent` 仅输出密码本身（无换行、无颜色、无修饰）。

## 安装

### 从源码构建

```bash
git clone https://github.com/zhangshuaike/tpm.git
cd tpm
make build        # 生成 ./bin/tpm
# 或安装到 PATH：
make install      # 安装到 $GOPATH/bin（通常是 ~/go/bin）
```

> 如果安装后仍提示 `command not found: tpm`，说明 `~/go/bin` 不在你的 PATH 中。
> 在 `~/.zshrc` 末尾加入下面这行后，执行 `source ~/.zshrc` 或重开终端：
>
> ```bash
> export PATH="$HOME/go/bin:$PATH"
> ```

### 交叉编译各平台发布版

```bash
make release      # 输出到 dist/tpm-<os>-<arch>[.exe]
```

### Linux 构建前置依赖

`go-keyring` 通过 D-Bus 与 Secret Service 通信。Debian/Ubuntu 系统在构建前需安装
开发头文件：

```bash
sudo apt-get install -y libsecret-1-dev
```

运行时还需要有可用的 Secret Service 服务（GNOME Keyring、KWallet 等）。

## 快速上手

### 1. 初始化

把数据目录指向你的云盘同步文件夹。主密钥会自动生成并存入系统密钥库。

```bash
# macOS + iCloud
tpm init --dir "~/Library/Mobile Documents/com~apple~CloudDocs/tpm-data"

# Windows + OneDrive
tpm init --dir "%USERPROFILE%\OneDrive\tpm-data"
```

在 **第二台设备** 上，安全地传输主密钥并导入：

```bash
tpm init --dir "~/OneDrive/tpm-data" --import
# 按提示粘贴 base64 主密钥
```

### 2. 添加凭证

凭证**按类型组织**：`code`（动态验证码 / TOTP）或 `password`（静态密码）。
同一个名字可以同时挂多种类型。

```bash
# 一个动态验证码（TOTP）
tpm add code github --secret "JBSWY3DPEHPK3PXP" --desc "GitHub 2FA"

# 让一个名字同时拥有 code 和 password
tpm add code     tencent --secret "JBSWY3DPEHPK3PXP"
tpm add password tencent --secret "s3cr3t!"
```

### 3. 读取凭证

语法是 `tpm <类型> <名字>`。当一个名字只有一种类型时，可以用裸名捷径
`tpm <名字>`。

```bash
# 显式指定类型
$ tpm code github
[TPM] Generating code for github...
Code: 482910 (Expires in 18s)

# 裸名捷径（该名字只有一种类型）
$ tpm github
Code: 482910 (Expires in 18s)

# 一个名字有多种类型时，要指明取哪种
$ tpm code tencent
$ tpm password tencent

# AI / 脚本模式 —— 只有值本身，无换行
$ tpm code github --silent
482910

# 直接注入环境变量
export MY_TOKEN=$(tpm code github --silent)
```

### 4. 管理条目

```bash
tpm list                                # 列出名字及其类型，如 "tencent [code, password]"
tpm show tencent                        # 查看一个名字下有哪些类型/元数据（不显示密文）
tpm rename github vk              # 重命名整个条目
tpm rename tencent tencent-pw --type password  # 只把某一种类型移动到另一个名字
tpm rm tencent --type password         # 只删除某一种类型
tpm rm github                    # 删除整个条目
```

## 工作原理

```
<data_dir>/                  # 位于你的云盘同步文件夹内
├── .tpm_meta                # 非敏感元数据（版本、加密算法）
└── entries/
    ├── github-2fa.enc       # AES-256-GCM 密文（nonce 前置）
    └── prod-mysql.enc

~/.config/tpm/env.toml       # 每台机器的本地配置：指向 <data_dir>
系统密钥库 (service=tpm)       # 保存 32 字节主密钥
```

读取流程：从 `env.toml` 加载 `data_dir` → 从密钥库取主密钥 → 读取
`entries/<name>.enc` → 解密 → 取出请求的类型对应凭证 → 调用其 Provider → 输出。
一个 `<name>.enc` 文件里可以同时存多种类型的凭证。

## 凭证类型

| 类型       | 状态   | 说明                                                          |
| ---------- | ------ | ------------------------------------------------------------- |
| `code`     | ✅     | 动态一次性验证码（RFC 6238 TOTP，6 位 / SHA1 / 30 秒）；兼容 Google 与 Microsoft 验证器 |
| `password` | ✅     | 静态密码 / Token，原样返回                                    |
| `script`   | 规划中 | 执行自定义命令动态生成值                                      |

新类型只需实现 `provider.PasswordProvider` 接口并调用 `provider.Register` 注册即可，
核心 CLI 逻辑无需改动。

## 命令一览

| 命令                                 | 说明                                        |
| ------------------------------------ | ------------------------------------------- |
| `tpm init --dir <path>`              | 初始化存储 + 配置主密钥                      |
| `tpm add <type> <name> --secret ...` | 给某个名字添加/更新某种类型的凭证           |
| `tpm <type> <name> [--silent]`       | 输出凭证值（如 `tpm code github`）     |
| `tpm <name> [--silent]`              | 该名字只有一种类型时的捷径                   |
| `tpm list`                           | 列出名字及其类型                            |
| `tpm show <name>`                    | 查看条目的类型/元数据（不显示密文）         |
| `tpm rename <old> <new> [--type <type>]` | 重命名条目，或把某一种类型移到另一个名字 |
| `tpm rm <name> [--type <type>]`      | 删除整个条目，或只删某一种类型              |

## 开发

```bash
make test    # 运行单元测试
make vet     # 静态检查
make fmt     # 格式化
```

## 路线图

- `script` Provider（执行自定义命令生成值）
- `tpm export-key`：便于第二台设备安全导入主密钥
- 可选的本地确认（Touch ID / 回车确认）—— AI 读取敏感验证码前需人工放行
- 可插拔存储后端（S3 / Git）

## 许可证

MIT
