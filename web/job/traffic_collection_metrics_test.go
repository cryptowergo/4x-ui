package job

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTrafficCollectionFailureAndRecovery(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "stats.prom")
	state := trafficCollectionState{}
	state.record(time.Unix(100, 0), "")
	state.record(time.Unix(110, 0), "fetch")
	if err := state.write(path); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"v_agent_xray_stats_collection_success 0\n",
		"v_agent_xray_stats_collection_last_success_timestamp_seconds 100\n",
		"v_agent_xray_stats_collection_last_attempt_timestamp_seconds 110\n",
		"v_agent_xray_stats_collection_failures_total 1\n",
	} {
		if !strings.Contains(string(data), expected) {
			t.Fatalf("missing metric %q", expected)
		}
	}
	state.record(time.Unix(120, 0), "")
	if err := state.write(path); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(path)
	if !strings.Contains(string(data), "v_agent_xray_stats_collection_success 1\n") ||
		!strings.Contains(string(data), "v_agent_xray_stats_collection_last_success_timestamp_seconds 120\n") {
		t.Fatal("recovery did not replace failed snapshot")
	}
	files, _ := os.ReadDir(filepath.Dir(path))
	if len(files) != 1 {
		t.Fatal("temporary metrics files leaked")
	}
}

func TestTrafficCollectionMetricsOptionalAndFailVisible(t *testing.T) {
	t.Parallel()
	state := trafficCollectionState{}
	if err := state.write(""); err != nil {
		t.Fatal(err)
	}
	if err := state.write(filepath.Join(t.TempDir(), "missing", "stats.prom")); err == nil {
		t.Fatal("missing output directory must report failure")
	}
}
