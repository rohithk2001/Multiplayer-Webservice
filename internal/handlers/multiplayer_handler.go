package handlers

import (
	"context"
	"log"

	// "time"

	"multiplayer-webservice/internal/cache"
	"multiplayer-webservice/internal/logic"
	"multiplayer-webservice/internal/proto"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MultiplayerService struct {
	proto.UnimplementedMultiplayerServiceServer
	Collection *mongo.Collection
	RedisCache *cache.RedisCache
}

// NewMultiplayerService initializes a new instance of MultiplayerService.
func NewMultiplayerService(collection *mongo.Collection, redisCache *cache.RedisCache) *MultiplayerService {
	return &MultiplayerService{
		Collection: collection,
		RedisCache: redisCache,
	}
}

// GetModeUsage fetches mode usage details.
func (s *MultiplayerService) GetModeUsage(ctx context.Context, req *proto.ModeUsageRequest) (*proto.ModeUsageResponse, error) {
	modes, err := logic.GetModeUsageLogic(ctx, s.Collection, s.RedisCache)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch game modes: %v", err)
	}

	return &proto.ModeUsageResponse{Modes: modes}, nil
}

// JoinMode adds a player to a mode.
func (s *MultiplayerService) JoinMode(ctx context.Context, req *proto.JoinModeRequest) (*proto.JoinModeResponse, error) {
	err := logic.JoinModeLogic(ctx, s.Collection, s.RedisCache, req.GetModeName(), req.GetPlayerId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to join mode: %v", err)
	}
	return &proto.JoinModeResponse{Message: "Player added successfully"}, nil
}

// LeaveMode removes a player from a mode.
func (s *MultiplayerService) LeaveMode(ctx context.Context, req *proto.LeaveModeRequest) (*proto.LeaveModeResponse, error) {
	err := logic.LeaveModeLogic(ctx, s.Collection, s.RedisCache, req.GetModeName(), req.GetPlayerId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to leave mode: %v", err)
	}
	return &proto.LeaveModeResponse{Message: "Player removed successfully"}, nil
}

// GetTotalActiveUsers fetches total active users
func getTotalActiveUsers(c *gin.Context) {
    conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
    if err != nil {
        log.Printf("gRPC connection error: %v", err)
        c.JSON(500, gin.H{"message": "Failed to connect to gRPC server"})
        return
    }
    defer conn.Close()

    client := proto.NewMultiplayerServiceClient(conn)
    resp, err := client.GetTotalActiveUsers(context.Background(), &proto.TotalActiveUsersRequest{})
    if err != nil {
        log.Printf("gRPC request error: %v", err)
        c.JSON(500, gin.H{"message": "Failed to fetch total active users"})
        return
    }

    c.JSON(200, gin.H{"totalActiveUsers": resp.TotalActiveUsers})
}

// GetModeDetails fetches mode details
func (s *MultiplayerService) GetModeDetails(ctx context.Context, req *proto.ModeDetailsRequest) (*proto.ModeDetailsResponse, error) {
	modeDetails, err := logic.GetModeDetailsLogic(ctx, s.Collection, s.RedisCache, req.GetModeName())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Mode not found: %v", err)
	}
	return modeDetails, nil
}

// GetActiveUsersByAreaCode fetches active users by area code
func (s *MultiplayerService) GetActiveUsersByAreaCode(ctx context.Context, req *proto.ActiveUsersByAreaCodeRequest) (*proto.ActiveUsersByAreaCodeResponse, error) {
	totalUsers, err := logic.GetActiveUsersByAreaCodeLogic(ctx, s.Collection, s.RedisCache, req.GetAreaCode())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch active users by area code: %v", err)
	}
	return &proto.ActiveUsersByAreaCodeResponse{TotalActiveUsers: totalUsers}, nil
}

// GetGameModeStats fetches game mode stats
func (s *MultiplayerService) GetGameModeStats(ctx context.Context, req *proto.GameModeStatsRequest) (*proto.GameModeStatsResponse, error) {
	stats, err := logic.GetGameModeStatsLogic(ctx, s.Collection, s.RedisCache)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to fetch game mode stats: %v", err)
	}
	return stats, nil
}

// GetPlayers fetches players in a mode
func (s *MultiplayerService) GetPlayers(ctx context.Context, req *proto.GetPlayersRequest) (*proto.GetPlayersResponse, error) {
	players, err := logic.GetPlayersLogic(ctx, s.Collection, s.RedisCache, req.GetModeName())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "Mode not found: %v", err)
	}
	return &proto.GetPlayersResponse{Players: players}, nil
}

// UpdateGameState modifies the game state of a mode
func (s *MultiplayerService) UpdateGameState(ctx context.Context, req *proto.UpdateGameStateRequest) (*proto.UpdateGameStateResponse, error) {
	err := logic.UpdateGameStateLogic(ctx, s.Collection, s.RedisCache, req.GetModeName(), req.GetGameState())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Failed to update game state: %v", err)
	}
	return &proto.UpdateGameStateResponse{Message: "Game state updated successfully"}, nil
}
