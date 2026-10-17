package executor

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	goplug "github.com/hashicorp/go-plugin"
	"github.com/jmylchreest/tinct/internal/plugin/protocol"
	"github.com/jmylchreest/tinct/pkg/plugin"
)

// TestMain can also run as the deliberately slow RPC plugin subprocess.
func TestMain(m *testing.M) {
	if os.Getenv("TINCT_TEST_RPC_HELPER") == "startup" {
		time.Sleep(time.Minute)
		return
	}
	if os.Getenv("TINCT_TEST_RPC_HELPER") == "1" {
		goplug.Serve(&goplug.ServeConfig{
			HandshakeConfig: protocol.Handshake,
			Plugins:         map[string]goplug.Plugin{"output": &plugin.OutputPluginRPC{Impl: &blockedOutput{}}},
		})
		return
	}
	os.Exit(m.Run())
}

type blockedOutput struct{ plugin.OutputPlugin }

func (*blockedOutput) Generate(context.Context, plugin.PaletteData) (map[string][]byte, error) {
	time.Sleep(time.Minute)
	return nil, nil
}

func TestRPCCancellationTerminatesProcess(t *testing.T) {
	t.Setenv("TINCT_TEST_RPC_HELPER", "1")
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	e := &PluginExecutor{path: path, protocolType: protocol.PluginTypeGoPlugin}
	defer e.Close()
	// Complete the handshake first, so the deadline exercises remote execution.
	if _, err := e.getOutputRPCClient(context.Background()); err != nil {
		t.Fatal(err)
	}
	process := e.client
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := e.ExecuteOutput(ctx, plugin.PaletteData{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("cancellation took %v", elapsed)
	}
	if !process.Exited() {
		t.Fatal("plugin process survived cancellation")
	}
	if _, err := e.getOutputRPCClient(context.Background()); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancelled session restarted: %v", err)
	}
	if e.client != nil || e.rpcClient != nil {
		t.Fatal("cancelled session retained")
	}
}

func TestRPCCancellationDuringHandshake(t *testing.T) {
	t.Setenv("TINCT_TEST_RPC_HELPER", "startup")
	path, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	e := &PluginExecutor{path: path, protocolType: protocol.PluginTypeGoPlugin}
	defer e.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	if _, err := e.ExecuteOutput(ctx, plugin.PaletteData{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("error = %v", err)
	}
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Fatalf("handshake cancellation took %v", elapsed)
	}
	if e.client != nil {
		t.Fatal("failed handshake retained a client")
	}
}
