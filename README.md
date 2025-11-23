# Hubitat Go MCP Server

A native Go implementation of a Model Context Protocol (MCP) server for Hubitat home automation systems. This server enables AI assistants like Claude Desktop to control and query Hubitat devices through the MCP protocol.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)
![License](https://img.shields.io/badge/License-MIT-green)
[![MCP](https://img.shields.io/badge/MCP-Compatible-blueviolet)](https://modelcontextprotocol.io)

## Features

- ✅ **Native MCP Implementation** - Built with [mcp-go](https://github.com/mark3labs/mcp-go)
- ✅ **Full Device Control** - Turn devices on/off, set levels, send custom commands
- ✅ **Device Discovery** - List all devices with capabilities
- ✅ **Multiple Modes** - stdio (Claude Desktop) or SSE (HTTP server)
- ✅ **Cross-Platform** - Linux, macOS, and Windows support
- ✅ **Systemd Integration** - Run as a system service on Linux
- ✅ **Environment Configuration** - Easy setup with .env files
- ✅ **Production Ready** - Includes Makefile and systemd service

## Prerequisites

1. **Hubitat Hub** with [Maker API](https://docs.hubitat.com/index.php?title=Maker_API) enabled
2. **Maker API App ID** and **Access Token** from your Hubitat hub
3. **Go 1.21+** installed
4. **Claude Desktop** or another MCP-compatible client

## Installation

### Windows

See **[WINDOWS.md](WINDOWS.md)** for complete Windows installation guide.

**Quick Start (Windows)**:
```powershell
git clone https://github.com/alexj212/hubitat-go-mcp.git
cd hubitat-go-mcp
.\install-windows.ps1
```

### Linux / macOS

#### Quick Start (Automated)

```bash
# Clone the repository
git clone https://github.com/alexj212/hubitat-go-mcp.git
cd hubitat-go-mcp

# Run installation script
./install-claude-desktop.sh
```

#### Manual Installation

```bash
# Clone the repository
git clone https://github.com/alexj212/hubitat-go-mcp.git
cd hubitat-go-mcp

# Copy environment template
cp .env.example .env

# Edit .env with your Hubitat configuration
nano .env

# Build and install
make install

# Optional: Install as systemd service (Linux only)
make install-service
sudo systemctl start hubitat-go-mcp
```

### Manual Build

```bash
# Download dependencies
make deps

# Build binary
make build

# Run directly
./hubitat-go-mcp
```

## Configuration

Create a `.env` file in the project directory:

```bash
# Hubitat Configuration
HUBITAT_BASE_URL=http://192.168.1.253/apps/api/131/devices
HUBITAT_TOKEN=your-access-token-here

# Server Configuration (optional)
PORT=5006
```

### Getting Hubitat Credentials

1. Log into your Hubitat hub (e.g., `http://192.168.1.253`)
2. Go to **Apps** → **Hubitat Maker API**
3. Note the **App ID** (in the URL: `/apps/api/[APP_ID]/`)
4. Copy the **Access Token**
5. Construct your base URL: `http://<hub-ip>/apps/api/<app-id>/devices`

## Claude Desktop Integration

### Automatic Setup

**Linux/macOS**:
```bash
./install-claude-desktop.sh
```

**Windows**:
```powershell
.\install-windows.ps1
```

### Manual Setup

Add this to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "hubitat": {
      "command": "/usr/local/bin/hubitat-go-mcp"
    }
  }
}
```

**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "hubitat": {
      "command": "C:\\path\\to\\hubitat-go-mcp\\hubitat-go-mcp.exe"
    }
  }
}
```

**Linux**: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "hubitat": {
      "command": "/usr/local/bin/hubitat-go-mcp"
    }
  }
}
```

Then restart Claude Desktop.

## Running Modes

The server supports two modes of operation:

### Mode 1: stdio (Default - for Claude Desktop)

This is the default mode used by Claude Desktop for local MCP server integration:

```bash
./hubitat-go-mcp
# or
./hubitat-go-mcp -mode=stdio
```

Claude Desktop will launch this automatically via the config file.

### Mode 2: SSE Server (for Remote Access)

Run as an HTTP server with Server-Sent Events for remote access:

```bash
./hubitat-go-mcp -mode=sse
# or specify a custom port
./hubitat-go-mcp -mode=sse -port=8080
```

This creates an HTTP server at:
- **SSE endpoint**: `http://localhost:5006/sse`
- **Status endpoint**: `http://localhost:5006/status`

You can then use tools like `mcp-proxy` or other MCP clients to connect remotely.

### Command-Line Options

```bash
./hubitat-go-mcp -help
```

Available options:
- `-mode=stdio|sse` - Server mode (default: stdio)
- `-port=PORT` - Port for SSE mode (overrides .env)
- `-version` - Print version and exit
- `-help` - Show help message

## Available Tools

The MCP server exposes the following tools to Claude:

### 1. `list_devices`
Lists all Hubitat devices with their capabilities and current states.

**Example**: "List all my Hubitat devices"

### 2. `get_device`
Gets detailed information about a specific device.

**Parameters**:
- `device_id` (string, required): The device ID

**Example**: "Get details for device 135"

### 3. `turn_on`
Turns on a device (switches, lights, etc.).

**Parameters**:
- `device_id` (string, required): The device ID

**Example**: "Turn on the kitchen lights"

### 4. `turn_off`
Turns off a device.

**Parameters**:
- `device_id` (string, required): The device ID

**Example**: "Turn off all bedroom lights"

### 5. `set_level`
Sets the brightness level for dimmable devices.

**Parameters**:
- `device_id` (string, required): The device ID
- `level` (number, required): Brightness level (0-100)

**Example**: "Set the living room lights to 50%"

### 6. `send_command`
Sends a custom command to a device.

**Parameters**:
- `device_id` (string, required): The device ID
- `command` (string, required): The command to send
- `value` (string, optional): Optional command parameter

**Example**: "Refresh device 85"

## Usage Examples

Once configured with Claude Desktop, you can:

```
"List all my Hubitat devices"

"Turn on the Office Lamp"

"Set the Master bedroom lights to 30%"

"What is the temperature in the Office?"

"Turn off all lights in the Kitchen"

"Get status of device 135"
```

## Systemd Service Management

```bash
# Start the service
sudo systemctl start hubitat-go-mcp

# Stop the service
sudo systemctl stop hubitat-go-mcp

# Restart the service
sudo systemctl restart hubitat-go-mcp

# Check status
sudo systemctl status hubitat-go-mcp

# View logs
sudo journalctl -u hubitat-go-mcp -f

# Enable auto-start on boot
sudo systemctl enable hubitat-go-mcp

# Disable auto-start
sudo systemctl disable hubitat-go-mcp
```

## Development

```bash
# Run tests
make test

# Run in development mode (requires 'air' for auto-reload)
make dev

# Clean build artifacts
make clean

# Show all make targets
make help
```

## Makefile Targets

- `make deps` - Download Go dependencies
- `make build` - Build the binary
- `make install` - Install binary to `/usr/local/bin`
- `make install-service` - Install systemd service
- `make uninstall-service` - Remove systemd service
- `make clean` - Clean build artifacts
- `make test` - Run tests
- `make run` - Build and run
- `make dev` - Development mode with auto-reload
- `make help` - Show help

## Architecture

This server uses:
- **[mcp-go](https://github.com/mark3labs/mcp-go)** - Go implementation of the Model Context Protocol
- **stdio transport** - Communicates with Claude Desktop via standard input/output
- **Hubitat Maker API** - REST API for device control
- **godotenv** - Environment variable management

## Troubleshooting

### Service won't start

```bash
# Check service status
sudo systemctl status hubitat-go-mcp

# View logs
sudo journalctl -u hubitat-go-mcp -n 50
```

### Configuration errors

Ensure your `.env` file has the correct values:
```bash
cat /docker/hubitat-go-mcp/.env
```

### Test Hubitat connection

```bash
# Test the Hubitat API directly
curl "http://192.168.1.253/apps/api/131/devices/all?access_token=YOUR_TOKEN"
```

### Claude Desktop not detecting server

1. Verify the binary is installed: `which hubitat-go-mcp`
2. Check Claude Desktop config file syntax
3. Restart Claude Desktop completely
4. Check Claude Desktop logs for errors

## Comparison with Python Version

| Feature | Go Version | Python Version |
|---------|------------|----------------|
| MCP Protocol | Native (mcp-go) | Requires mcp-proxy |
| Performance | Fast, compiled | Slower, interpreted |
| Dependencies | Single binary | Python + packages |
| Memory | Low (~10MB) | Higher (~50MB+) |
| Setup | Direct integration | Needs proxy layer |

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT License - See LICENSE file for details

## Links

- [Model Context Protocol](https://modelcontextprotocol.io)
- [mcp-go SDK](https://github.com/mark3labs/mcp-go)
- [Hubitat Maker API](https://docs.hubitat.com/index.php?title=Maker_API)
- [Claude Desktop](https://claude.ai/download)

## Author

Alex Jeannopoulos ([@alexj212](https://github.com/alexj212))

## Acknowledgments

- Built with [mcp-go](https://github.com/mark3labs/mcp-go) by Mark3Labs
- Inspired by the [hubitat-mcp](https://github.com/abeardmore/hubitat-mcp) Python implementation
