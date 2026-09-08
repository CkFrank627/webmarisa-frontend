# start_all.ps1
$ErrorActionPreference = "Stop"

$root = "C:\web-marisa-1.0.0\web-marisa"
$serverDir = Join-Path $root "server"
$clientDir = Join-Path $root "client"

# 1) 载入私密环境变量（如果文件存在）
$envFile = Join-Path $root ".env.local.ps1"
if (Test-Path $envFile) {
  . $envFile
} else {
  Write-Host "WARNING: $envFile not found. Create it to store LLM key safely." -ForegroundColor Yellow
}

# 2) 启动后端（新窗口）
Start-Process powershell -ArgumentList @(
  "-NoExit",
  "-Command",
  "cd `"$serverDir`"; go run ."
)

# 3) 启动前端（新窗口）
Start-Process powershell -ArgumentList @(
  "-NoExit",
  "-Command",
  "cd `"$clientDir`"; npm run serve"
)

Write-Host "Started server + client in two new PowerShell windows."
Write-Host "Frontend: http://localhost:8888/   Backend: http://localhost:3000/"
