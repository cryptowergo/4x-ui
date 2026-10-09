package job

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// trafficCollectionState describes collection, not Xray data-plane availability.
// The optional textfile contains no peer labels or configuration values.
type trafficCollectionState struct {
	success     bool
	lastAttempt time.Time
	lastSuccess time.Time
	failures    uint64
}

func (s *trafficCollectionState) record(now time.Time, failure string) {
	s.lastAttempt = now
	s.success = failure == ""
	if s.success {
		s.lastSuccess = now
	} else {
		s.failures++
	}
}

func unixSeconds(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.Unix()
}

func (s *trafficCollectionState) write(path string) error {
	if path == "" {
		return nil
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".xui-traffic-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	success := 0
	if s.success {
		success = 1
	}
	_, err = fmt.Fprintf(file, `# HELP v_agent_xray_stats_collection_success Last complete Xray traffic collection and database update succeeded.
# TYPE v_agent_xray_stats_collection_success gauge
v_agent_xray_stats_collection_success %d
# TYPE v_agent_xray_stats_collection_last_attempt_timestamp_seconds gauge
v_agent_xray_stats_collection_last_attempt_timestamp_seconds %d
# TYPE v_agent_xray_stats_collection_last_success_timestamp_seconds gauge
v_agent_xray_stats_collection_last_success_timestamp_seconds %d
# TYPE v_agent_xray_stats_collection_failures_total counter
v_agent_xray_stats_collection_failures_total %d
`, success, unixSeconds(s.lastAttempt), unixSeconds(s.lastSuccess), s.failures)
	if err != nil {
		return err
	}
	if err = file.Chmod(0o644); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), path)
}
