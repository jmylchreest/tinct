package plugin

import (
	"context"
	"image/color"
	"testing"
)

// Mock implementations for testing.
type mockInputPlugin struct {
	colors      []color.Color
	metadata    PluginInfo
	wallpaper   string
	flagHelp    []FlagHelp
	generateErr error
}

func (m *mockInputPlugin) Generate(_ context.Context, _ InputOptions) ([]color.Color, error) {
	if m.generateErr != nil {
		return nil, m.generateErr
	}
	return m.colors, nil
}

func (m *mockInputPlugin) GetMetadata() PluginInfo {
	return m.metadata
}

func (m *mockInputPlugin) WallpaperPath() string {
	return m.wallpaper
}

func (m *mockInputPlugin) GetFlagHelp() []FlagHelp {
	return m.flagHelp
}

type mockOutputPlugin struct {
	files       map[string][]byte
	skipPreExec bool
	skipReason  string
	metadata    PluginInfo
	flagHelp    []FlagHelp
	generateErr error
	preExecErr  error
	postExecErr error
}

func (m *mockOutputPlugin) Generate(_ context.Context, _ PaletteData) (map[string][]byte, error) {
	if m.generateErr != nil {
		return nil, m.generateErr
	}
	return m.files, nil
}

func (m *mockOutputPlugin) PreExecute(_ context.Context) (bool, string, error) {
	if m.preExecErr != nil {
		return false, "", m.preExecErr
	}
	return m.skipPreExec, m.skipReason, nil
}

func (m *mockOutputPlugin) PostExecute(_ context.Context, _ []string) error {
	return m.postExecErr
}

func (m *mockOutputPlugin) GetMetadata() PluginInfo {
	return m.metadata
}

func (m *mockOutputPlugin) GetFlagHelp() []FlagHelp {
	return m.flagHelp
}

// TestInputPluginRPC tests the input plugin RPC wrapper.
func TestInputPluginRPC(t *testing.T) {
	mock := &mockInputPlugin{
		colors: []color.Color{
			color.RGBA{R: 255, G: 0, B: 0, A: 255},
			color.RGBA{R: 0, G: 255, B: 0, A: 255},
			color.RGBA{R: 0, G: 0, B: 255, A: 255},
		},
		metadata: PluginInfo{
			Name:            "test-input",
			Type:            "input",
			Version:         "1.0.0",
			ProtocolVersion: ProtocolVersion,
			Description:     "Test input plugin",
			PluginProtocol:  string(PluginTypeGoPlugin),
		},
		wallpaper: "/path/to/wallpaper.jpg",
		flagHelp: []FlagHelp{
			{Name: "test-flag", Type: "string", Default: "default", Description: "Test flag", Required: false},
		},
	}

	rpc := &InputPluginRPC{Impl: mock}

	t.Run("Server", func(t *testing.T) {
		server, err := rpc.Server(nil)
		if err != nil {
			t.Fatalf("Server() error = %v", err)
		}
		if server == nil {
			t.Fatal("Server() returned nil server")
		}

		rpcServer, ok := server.(*InputPluginRPCServer)
		if !ok {
			t.Fatal("Server() returned wrong type")
		}
		if rpcServer.Impl != mock {
			t.Fatal("Server() impl not set correctly")
		}
	})

	t.Run("Client", func(t *testing.T) {
		client, err := rpc.Client(nil, nil)
		if err != nil {
			t.Fatalf("Client() error = %v", err)
		}
		if client == nil {
			t.Fatal("Client() returned nil client")
		}
	})
}

// TestOutputPluginRPC tests the output plugin RPC wrapper.
func TestOutputPluginRPC(t *testing.T) {
	mock := &mockOutputPlugin{
		files: map[string][]byte{
			"theme.conf": []byte("color=#ff0000"),
		},
		metadata: PluginInfo{
			Name:            "test-output",
			Type:            "output",
			Version:         "1.0.0",
			ProtocolVersion: ProtocolVersion,
			Description:     "Test output plugin",
			PluginProtocol:  string(PluginTypeGoPlugin),
		},
		flagHelp: []FlagHelp{
			{Name: "output-dir", Type: "string", Default: "", Description: "Output directory", Required: false},
		},
	}

	rpc := &OutputPluginRPC{Impl: mock}

	t.Run("Server", func(t *testing.T) {
		server, err := rpc.Server(nil)
		if err != nil {
			t.Fatalf("Server() error = %v", err)
		}
		if server == nil {
			t.Fatal("Server() returned nil server")
		}

		rpcServer, ok := server.(*OutputPluginRPCServer)
		if !ok {
			t.Fatal("Server() returned wrong type")
		}
		if rpcServer.Impl != mock {
			t.Fatal("Server() impl not set correctly")
		}
	})

	t.Run("Client", func(t *testing.T) {
		client, err := rpc.Client(nil, nil)
		if err != nil {
			t.Fatalf("Client() error = %v", err)
		}
		if client == nil {
			t.Fatal("Client() returned nil client")
		}
	})
}

