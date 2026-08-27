package plugin

import (
	"context"
	"errors"
	"image/color"
	"net"
	"net/rpc"
	"strings"
	"testing"
	"time"
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

func (m *mockInputPlugin) WallpaperRawPath() string {
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

func (m *mockOutputPlugin) PreExecute(_ context.Context) (skip bool, reason string, err error) {
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

	rpcPlugin := &InputPluginRPC{Impl: mock}

	t.Run("Server", func(t *testing.T) {
		server, err := rpcPlugin.Server(nil)
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
		client, err := rpcPlugin.Client(nil, nil)
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

	rpcPlugin := &OutputPluginRPC{Impl: mock}

	t.Run("Server", func(t *testing.T) {
		server, err := rpcPlugin.Server(nil)
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
		client, err := rpcPlugin.Client(nil, nil)
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

// --- Configure -------------------------------------------------------------

type configurableOutput struct {
	OutputPlugin
	got   ConfigureRequest
	calls int
	err   error
}

func (c *configurableOutput) Configure(req ConfigureRequest) error {
	c.got = req
	c.calls++
	return c.err
}

// A plugin implementing Configurable receives the host's args.
func TestOutputRPCServerConfigure(t *testing.T) {
	impl := &configurableOutput{}
	srv := &OutputPluginRPCServer{Impl: impl}

	var resp string
	req := ConfigureRequest{Args: map[string]any{"output-dir": "/tmp/x"}, DryRun: true, Verbose: true}
	if err := srv.Configure(req, &resp); err != nil {
		t.Fatalf("Configure returned error: %v", err)
	}
	if resp != "" {
		t.Errorf("resp = %q, want empty", resp)
	}
	if impl.calls != 1 {
		t.Fatalf("Configure called %d times, want 1", impl.calls)
	}
	if got := impl.got.Args["output-dir"]; got != "/tmp/x" {
		t.Errorf("Args[output-dir] = %v, want /tmp/x", got)
	}
	if !impl.got.DryRun || !impl.got.Verbose {
		t.Errorf("DryRun/Verbose not propagated: %+v", impl.got)
	}
}

// A plugin error is reported in the response rather than as a transport
// error, so the host can decide it is non-fatal.
func TestOutputRPCServerConfigureReportsPluginError(t *testing.T) {
	impl := &configurableOutput{err: errors.New("bad output dir")}
	srv := &OutputPluginRPCServer{Impl: impl}

	var resp string
	if err := srv.Configure(ConfigureRequest{}, &resp); err != nil {
		t.Fatalf("Configure returned transport error: %v", err)
	}
	if resp != "bad output dir" {
		t.Errorf("resp = %q, want the plugin's error text", resp)
	}
}

// Plugins that predate Configurable must not break: the server treats
// them as a successful no-op.
func TestOutputRPCServerConfigureNonConfigurableIsNoOp(t *testing.T) {
	srv := &OutputPluginRPCServer{Impl: &nonConfigurableOutput{}}

	var resp string
	if err := srv.Configure(ConfigureRequest{Args: map[string]any{"a": 1}}, &resp); err != nil {
		t.Fatalf("Configure returned error: %v", err)
	}
	if resp != "" {
		t.Errorf("resp = %q, want empty", resp)
	}
}

type nonConfigurableOutput struct{ OutputPlugin }

// --- bounded RPC -----------------------------------------------------------

// slowService has one method that answers and one that never does, so we
// can exercise both branches of call().
type slowService struct{ release chan struct{} }

func (s *slowService) Echo(arg string, reply *string) error {
	*reply = arg
	return nil
}

func (s *slowService) Hang(_ string, reply *string) error {
	<-s.release // blocks until the test tears the fixture down
	*reply = "never"
	return nil
}

// newLoopbackClient serves slowService over an in-memory pipe and
// returns a connected rpc.Client. The service itself is released by the
// cleanup below, so callers never need a handle on it.
func newLoopbackClient(t *testing.T) *rpc.Client {
	t.Helper()

	svc := &slowService{release: make(chan struct{})}
	srv := rpc.NewServer()
	if err := srv.RegisterName("Plugin", svc); err != nil {
		t.Fatalf("RegisterName: %v", err)
	}

	serverConn, clientConn := net.Pipe()
	go srv.ServeConn(serverConn)

	client := rpc.NewClient(clientConn)
	t.Cleanup(func() {
		close(svc.release)
		_ = client.Close()
	})
	return client
}

// A responsive method returns its reply and no error.
func TestCallReturnsReply(t *testing.T) {
	client := newLoopbackClient(t)

	var got string
	if err := call(client, "Plugin.Echo", "hello", &got, callTimeout); err != nil {
		t.Fatalf("call returned error: %v", err)
	}
	if got != "hello" {
		t.Errorf("reply = %q, want %q", got, "hello")
	}
}

// The regression this exists for: a method that never replies must fail
// on a deadline rather than block the host forever.
func TestCallTimesOutInsteadOfHanging(t *testing.T) {
	client := newLoopbackClient(t)

	done := make(chan error, 1)
	go func() {
		var got string
		done <- call(client, "Plugin.Hang", "x", &got, 100*time.Millisecond)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("call succeeded; want a timeout error")
		}
		for _, want := range []string{"Plugin.Hang", "did not respond", "protocol version"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("error %q, want it to mention %q", err.Error(), want)
			}
		}
	case <-time.After(10 * time.Second):
		t.Fatal("call did not return; the deadline is not being enforced")
	}
}

// A missing method must surface as an error call() passes through, so
// isMissingMethodErr can classify it for the optional-RPC fallbacks.
func TestCallSurfacesMissingMethod(t *testing.T) {
	client := newLoopbackClient(t)

	var got string
	err := call(client, "Plugin.NoSuchMethod", "x", &got, callTimeout)
	if err == nil {
		t.Fatal("call succeeded on an unknown method")
	}
	if !isMissingMethodErr(err) {
		t.Errorf("error %q not recognised as a missing method", err.Error())
	}
}
