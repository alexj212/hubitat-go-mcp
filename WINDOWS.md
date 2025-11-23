# Windows Installation Guide

Complete guide for installing Hubitat Go MCP Server on Windows.

## Prerequisites

1. **Go 1.21 or later**
   - Download from: https://go.dev/dl/
   - During installation, ensure "Add to PATH" is checked
   - Verify: Open PowerShell and run `go version`

2. **Git** (optional, for cloning)
   - Download from: https://git-scm.com/download/win

3. **Claude Desktop**
   - Download from: https://claude.ai/download

## Quick Install (Automated)

### Option 1: Using PowerShell Script

1. **Download or clone the repository**
   ```powershell
   git clone https://github.com/alexj212/hubitat-go-mcp.git
   cd hubitat-go-mcp
   ```

2. **Run the installation script**
   ```powershell
   .\install-windows.ps1
   ```

3. **Configure your Hubitat credentials**
   ```powershell
   notepad .env
   ```

   Update with your values:
   ```
   HUBITAT_BASE_URL=http://192.168.1.253/apps/api/131/devices
   HUBITAT_TOKEN=your-access-token-here
   PORT=5006
   ```

4. **Restart Claude Desktop**

Done! The MCP server is now configured.

## Manual Installation

### Step 1: Build the Binary

```powershell
# Download dependencies
go mod download

# Build
go build -o hubitat-go-mcp.exe
```

### Step 2: Configure Environment

```powershell
# Copy template
copy .env.example .env

# Edit with your credentials
notepad .env
```

### Step 3: Configure Claude Desktop

1. **Locate Claude Desktop config file**:
   ```
   %APPDATA%\Claude\claude_desktop_config.json
   ```

2. **Create or edit the file** (use full path to exe):
   ```json
   {
     "mcpServers": {
       "hubitat": {
         "command": "C:\\Users\\YourName\\hubitat-go-mcp\\hubitat-go-mcp.exe"
       }
     }
   }
   ```

   **Important**: Use double backslashes `\\` in the path!

3. **Restart Claude Desktop**

## Running Modes

### Mode 1: Claude Desktop (stdio mode - default)

This is the default mode and is automatically used by Claude Desktop:

```powershell
.\hubitat-go-mcp.exe
```

### Mode 2: HTTP/SSE Server (for remote access)

Run as a background service accessible via HTTP:

```powershell
.\hubitat-go-mcp.exe -mode=sse
```

Access at: `http://localhost:5006/sse`

To specify a different port:
```powershell
.\hubitat-go-mcp.exe -mode=sse -port=8080
```

## Running as a Windows Service

### Option 1: Using NSSM (Non-Sucking Service Manager)

1. **Download NSSM**: https://nssm.cc/download

2. **Install the service**:
   ```powershell
   # For stdio mode (Claude Desktop)
   nssm install HubitatMCP "C:\path\to\hubitat-go-mcp.exe"

   # For SSE mode (HTTP server)
   nssm install HubitatMCP "C:\path\to\hubitat-go-mcp.exe" "-mode=sse"

   # Set working directory
   nssm set HubitatMCP AppDirectory "C:\path\to\hubitat-go-mcp"
   ```

3. **Start the service**:
   ```powershell
   nssm start HubitatMCP
   ```

4. **Manage the service**:
   ```powershell
   nssm stop HubitatMCP
   nssm restart HubitatMCP
   nssm remove HubitatMCP
   ```

### Option 2: Using Task Scheduler

1. Open Task Scheduler
2. Create Basic Task
3. Name: "Hubitat MCP Server"
4. Trigger: "At startup"
5. Action: "Start a program"
6. Program: `C:\path\to\hubitat-go-mcp.exe`
7. Arguments: `-mode=sse` (if using SSE mode)
8. Start in: `C:\path\to\hubitat-go-mcp`

## Getting Hubitat Credentials

1. **Log into your Hubitat hub**
   - Usually: `http://192.168.1.XXX`

2. **Go to Apps → Hubitat Maker API**

