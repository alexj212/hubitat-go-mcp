package main

import (
	"context"
	"encoding/json"
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
		server.WithToolCapabilities(true),
	)

	// Tool: List all devices
	listDevicesTool := mcp.NewTool("list_devices",
		mcp.WithDescription("List all Hubitat devices with their capabilities and current states"),
	)

	s.AddTool(listDevicesTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	s.AddTool(getDeviceTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, ok := request.Params.Arguments["device_id"].(string)
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

	s.AddTool(turnOnTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, ok := request.Params.Arguments["device_id"].(string)
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

	s.AddTool(turnOffTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, ok := request.Params.Arguments["device_id"].(string)
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

	s.AddTool(setLevelTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, ok := request.Params.Arguments["device_id"].(string)
		if !ok {
			return mcp.NewToolResultError("device_id must be a string"), nil
		}

		level, ok := request.Params.Arguments["level"].(float64)
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

	s.AddTool(customCommandTool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		deviceID, ok := request.Params.Arguments["device_id"].(string)
		if !ok {
			return mcp.NewToolResultError("device_id must be a string"), nil
		}

		command, ok := request.Params.Arguments["command"].(string)
		if !ok {
			return mcp.NewToolResultError("command must be a string"), nil
		}

		value, hasValue := request.Params.Arguments["value"].(string)

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

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	log.Printf("Starting Hubitat MCP Server v%s", serverVersion)
	log.Printf("Hubitat API: %s", config.BaseURL)
	log.Printf("Listening on port: %s", config.Port)

	mcpServer := setupMCPServer()

	if err := mcpServer.ServeStdio(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
