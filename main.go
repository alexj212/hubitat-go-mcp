package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	serverName    = "hubitat-mcp"
	serverVersion = "1.0.0"
)

var (
	modeFlag    = flag.String("mode", "stdio", "Server mode: 'stdio' for MCP stdio transport, 'sse' for HTTP/SSE server")
	portFlag    = flag.String("port", "", "Port for SSE server mode (overrides .env PORT)")
	versionFlag = flag.Bool("version", false, "Print version and exit")
	helpFlag    = flag.Bool("help", false, "Show help message")
)

type HubitatDevice struct {
	ID           string                 `json:"id"`
	Label        string                 `json:"label"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Capabilities []string               `json:"capabilities"`
	Commands     []map[string]string    `json:"commands"`
	Attributes   map[string]interface{} `json:"attributes"`
}

type Config struct {
	BaseURL string
	Token   string
	Port    string
}

var config Config

func loadConfig() error {
	// Load .env file if it exists
	_ = godotenv.Load()

	config.BaseURL = os.Getenv("HUBITAT_BASE_URL")
	config.Token = os.Getenv("HUBITAT_TOKEN")
	config.Port = os.Getenv("PORT")

	if config.Port == "" {
		config.Port = "5006"
	}

	if config.BaseURL == "" {
		return fmt.Errorf("HUBITAT_BASE_URL is required")
	}

	if config.Token == "" {
		return fmt.Errorf("HUBITAT_TOKEN is required")
	}

	return nil
}

func getDevices() ([]HubitatDevice, error) {
	url := fmt.Sprintf("%s/all?access_token=%s", config.BaseURL, config.Token)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch devices: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("hubitat API returned status %d: %s", resp.StatusCode, string(body))
	}

	var devices []HubitatDevice
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		return nil, fmt.Errorf("failed to decode devices: %w", err)
	}

	return devices, nil
}

func sendCommand(deviceID, command string) error {
	url := fmt.Sprintf("%s/%s/%s?access_token=%s", config.BaseURL, deviceID, command, config.Token)

	resp, err := http.Post(url, "", nil)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hubitat API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func sendCommandWithValue(deviceID, command, value string) error {
	url := fmt.Sprintf("%s/%s/%s/%s?access_token=%s", config.BaseURL, deviceID, command, value, config.Token)

	resp, err := http.Post(url, "", nil)
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("hubitat API returned status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func setupMCPServer() *server.MCPServer {
	s := server.NewMCPServer(
		serverName,
		serverVersion,
	)

	// Tool: List all devices
	listDevicesTool := mcp.NewTool("list_devices",
		mcp.WithDescription("List all Hubitat devices with their capabilities and current states"),
	)

	s.AddTool(listDevicesTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		devices, err := getDevices()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var deviceList []string
		for _, device := range devices {
			capabilities := strings.Join(device.Capabilities, ", ")
			deviceList = append(deviceList, fmt.Sprintf(
				"ID: %s | Label: %s | Type: %s | Capabilities: %s",
				device.ID, device.Label, device.Type, capabilities,
			))
		}

		return mcp.NewToolResultText(strings.Join(deviceList, "\n")), nil
	})

	// Tool: Get device details
	getDeviceTool := mcp.NewTool("get_device",
		mcp.WithDescription("Get detailed information about a specific Hubitat device"),
		mcp.WithString("device_id",
			mcp.Required(),
			mcp.Description("The ID of the device to query"),
		),
	)

	s.AddTool(getDeviceTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		deviceID, ok := arguments["device_id"].(string)
		if !ok {
			return mcp.NewToolResultError("device_id must be a string"), nil
		}

		devices, err := getDevices()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		for _, device := range devices {
			if device.ID == deviceID {
				deviceJSON, err := json.MarshalIndent(device, "", "  ")
				if err != nil {
					return mcp.NewToolResultError("failed to marshal device data"), nil
				}
				return mcp.NewToolResultText(string(deviceJSON)), nil
			}
		}

		return mcp.NewToolResultError(fmt.Sprintf("Device with ID %s not found", deviceID)), nil
	})

	// Tool: Turn device on
	turnOnTool := mcp.NewTool("turn_on",
		mcp.WithDescription("Turn on a Hubitat device (switches, lights, etc.)"),
		mcp.WithString("device_id",
			mcp.Required(),
			mcp.Description("The ID of the device to turn on"),
		),
	)

	s.AddTool(turnOnTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		deviceID, ok := arguments["device_id"].(string)
		if !ok {
			return mcp.NewToolResultError("device_id must be a string"), nil
		}

		if err := sendCommand(deviceID, "on"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully turned on device %s", deviceID)), nil
	})

	// Tool: Turn device off
	turnOffTool := mcp.NewTool("turn_off",
		mcp.WithDescription("Turn off a Hubitat device (switches, lights, etc.)"),
		mcp.WithString("device_id",
			mcp.Required(),
			mcp.Description("The ID of the device to turn off"),
		),
	)

	s.AddTool(turnOffTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		deviceID, ok := arguments["device_id"].(string)
		if !ok {
			return mcp.NewToolResultError("device_id must be a string"), nil
		}

		if err := sendCommand(deviceID, "off"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully turned off device %s", deviceID)), nil
	})

	// Tool: Set level (dimmer)
	setLevelTool := mcp.NewTool("set_level",
		mcp.WithDescription("Set the level of a dimmable device (0-100)"),
		mcp.WithString("device_id",
			mcp.Required(),
			mcp.Description("The ID of the device"),
		),
		mcp.WithNumber("level",
			mcp.Required(),
			mcp.Description("The level to set (0-100)"),
		),
	)

	s.AddTool(setLevelTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		deviceID, ok := arguments["device_id"].(string)
		if !ok {
			return mcp.NewToolResultError("device_id must be a string"), nil
		}

		level, ok := arguments["level"].(float64)
		if !ok {
			return mcp.NewToolResultError("level must be a number"), nil
		}

		if level < 0 || level > 100 {
			return mcp.NewToolResultError("level must be between 0 and 100"), nil
		}

		levelStr := strconv.Itoa(int(level))
		if err := sendCommandWithValue(deviceID, "setLevel", levelStr); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully set device %s to level %d", deviceID, int(level))), nil
	})

	// Tool: Send custom command
	customCommandTool := mcp.NewTool("send_command",
		mcp.WithDescription("Send a custom command to a Hubitat device"),
		mcp.WithString("device_id",
			mcp.Required(),
			mcp.Description("The ID of the device"),
		),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The command to send (e.g., 'refresh', 'configure')"),
		),
		mcp.WithString("value",
			mcp.Description("Optional value parameter for the command"),
		),
	)

	s.AddTool(customCommandTool, func(arguments map[string]interface{}) (*mcp.CallToolResult, error) {
		deviceID, ok := arguments["device_id"].(string)
		if !ok {
			return mcp.NewToolResultError("device_id must be a string"), nil
		}

		command, ok := arguments["command"].(string)
		if !ok {
			return mcp.NewToolResultError("command must be a string"), nil
		}

		value, hasValue := arguments["value"].(string)

		var err error
		if hasValue && value != "" {
			err = sendCommandWithValue(deviceID, command, value)
		} else {
			err = sendCommand(deviceID, command)
		}

		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully sent command '%s' to device %s", command, deviceID)), nil
	})

	return s
}

func printHelp() {
	fmt.Printf("Hubitat Go MCP Server v%s\n", serverVersion)
	fmt.Println("\nUsage:")
	fmt.Printf("  %s [options]\n\n", os.Args[0])
	fmt.Println("Options:")
	flag.PrintDefaults()
	fmt.Println("\nModes:")
	fmt.Println("  stdio - MCP stdio transport (default, for Claude Desktop)")
	fmt.Println("  sse   - HTTP/SSE server (for remote connections)")
	fmt.Println("\nExamples:")
	fmt.Printf("  %s                           # Run in stdio mode (default)\n", os.Args[0])
	fmt.Printf("  %s -mode=sse                 # Run as HTTP/SSE server\n", os.Args[0])
	fmt.Printf("  %s -mode=sse -port=8080      # Run SSE server on port 8080\n", os.Args[0])
	fmt.Println("\nEnvironment variables (from .env file):")
	fmt.Println("  HUBITAT_BASE_URL - Hubitat Maker API base URL (required)")
	fmt.Println("  HUBITAT_TOKEN    - Hubitat access token (required)")
	fmt.Println("  PORT             - Default port for SSE mode (default: 5006)")
}

func main() {
	flag.Parse()

	if *helpFlag {
		printHelp()
		os.Exit(0)
	}

	if *versionFlag {
		fmt.Printf("Hubitat Go MCP Server v%s\n", serverVersion)
		os.Exit(0)
	}

	if err := loadConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	// Override port if specified via flag
	if *portFlag != "" {
		config.Port = *portFlag
	}

	log.Printf("Starting Hubitat MCP Server v%s", serverVersion)
	log.Printf("Hubitat API: %s", config.BaseURL)
	log.Printf("Mode: %s", *modeFlag)

	mcpServer := setupMCPServer()

	switch *modeFlag {
	case "stdio":
		log.Println("Running in stdio mode (for Claude Desktop)")
		if err := server.ServeStdio(mcpServer); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	case "sse":
		log.Printf("Running in SSE mode on port %s", config.Port)
		baseURL := fmt.Sprintf("http://localhost:%s", config.Port)
		sseServer := server.NewSSEServer(mcpServer, baseURL)

		addr := fmt.Sprintf(":%s", config.Port)
		log.Printf("SSE server listening on %s", addr)
		log.Printf("SSE endpoint: %s/sse", baseURL)
		log.Printf("Status endpoint: %s/status", baseURL)

		if err := sseServer.Start(addr); err != nil {
			log.Fatalf("Server error: %v", err)
		}

	default:
		log.Fatalf("Invalid mode: %s (must be 'stdio' or 'sse')", *modeFlag)
	}
}
