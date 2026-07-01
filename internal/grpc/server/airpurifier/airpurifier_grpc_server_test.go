package airpurifier

import (
	"context"
	"testing"
	"time"

	"eco-knock-be-embedded/internal/airpurifier/xiaomi/client"
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	airservice "eco-knock-be-embedded/internal/airpurifier/xiaomi/service"
	airpurifierpb "eco-knock-be-embedded/internal/grpc/pb/airpurifier/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestSetAirPurifierPowerReturnsCurrentStatus(t *testing.T) {
	t.Parallel()

	server := newStubGRPCServer(t)

	response, err := server.SetAirPurifierPower(context.Background(), &airpurifierpb.SetAirPurifierPowerRequest{
		On: false,
	})
	if err != nil {
		t.Fatalf("unexpected set power error: %v", err)
	}

	if response.GetCurrent().GetIsOn() {
		t.Fatal("expected current purifier status to be off")
	}
}

func TestSetAirPurifierModeReturnsCurrentStatus(t *testing.T) {
	t.Parallel()

	server := newStubGRPCServer(t)

	response, err := server.SetAirPurifierMode(context.Background(), &airpurifierpb.SetAirPurifierModeRequest{
		Mode: "favorite",
	})
	if err != nil {
		t.Fatalf("unexpected set mode error: %v", err)
	}

	if response.GetCurrent().GetMode() != "favorite" {
		t.Fatalf("expected current mode favorite, got %q", response.GetCurrent().GetMode())
	}
}

func TestSetAirPurifierFavoriteLevelReturnsCurrentStatus(t *testing.T) {
	t.Parallel()

	server := newStubGRPCServer(t)

	response, err := server.SetAirPurifierFavoriteLevel(context.Background(), &airpurifierpb.SetAirPurifierFavoriteLevelRequest{
		Level: 7,
	})
	if err != nil {
		t.Fatalf("unexpected set favorite level error: %v", err)
	}

	if response.GetCurrent().GetFavoriteLevel() != 7 {
		t.Fatalf("expected current favorite level 7, got %d", response.GetCurrent().GetFavoriteLevel())
	}
}

func TestSetAirPurifierFavoriteLevelRejectsInvalidLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level int32
	}{
		{name: "below minimum", level: -1},
		{name: "above maximum", level: 18},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := newStubGRPCServer(t)

			_, err := server.SetAirPurifierFavoriteLevel(context.Background(), &airpurifierpb.SetAirPurifierFavoriteLevelRequest{
				Level: tt.level,
			})
			if status.Code(err) != codes.InvalidArgument {
				t.Fatalf("expected InvalidArgument, got %v", err)
			}
		})
	}
}

func TestSetAirPurifierModeRejectsInvalidMode(t *testing.T) {
	t.Parallel()

	server := newStubGRPCServer(t)

	_, err := server.SetAirPurifierMode(context.Background(), &airpurifierpb.SetAirPurifierModeRequest{
		Mode: "turbo",
	})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func newStubGRPCServer(t *testing.T) *GRPCServer {
	t.Helper()

	conf, err := airconfig.New("127.0.0.1:54321", "00112233445566778899aabbccddeeff", time.Second)
	if err != nil {
		t.Fatalf("unexpected new config error: %v", err)
	}

	service, err := airservice.NewWithClientMode(conf, client.ModeStub, nil)
	if err != nil {
		t.Fatalf("unexpected new service error: %v", err)
	}

	server, err := NewGRPCServer(service)
	if err != nil {
		t.Fatalf("unexpected new grpc server error: %v", err)
	}

	return server
}
