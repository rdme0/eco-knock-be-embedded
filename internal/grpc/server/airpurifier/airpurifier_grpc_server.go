package airpurifier

import (
	"context"
	"errors"
	"time"

	"eco-knock-be-embedded/internal/airpurifier/xiaomi/constant"
	airservice "eco-knock-be-embedded/internal/airpurifier/xiaomi/service"
	"eco-knock-be-embedded/internal/common/apperror"
	airpurifierpb "eco-knock-be-embedded/internal/grpc/pb/airpurifier/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

var ErrAirPurifierServiceRequired = errors.New("공기청정기 서비스가 필요합니다")

type GRPCServer struct {
	airpurifierpb.UnimplementedAirPurifierServiceServer
	airPurifierService *airservice.XiaomiAirPurifierService
}

func NewGRPCServer(service *airservice.XiaomiAirPurifierService) (*GRPCServer, error) {
	if service == nil {
		return nil, ErrAirPurifierServiceRequired
	}

	return &GRPCServer{
		airPurifierService: service,
	}, nil
}

func (server *GRPCServer) GetCurrentAirPurifier(
	ctx context.Context,
	_ *airpurifierpb.GetCurrentAirPurifierRequest,
) (*airpurifierpb.GetCurrentAirPurifierResponse, error) {
	return server.currentAirPurifier(ctx)
}

func (server *GRPCServer) SetAirPurifierPower(
	ctx context.Context,
	request *airpurifierpb.SetAirPurifierPowerRequest,
) (*airpurifierpb.SetAirPurifierControlResponse, error) {
	if err := server.airPurifierService.SetPower(ctx, request.GetOn()); err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.AirPurifierControlFailed, err))
	}

	return server.currentControlResponse(ctx)
}

func (server *GRPCServer) SetAirPurifierMode(
	ctx context.Context,
	request *airpurifierpb.SetAirPurifierModeRequest,
) (*airpurifierpb.SetAirPurifierControlResponse, error) {
	mode, ok := parseOperationMode(request.GetMode())
	if !ok {
		return nil, apperror.ToGRPCError(apperror.New(apperror.AirPurifierInvalidCommand, nil))
	}

	if err := server.airPurifierService.SetMode(ctx, mode); err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.AirPurifierControlFailed, err))
	}

	return server.currentControlResponse(ctx)
}

func (server *GRPCServer) SetAirPurifierFavoriteLevel(
	ctx context.Context,
	request *airpurifierpb.SetAirPurifierFavoriteLevelRequest,
) (*airpurifierpb.SetAirPurifierControlResponse, error) {
	level := request.GetLevel()
	if level < 0 || level > 17 {
		return nil, apperror.ToGRPCError(apperror.New(apperror.AirPurifierInvalidCommand, nil))
	}

	if err := server.airPurifierService.SetFavoriteLevel(ctx, int(level)); err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.AirPurifierControlFailed, err))
	}

	return server.currentControlResponse(ctx)
}

func (server *GRPCServer) currentControlResponse(ctx context.Context) (*airpurifierpb.SetAirPurifierControlResponse, error) {
	current, err := server.currentAirPurifier(ctx)
	if err != nil {
		return nil, err
	}

	return &airpurifierpb.SetAirPurifierControlResponse{
		Current: current,
	}, nil
}

func (server *GRPCServer) currentAirPurifier(ctx context.Context) (*airpurifierpb.GetCurrentAirPurifierResponse, error) {
	result, err := server.airPurifierService.Status(ctx)
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.AirPurifierReadFailed, err))
	}

	response := &airpurifierpb.GetCurrentAirPurifierResponse{
		Power:               result.Power,
		IsOn:                result.IsOn,
		Aqi:                 int32(result.AQI),
		AverageAqi:          int32(result.AverageAQI),
		Humidity:            int32(result.Humidity),
		Mode:                string(result.Mode),
		FavoriteLevel:       int32(result.FavoriteLevel),
		FilterLifeRemaining: int32(result.FilterLifeRemaining),
		FilterHoursUsed:     int32(result.FilterHoursUsed),
		MotorSpeed:          int32(result.MotorSpeed),
		PurifyVolume:        int32(result.PurifyVolume),
		Led:                 result.LED,
		ChildLock:           result.ChildLock,
		MeasuredAtUnixMs:    time.Now().UnixMilli(),
	}

	if result.Temperature != nil {
		response.TemperatureC = wrapperspb.Double(*result.Temperature)
	}

	if result.LEDBrightness != nil {
		response.LedBrightness = wrapperspb.Int32(int32(*result.LEDBrightness))
	}

	if result.Buzzer != nil {
		response.Buzzer = wrapperspb.Bool(*result.Buzzer)
	}

	return response, nil
}

func parseOperationMode(value string) (constant.OperationMode, bool) {
	switch constant.OperationMode(value) {
	case constant.OperationModeAuto,
		constant.OperationModeSilent,
		constant.OperationModeFavorite,
		constant.OperationModeIdle,
		constant.OperationModeMedium,
		constant.OperationModeHigh,
		constant.OperationModeStrong,
		constant.OperationModeLow:
		return constant.OperationMode(value), true
	default:
		return "", false
	}
}
