# Assistant Qisumi Windows 更新脚本
# 用法: .\update.ps1 [-Version <String>]
# Version: 要更新的版本号（可选，默认拉取最新代码）

param(
    [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

# 函数
function Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
    exit 1
}

function Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Cyan
}

Write-Host ""
Info "开始更新 Assistant Qisumi (版本: $Version)"
Write-Host ""

# 备份数据库
Info "备份数据库..."
$backupDir = "backups\$(Get-Date -Format 'yyyyMMdd_HHmmss')"
New-Item -ItemType Directory -Force -Path $backupDir | Out-Null

try {
    $timestamp = [int](Get-Date -UFormat %s)
    docker exec qisumi-backend cp "/app/data/assistant.db" "/tmp/backup_$timestamp.db"
    docker cp "qisumi-backend:/app/data/assistant.db" "$backupDir\assistant.db"
    Success "数据库已备份到: $backupDir"
} catch {
    Warning "数据库备份失败，继续更新"
}

# 拉取最新代码
if ($Version -eq "latest") {
    Info "拉取最新代码..."
    git fetch --all
    git checkout master
    git pull origin master
} else {
    Info "切换到版本 $Version..."
    git fetch --all --tags
    $tag = "v$Version"
    git checkout $tag
    if ($LASTEXITCODE -ne 0) {
        Error "版本 v$Version 不存在"
    }
}

# 检查配置文件变更
if (Test-Path .env.example) {
    Info "检查环境变量配置..."
    $envDiff = Compare-Object (Get-Content .env) (Get-Content .env.example) -ErrorAction SilentlyContinue
    if ($envDiff) {
        Warning "检测到 .env.example 有新配置项"
        Warning "建议检查并更新 .env 文件"
    }
}

# 重新构建镜像
Info "重新构建 Docker 镜像..."
docker compose build --no-cache

# 重启服务
Info "重启服务..."
docker compose down
docker compose up -d

# 等待服务启动
Info "等待服务启动..."
Start-Sleep -Seconds 15

# 检查服务状态
Info "检查服务状态..."
$status = docker compose ps
if ($status -match "Exit") {
    Error "服务启动失败，请查看日志"
}

Success "更新完成！"
Write-Host ""

# 显示当前版本
try {
    $currentVersion = git describe --tags --always 2>$null
    if ($currentVersion) {
        Info "当前版本: $currentVersion"
    }
} catch {}

Info "查看日志: docker compose logs -f"
Write-Host ""
