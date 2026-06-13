param(
  [switch]$SkipTauri
)

$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
$BackendDir = Join-Path $Root "backend"
$FrontendDir = Join-Path $Root "frontend"
$TauriDir = Join-Path $Root "src-tauri"
$SidecarDir = Join-Path $TauriDir "binaries"
$SidecarName = "phototransfer-backend-x86_64-pc-windows-msvc.exe"
$SidecarPath = Join-Path $SidecarDir $SidecarName

function Resolve-CommandPath {
  param(
    [string]$Name,
    [string[]]$Candidates
  )

  $cmd = Get-Command $Name -ErrorAction SilentlyContinue
  if ($cmd) {
    return $cmd.Source
  }

  foreach ($candidate in $Candidates) {
    if (Test-Path $candidate) {
      return $candidate
    }
  }

  throw "Cannot find $Name. Install it or add it to PATH."
}

$Npm = Resolve-CommandPath "npm.cmd" @(
  "C:\nvm4w\nodejs\npm.cmd",
  "C:\Program Files\nodejs\npm.cmd"
)

$Cargo = $null
if (-not $SkipTauri) {
  $Cargo = Resolve-CommandPath "cargo.exe" @(
    "$env:USERPROFILE\.cargo\bin\cargo.exe",
    "C:\Users\Administrator\.cargo\bin\cargo.exe"
  )
  $CargoBin = Split-Path -Parent $Cargo
  if ($env:Path -notlike "*$CargoBin*") {
    $env:Path = "$CargoBin;$env:Path"
  }
}

New-Item -ItemType Directory -Force -Path $SidecarDir | Out-Null

Push-Location $BackendDir
try {
  $env:GOCACHE = Join-Path $Root ".gocache"
  $env:GOMODCACHE = Join-Path $Root ".gomodcache"
  go test ./...
  go build -o $SidecarPath ./cmd/server
}
finally {
  Pop-Location
}

Push-Location $FrontendDir
try {
  & $Npm install
  & $Npm run typecheck
  & $Npm run build
}
finally {
  Pop-Location
}

if (-not $SkipTauri) {
  Push-Location $Root
  try {
    & $Npm install
    & $Npm run tauri:build
  }
  finally {
    Pop-Location
  }
}

Write-Host "Build preparation complete."
Write-Host "Backend sidecar: $SidecarPath"
Write-Host "Tauri artifacts: $TauriDir\target\release\bundle"
