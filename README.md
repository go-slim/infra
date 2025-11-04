# Infra

简体中文 | [English](README.en-US.md)

Go 基础设施库。

## 概述

这是一个 Go 模块基础设施库，为 Go 应用程序提供通用工具和实用程序。

## 当前状态

📦 **初始设置**

此项目目前处于初始设置阶段，包含：

- Go 模块配置 (`go.mod`)
- 项目文档
- 使用 Jujutsu (jj) 的开发环境设置

## 开发

此项目使用 [Jujutsu (jj)](https://github.com/martinvonz/jj) 进行版本控制。

### 开始使用

```bash
# 克隆仓库
git clone <repository-url>
cd infra

# 安装 jj（如果尚未安装）
# macOS: brew install jujutsu
# 其他平台: https://github.com/martinvonz/jj/releases

# 检查仓库状态
jj status
jj log
```

### 贡献

1. 创建新变更：`jj new`
2. 进行更改
3. 运行测试：`go test ./...`
4. 提交：`jj commit -m "描述"`
5. 推送更改到远程仓库

## 提交日志规范

请遵循以下提交日志格式：

### 格式

```
<类型>: <描述>

[可选的详细说明]

[可选的关闭问题]
```

### 类型

- `功能`: 新功能
- `修复`: 修复 bug
- `文档`: 文档相关
- `样式`: 代码格式化（不影响功能）
- `重构`: 重构代码
- `测试`: 添加或修改测试
- `构建`: 构建系统或依赖相关
- `CI`: CI 配置相关

### 示例

```
功能: 添加重试机制

实现了基于指数退避的重试算法，支持：
- 最大重试次数配置
- 初始延迟设置
- 最大延迟限制

Closes #1
```

## 模块信息

- **模块**: `go-slim.dev/infra`
- **Go 版本**: 1.25

## 许可证

[在此添加您的许可证]
