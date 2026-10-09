//go:build integration

package xray

import (
	"context"
	"net"
	"strings"
	"testing"

	"github.com/mhsanaei/3x-ui/v2/logger"
	"github.com/op/go-logging"
	statsService "github.com/xtls/xray-core/app/stats/command"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type largeStatsServer struct {
	statsService.UnimplementedStatsServiceServer
	size int
}

func (s *largeStatsServer) QueryStats(context.Context, *statsService.QueryStatsRequest) (*statsService.QueryStatsResponse, error) {
	return &statsService.QueryStatsResponse{Stat: []*statsService.Stat{{
		Name: "inbound>>>" + strings.Repeat("a", s.size) + ">>>traffic>>>uplink", Value: 7,
	}}}, nil
}

func TestStatsReceiveBudget(t *testing.T) {
	logger.InitLogger(logging.ERROR)
	for _, tc := range []struct {
		name     string
		size     int
		wantCode codes.Code
	}{
		{"above_default_limit", 5 << 20, codes.OK},
		{"bounded_limit", 33 << 20, codes.ResourceExhausted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			if err != nil {
				t.Fatal(err)
			}
			server := grpc.NewServer()
			statsService.RegisterStatsServiceServer(server, &largeStatsServer{size: tc.size})
			go server.Serve(listener)
			t.Cleanup(server.Stop)
			client := &XrayAPI{}
			if err := client.Init(listener.Addr().(*net.TCPAddr).Port); err != nil {
				t.Fatal(err)
			}
			t.Cleanup(client.Close)
			traffic, _, err := client.GetTraffic(false)
			if status.Code(err) != tc.wantCode {
				t.Fatalf("got %v, want %v", err, tc.wantCode)
			}
			if err == nil && (len(traffic) != 1 || traffic[0].Up != 7) {
				t.Fatal("large response traffic lost")
			}
		})
	}
}
