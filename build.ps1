# Assistant Qisumi Windows 构建脚本
# 构建前端和后端，并将前端静态文件集成到后端

$ErrorActionPreference = "Stop"

function Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
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
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  Assistant Qisumi - 完整构建" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# 清理旧的构建产物
Info "清理旧的构建产物..."
if (Test-Path "static") { Remove-Item -Recurse -Force "static" }
if (Test-Path "assistant-qisumi.exe") { Remove-Item -Force "assistant-qisumi.exe" }
Success "清理完成"

# 构建前端
Info "构建前端..."
Push-Location frontend
npm install
npm run build
Pop-Location
Success "前端构建完成"

# 创建 static 目录
Info "准备静态文件目录..."
New-Item -ItemType Directory -Force -Path "static" | Out-Null

# 复制前端构建产物到 static 目录
Info "复制前端静态文件..."
Copy-Item -Path "frontend/dist\*" -Destination "static\" -Recurse -Force
Success "静态文件复制完成"

# 构建后端
Info "构建后端"
$env:CGO_ENABLED = "1"
go build -o assistant-qisumi.exe ./cmd/server
Success "后端构建完成"

Write-Host ""
Success "========================================="
Success "  构建完成！"
Success "========================================="
Write-Host ""
Info "可执行文件: .\assistant-qisumi.exe"
Info "静态文件目录: .\static\"
Write-Host ""
Info "运行应用:"
Write-Host "  .\assistant-qisumi.exe" -ForegroundColor White
Write-Host ""