3. **Note the App ID** from the URL:
   ```
   http://192.168.1.253/installedapp/configure/131/mainPage
                                                ^^^
                                              App ID
   ```

4. **Copy the Access Token** from the page

5. **Construct your base URL**:
   ```
   http://<hub-ip>/apps/api/<app-id>/devices
   ```

   Example:
   ```
   http://192.168.1.253/apps/api/131/devices
   ```

## Testing the Installation

### Test 1: Version Check
```powershell
.\hubitat-go-mcp.exe -version
```

### Test 2: Help
```powershell
.\hubitat-go-mcp.exe -help
```

### Test 3: Verify Configuration
```powershell
# This will show startup logs
.\hubitat-go-mcp.exe
# Press Ctrl+C to stop
```

### Test 4: Test in Claude Desktop

1. Open Claude Desktop
2. Look for tools/hammer icon
3. Type: "List my Hubitat devices"
4. Claude should show your devices

## Troubleshooting

### "Go is not recognized"

The Go installation path is not in your PATH. Either:
- Reinstall Go and check "Add to PATH"
- Or add manually to PATH: `C:\Program Files\Go\bin`

### "cannot find .env file"

Make sure you're running the command from the `hubitat-go-mcp` directory:
```powershell
cd C:\path\to\hubitat-go-mcp
.\hubitat-go-mcp.exe
```

### Claude Desktop doesn't see the server

1. Check config file location: `%APPDATA%\Claude\claude_desktop_config.json`
2. Verify path uses double backslashes: `C:\\Users\\...`
3. Ensure path is absolute, not relative
4. Restart Claude Desktop completely

### "Configuration error: HUBITAT_BASE_URL is required"

Your `.env` file is missing or incorrect:
1. Check file exists: `dir .env`
2. Open and verify: `notepad .env`
3. Ensure no spaces around `=`

### SSE server won't start

Check if port is already in use:
```powershell
netstat -ano | findstr :5006
```

Try a different port:
```powershell
.\hubitat-go-mcp.exe -mode=sse -port=8080
```

## Building for Different Architectures

### Build for Windows AMD64 (most common)
```powershell
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o hubitat-go-mcp-amd64.exe
```

### Build for Windows ARM64
```powershell
$env:GOOS="windows"; $env:GOARCH="arm64"; go build -o hubitat-go-mcp-arm64.exe
```

### Build for Windows 32-bit
```powershell
$env:GOOS="windows"; $env:GOARCH="386"; go build -o hubitat-go-mcp-386.exe
```

## Uninstallation

### Remove from Claude Desktop

Edit `%APPDATA%\Claude\claude_desktop_config.json` and remove the `"hubitat"` entry.

### Remove Service (if installed)

```powershell
# If using NSSM
nssm stop HubitatMCP
nssm remove HubitatMCP confirm

# If using Task Scheduler
# Open Task Scheduler and delete the task
```

### Delete Files

```powershell
cd C:\path\to\hubitat-go-mcp
cd ..
rmdir /s hubitat-go-mcp
```

## Support

- GitHub Issues: https://github.com/alexj212/hubitat-go-mcp/issues
- Hubitat Documentation: https://docs.hubitat.com/index.php?title=Maker_API
- MCP Protocol: https://modelcontextprotocol.io

## Advanced Configuration

### Custom Environment Variables

Instead of using `.env`, you can set system environment variables:

```powershell
$env:HUBITAT_BASE_URL="http://192.168.1.253/apps/api/131/devices"
$env:HUBITAT_TOKEN="your-token-here"
$env:PORT="5006"

.\hubitat-go-mcp.exe -mode=sse
```

### Running Multiple Instances

You can run multiple instances for different Hubitat hubs:

**Claude Desktop config**:
```json
{
  "mcpServers": {
    "hubitat-home": {
      "command": "C:\\Users\\You\\hubitat-home\\hubitat-go-mcp.exe"
    },
    "hubitat-office": {
      "command": "C:\\Users\\You\\hubitat-office\\hubitat-go-mcp.exe"
    }
  }
}
```

Each directory needs its own `.env` file with different credentials.
