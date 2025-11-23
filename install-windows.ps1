# Hubitat Go MCP Server - Windows Installation Script
# Requires: Go 1.21+ installed and in PATH

param(
    [switch]$SkipBuild,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

function Show-Help {
    Write-Host "Hubitat Go MCP Server - Windows Installation"
    Write-Host ""
    Write-Host "Usage: .\install-windows.ps1 [options]"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -SkipBuild    Skip building the binary (use existing)"
    Write-Host "  -Help         Show this help message"
    Write-Host ""
    Write-Host "Prerequisites:"
    Write-Host "  - Go 1.21 or later installed"
    Write-Host "  - Git (optional, for cloning)"
    Write-Host ""
    Write-Host "What this script does:"
    Write-Host "  1. Builds the hubitat-go-mcp.exe binary"
    Write-Host "  2. Creates/updates Claude Desktop config"
    Write-Host "  3. Copies .env.example to .env (if needed)"
    Write-Host ""
}

if ($Help) {
    Show-Help
    exit 0
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Hubitat Go MCP - Windows Installation" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "✓ Go detected: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "✗ Go is not installed or not in PATH" -ForegroundColor Red
    Write-Host "  Please install Go from: https://go.dev/dl/" -ForegroundColor Yellow
    exit 1
}

# Get script directory
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $scriptDir

# Build the binary
if (-not $SkipBuild) {
    Write-Host ""
    Write-Host "Building hubitat-go-mcp.exe..." -ForegroundColor Cyan

    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Failed to download dependencies" -ForegroundColor Red
        exit 1
    }

    go build -o hubitat-go-mcp.exe
    if ($LASTEXITCODE -ne 0) {
        Write-Host "✗ Build failed" -ForegroundColor Red
        exit 1
    }

    Write-Host "✓ Build successful" -ForegroundColor Green
} else {
    if (-not (Test-Path "hubitat-go-mcp.exe")) {
        Write-Host "✗ hubitat-go-mcp.exe not found and -SkipBuild specified" -ForegroundColor Red
        exit 1
    }
    Write-Host "✓ Using existing hubitat-go-mcp.exe" -ForegroundColor Green
}

# Check/Create .env file
Write-Host ""
if (-not (Test-Path ".env")) {
    Write-Host "Creating .env file from template..." -ForegroundColor Cyan
    Copy-Item ".env.example" ".env"
    Write-Host "✓ .env file created" -ForegroundColor Green
    Write-Host "⚠  Please edit .env and add your Hubitat credentials!" -ForegroundColor Yellow
} else {
    Write-Host "✓ .env file already exists" -ForegroundColor Green
}

# Get binary path
$binaryPath = Join-Path $scriptDir "hubitat-go-mcp.exe"
$binaryPath = $binaryPath.Replace('\', '\\')  # Escape backslashes for JSON

# Configure Claude Desktop
Write-Host ""
Write-Host "Configuring Claude Desktop..." -ForegroundColor Cyan

$claudeConfigDir = Join-Path $env:APPDATA "Claude"
$claudeConfigFile = Join-Path $claudeConfigDir "claude_desktop_config.json"

# Create config directory if it doesn't exist
if (-not (Test-Path $claudeConfigDir)) {
    New-Item -ItemType Directory -Path $claudeConfigDir -Force | Out-Null
    Write-Host "✓ Created Claude config directory" -ForegroundColor Green
}

# Check if config file exists
if (Test-Path $claudeConfigFile) {
    Write-Host "⚠  Config file already exists: $claudeConfigFile" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Please manually add the following to your mcpServers section:" -ForegroundColor Yellow
    Write-Host ""
    Write-Host '  "hubitat": {' -ForegroundColor White
    Write-Host '    "command": "' -NoNewline -ForegroundColor White
    Write-Host $binaryPath -NoNewline -ForegroundColor Cyan
    Write-Host '"' -ForegroundColor White
    Write-Host '  }' -ForegroundColor White
    Write-Host ""
    Write-Host "Config file location: $claudeConfigFile" -ForegroundColor Gray
} else {
    # Create new config file
    $config = @{
        mcpServers = @{
            hubitat = @{
                command = $binaryPath
            }
        }
    }

    $config | ConvertTo-Json -Depth 10 | Set-Content $claudeConfigFile
    Write-Host "✓ Created Claude Desktop config: $claudeConfigFile" -ForegroundColor Green
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Installation Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Binary location: $binaryPath" -ForegroundColor White
Write-Host "Config file:     $claudeConfigFile" -ForegroundColor White
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Edit .env file with your Hubitat credentials"
Write-Host "  2. Restart Claude Desktop"
Write-Host "  3. Look for the tools icon in Claude Desktop"
Write-Host "  4. Try: 'List my Hubitat devices'"
Write-Host ""
Write-Host "Testing the installation:" -ForegroundColor Cyan
Write-Host "  .\hubitat-go-mcp.exe -help" -ForegroundColor Gray
Write-Host "  .\hubitat-go-mcp.exe -version" -ForegroundColor Gray
Write-Host ""
Write-Host "Running as SSE server (for remote access):" -ForegroundColor Cyan
Write-Host "  .\hubitat-go-mcp.exe -mode=sse" -ForegroundColor Gray
Write-Host ""
