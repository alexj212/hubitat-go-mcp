#!/bin/bash
# Installation script for adding Hubitat MCP to Claude Desktop

set -e

echo "🚀 Hubitat Go MCP - Claude Desktop Installation"
echo "================================================"
echo ""

# Install the binary
echo "📦 Installing binary to /usr/local/bin..."
cd "$(dirname "$0")"
make install

# Detect OS and set config path
if [[ "$OSTYPE" == "darwin"* ]]; then
    CONFIG_DIR="$HOME/Library/Application Support/Claude"
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    CONFIG_DIR="$HOME/.config/Claude"
else
    echo "❌ Unsupported operating system: $OSTYPE"
    echo "Please manually configure Claude Desktop."
    echo "Config file location (Windows): %APPDATA%\Claude\claude_desktop_config.json"
    exit 1
fi

CONFIG_FILE="$CONFIG_DIR/claude_desktop_config.json"

# Create config directory if it doesn't exist
mkdir -p "$CONFIG_DIR"

# Check if config file exists
if [ -f "$CONFIG_FILE" ]; then
    echo "⚠️  Config file already exists at: $CONFIG_FILE"
    echo ""
    echo "Please manually add the following to your mcpServers section:"
    echo ""
    echo '  "hubitat": {'
    echo '    "command": "/usr/local/bin/hubitat-go-mcp"'
    echo '  }'
    echo ""
    echo "Full path: $CONFIG_FILE"
else
    echo "📝 Creating new config file..."
    cat > "$CONFIG_FILE" << 'EOF'
{
  "mcpServers": {
    "hubitat": {
      "command": "/usr/local/bin/hubitat-go-mcp"
    }
  }
}
EOF
    echo "✅ Config file created at: $CONFIG_FILE"
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "Next steps:"
echo "1. Make sure your .env file is configured with Hubitat credentials"
echo "2. Restart Claude Desktop"
echo "3. Look for the tools/hammer icon in Claude Desktop"
echo "4. Try: 'List my Hubitat devices'"
echo ""
echo "For troubleshooting, check: $CONFIG_FILE"
