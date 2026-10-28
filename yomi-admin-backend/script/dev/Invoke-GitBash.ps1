# 在 Windows 上通过 Git Bash 执行 .sh 脚本（供 Makefile 调用）
# 用法:
#   powershell -NoProfile -ExecutionPolicy Bypass -File script/dev/Invoke-GitBash.ps1 script/dev/test.sh
#   powershell ... Invoke-GitBash.ps1 script/_build/build.sh linux_amd64

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$ScriptPath,

    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$ScriptArguments
)

$ErrorActionPreference = "Stop"

function Get-GitBashExecutable {
    # make（32 位）下 $env:ProgramFiles 可能指向 Program Files (x86)，
    # 因此同时探测 ProgramW6432 与两个标准安装位置。
    $rootCandidates = @(
        [Environment]::GetEnvironmentVariable("ProgramW6432"),
        [Environment]::GetEnvironmentVariable("ProgramFiles"),
        [Environment]::GetEnvironmentVariable("ProgramFiles(x86)"),
        [Environment]::GetEnvironmentVariable("LocalAppData")
    )
    $roots = @()
    foreach ($root in $rootCandidates) {
        if ($root) {
            $roots += $root
        }
    }

    $candidates = @()
    foreach ($root in $roots) {
        $candidates += (Join-Path $root "Git\bin\bash.exe")
        $candidates += (Join-Path $root "Git\usr\bin\bash.exe")
    }
    $candidates += "C:\Program Files\Git\bin\bash.exe"
    $candidates += "C:\Program Files\Git\usr\bin\bash.exe"

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return (Resolve-Path $candidate).Path
        }
    }

    $git = Get-Command git -ErrorAction SilentlyContinue
    if ($git) {
        $bash = Join-Path (Split-Path (Split-Path $git.Source -Parent) -Parent) "bin\bash.exe"
        if (Test-Path $bash) {
            return (Resolve-Path $bash).Path
        }
    }

    return $null
}

$bash = Get-GitBashExecutable
if (-not $bash) {
    Write-Error @"
Git Bash not found. Install Git for Windows: https://git-scm.com/download/win
Expected bash.exe under Git installation (e.g. C:\Program Files\Git\bin\bash.exe).
"@
    exit 1
}

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
if (-not [System.IO.Path]::IsPathRooted($ScriptPath)) {
    $ScriptPath = Join-Path $repoRoot $ScriptPath
}
if (-not (Test-Path $ScriptPath)) {
    Write-Error "Script not found: $ScriptPath"
    exit 1
}

$scriptForBash = (Resolve-Path $ScriptPath).Path -replace "\\", "/"

& $bash $scriptForBash @ScriptArguments
exit $LASTEXITCODE
