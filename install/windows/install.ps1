$ErrorActionPreference = "Stop"
$Root = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$Dest = Join-Path $env:LOCALAPPDATA "NekoPresenceKey"
New-Item -ItemType Directory -Force -Path $Dest | Out-Null
Copy-Item (Join-Path $Root "dist\nekopresence-windows-amd64.exe") (Join-Path $Dest "nekopresence.exe") -Force
$Exe = Join-Path $Dest "nekopresence.exe"
$Action = New-ScheduledTaskAction -Execute $Exe -Argument '--listen 0.0.0.0:45873 --timeout 20'
$Trigger = New-ScheduledTaskTrigger -AtLogOn -User $env:USERNAME
$Principal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -LogonType Interactive -RunLevel Limited
Register-ScheduledTask -TaskName "NekoPresenceKey" -Action $Action -Trigger $Trigger -Principal $Principal -Force | Out-Null
Start-ScheduledTask -TaskName "NekoPresenceKey"
Write-Host "NekoPresence Key installed and started for $env:USERNAME"
Write-Host "For initial pairing, stop the task and run: $Exe --pair"
