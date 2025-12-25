# Assistant Qisumi Windows 部署脚本
# 用法: .\deploy.ps1 [-Environment <String>]
# Environment: dev (默认) | prod

param(
    [string]$Environment = "dev"
)

$ErrorActionPreference = "Stop"

# 函数：打印成功消息
function Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

# 函数：打印警告消息
function Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

# 函数：打印错误消息
function Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
    exit 1
}

# 函数：打印信息
function Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Cyan
}

Write-Host ""
Info "开始部署 Assistant Qisumi (环境: $Environment)"
Write-Host ""

# 检查 Docker 是否安装
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Error "Docker 未安装，请先安装 Docker Desktop"
}

# 检查 Docker Compose 是否可用
$composeVersion = docker compose version 2>&1
if ($LASTEXITCODE -ne 0) {
    Error "Docker Compose 不可用，请确保 Docker Desktop 正在运行"
}

# 检查 .env 文件
if (-not (Test-Path .env)) {
    Warning ".env 文件不存在"
    if (Test-Path .env.example) {
        Info "从 .env.example 复制配置文件..."
        Copy-Item .env.example .env
        Warning "请编辑 .env 文件配置必要的环境变量"
        Pause
    } else {
        Error ".env.example 文件不存在"
    }
}

# 拉取最新代码
if (Test-Path .git) {
    Info "拉取最新代码..."
    try {
        git pull origin master
    } catch {
        Warning "无法拉取代码，继续使用本地代码"
    }
}

# 停止现有容器
Info "停止现有容器..."
docker compose down 2>$null

# 构建镜像
Info "构建 Docker 镜像..."
if ($Environment -eq "prod") {
    docker compose build --no-cache
} else {
    docker compose build
}

# 启动服务
Info "启动服务..."
if ($Environment -eq "prod") {
    docker compose --profile production up -d
} else {
    docker compose up -d
}

# 等待服务启动
Info "等待服务启动..."
Start-Sleep -Seconds 10

# 检查服务状态
Info "检查服务状态..."
docker compose ps

Write-Host ""
Success "部署完成！"
Write-Host ""
Info "服务访问地址:"
if ($Environment -eq "prod") {
    Write-Host "  前端: http://localhost" -ForegroundColor White
    Write-Host "  后端 API: http://localhost:4569" -ForegroundColor White
    Write-Host "  HTTPS: https://localhost (需要配置 SSL 证书)" -ForegroundColor White
} else {
    Write-Host "  前端: http://localhost" -ForegroundColor White
    Write-Host "  后端 API: http://localhost:4569" -ForegroundColor White
}
Write-Host ""
Info "查看日志: docker compose logs -f"
Info "停止服务: docker compose down"
Info "重启服务: docker compose restart"
Write-Host ""

# 显示初始日志
Info "显示最近日志..."
docker compose logs --tail=20