// TestInputPluginRPCServer tests the RPC server methods.
func TestInputPluginRPCServer(t *testing.T) {
	mock := &mockInputPlugin{
		colors: []color.Color{
			color.RGBA{R: 128, G: 128, B: 128, A: 255},
		},
		metadata: PluginInfo{
			Name:            "test",
			ProtocolVersion: ProtocolVersion,
		},
		wallpaper: "/test/wallpaper.png",
		flagHelp: []FlagHelp{
			{Name: "flag1", Type: "string"},
		},
	}

	server := &InputPluginRPCServer{Impl: mock}

	t.Run("Generate", func(t *testing.T) {
		opts := InputOptions{Verbose: true}
		var resp []byte
		err := server.Generate(opts, &resp)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if len(resp) == 0 {
			t.Fatal("Generate() returned empty response")
		}
	})

	t.Run("GetMetadata", func(t *testing.T) {
		var resp PluginInfo
		err := server.GetMetadata(nil, &resp)
		if err != nil {
			t.Fatalf("GetMetadata() error = %v", err)
		}
		if resp.Name != "test" {
			t.Errorf("GetMetadata() name = %q, want %q", resp.Name, "test")
		}
	})

	t.Run("WallpaperPath", func(t *testing.T) {
		var resp string
		err := server.WallpaperPath(nil, &resp)
		if err != nil {
			t.Fatalf("WallpaperPath() error = %v", err)
		}
		if resp != "/test/wallpaper.png" {
			t.Errorf("WallpaperPath() = %q, want %q", resp, "/test/wallpaper.png")
		}
	})

	t.Run("GetFlagHelp", func(t *testing.T) {
		var resp []FlagHelp
		err := server.GetFlagHelp(nil, &resp)
		if err != nil {
			t.Fatalf("GetFlagHelp() error = %v", err)
		}
		if len(resp) != 1 {
			t.Fatalf("GetFlagHelp() returned %d flags, want 1", len(resp))
		}
		if resp[0].Name != "flag1" {
			t.Errorf("GetFlagHelp()[0].Name = %q, want %q", resp[0].Name, "flag1")
		}
	})
}

// TestOutputPluginRPCServer tests the output RPC server methods.
func TestOutputPluginRPCServer(t *testing.T) {
	mock := &mockOutputPlugin{
		files: map[string][]byte{
			"config.ini": []byte("setting=value"),
		},
		metadata: PluginInfo{
			Name: "test-output",
		},
		flagHelp: []FlagHelp{
			{Name: "output-flag", Type: "bool"},
		},
	}

	server := &OutputPluginRPCServer{Impl: mock}

	t.Run("Generate", func(t *testing.T) {
		palette := PaletteData{
			Colours:    make(map[string]CategorisedColour),
			AllColours: []CategorisedColour{},
			ThemeType:  "dark",
		}
		var resp map[string][]byte
		err := server.Generate(palette, &resp)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if len(resp) == 0 {
			t.Fatal("Generate() returned empty files map")
		}
		if _, ok := resp["config.ini"]; !ok {
			t.Error("Generate() missing expected file 'config.ini'")
		}
	})

	t.Run("PreExecute", func(t *testing.T) {
		var resp struct {
			Skip   bool
			Reason string
			Error  string
		}
		err := server.PreExecute(nil, &resp)
		if err != nil {
			t.Fatalf("PreExecute() error = %v", err)
		}
	})

	t.Run("PostExecute", func(t *testing.T) {
		files := []string{"file1.txt", "file2.txt"}
		var resp string
		err := server.PostExecute(files, &resp)
		if err != nil {
			t.Fatalf("PostExecute() error = %v", err)
		}
	})

	t.Run("GetMetadata", func(t *testing.T) {
		var resp PluginInfo
		err := server.GetMetadata(nil, &resp)
		if err != nil {
			t.Fatalf("GetMetadata() error = %v", err)
		}
		if resp.Name != "test-output" {
			t.Errorf("GetMetadata() name = %q, want %q", resp.Name, "test-output")
		}
	})

	t.Run("GetFlagHelp", func(t *testing.T) {
		var resp []FlagHelp
		err := server.GetFlagHelp(nil, &resp)
		if err != nil {
			t.Fatalf("GetFlagHelp() error = %v", err)
		}
		if len(resp) != 1 {
			t.Fatalf("GetFlagHelp() returned %d flags, want 1", len(resp))
		}
	})
}

// TestRPCError tests the RPCError type.
func TestRPCError(t *testing.T) {
	err := &RPCError{Message: "test error"}
	if err.Error() != "test error" {
		t.Errorf("RPCError.Error() = %q, want %q", err.Error(), "test error")
	}
}

// TestPluginInfo tests PluginInfo structure.
func TestPluginInfo(t *testing.T) {
	info := PluginInfo{
		Name:            "test-plugin",
		Type:            "input",
		Version:         "2.0.0",
		ProtocolVersion: "0.0.1",
		Description:     "A test plugin",
		PluginProtocol:  "go-plugin",
	}

	if info.Name != "test-plugin" {
		t.Errorf("Name = %q, want %q", info.Name, "test-plugin")
	}
	if info.Type != "input" {
		t.Errorf("Type = %q, want %q", info.Type, "input")
	}
	if info.Version != "2.0.0" {
		t.Errorf("Version = %q, want %q", info.Version, "2.0.0")
	}
}

// TestFlagHelp tests FlagHelp structure.
func TestFlagHelp(t *testing.T) {
	flag := FlagHelp{
		Name:        "test-flag",
		Shorthand:   "t",
		Type:        "string",
		Default:     "default-value",
		Description: "Test flag description",
		Required:    true,
	}

	if flag.Name != "test-flag" {
		t.Errorf("Name = %q, want %q", flag.Name, "test-flag")
	}
	if flag.Shorthand != "t" {
		t.Errorf("Shorthand = %q, want %q", flag.Shorthand, "t")
	}
	if !flag.Required {
		t.Error("Required = false, want true")
	}
}
