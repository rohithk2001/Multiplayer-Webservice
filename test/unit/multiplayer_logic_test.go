package unit

import (
	"context"
	"testing"
	"time"

	"multiplayer-webservice/internal/cache"
	"multiplayer-webservice/internal/logic" // Adjust the import path as necessary

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) *mongo.Collection {
    // Set up MongoDB connection
    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        t.Fatalf("Failed to connect to MongoDB: %v", err)
    }

    // Create a test database and collection
    db := client.Database("testdb")
    collection := db.Collection("testcollection")

    // Clean up the collection before tests (drop if it exists)
    if err := collection.Drop(context.TODO()); err != nil {
        t.Fatalf("Failed to drop collection: %v", err)
    }

    return collection
}

func TestGetModeUsageLogic(t *testing.T) {
    collection := setupTestDB(t)
    ctx := context.Background()

    // Initialize Redis cache
    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }

    // Insert test data
    testMode := logic.ModeUsage{
        ModeName:    "TestMode",
        ActiveUsers: 5,
        AreaCode:    "123",
        Players:     []string{"player1", "player2"},
        GameState:   "active",
        LastUpdated: time.Now(),
    }
    _, err = collection.InsertOne(ctx, testMode)
    if err != nil {
        t.Fatalf("Failed to insert test data: %v", err)
    }

    // Call the logic function with the Redis cache
    modes, err := logic.GetModeUsageLogic(ctx, collection, redisCache)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Assertions
    if len(modes) != 1 {
        t.Fatalf("expected 1 mode, got %d", len(modes))
    }
    if modes[0].ModeName != "TestMode" {
        t.Fatalf("expected mode name 'TestMode', got %s", modes[0].ModeName)
    }
    if modes[0].AreaCode != "123" {
        t.Fatalf("expected area code '123', got %s", modes[0].AreaCode)
    }
    if modes[0].ActiveUsers != 5 {
        t.Fatalf("expected 5 active users, got %d", modes[0].ActiveUsers)
    }
}
// Test for GetTotalActiveUsersLogic
func TestGetTotalActiveUsersLogic(t *testing.T) {
    collection := setupTestDB(t)
    ctx := context.Background()

    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }


    // Insert test data
    collection.InsertOne(ctx, logic.ModeUsage{ActiveUsers: 3})
    collection.InsertOne(ctx, logic.ModeUsage{ActiveUsers: 2})

    total, err := logic.GetTotalActiveUsersLogic(ctx, collection, redisCache)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if total != 5 {
        t.Fatalf("expected total active users 5, got %d", total)
    }
}

func TestJoinModeLogic(t *testing.T) {
    collection := setupTestDB(t)
    ctx := context.Background()

    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }

    // Insert initial mode
    collection.InsertOne(ctx, logic.ModeUsage{
        ModeName:    "TestMode",
        ActiveUsers: 0,
        Players:     []string{},
    })

    err = logic.JoinModeLogic(ctx, collection, redisCache ,"TestMode", "player1")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    var mode logic.ModeUsage
    collection.FindOne(ctx, bson.M{"mode_name": "TestMode"}).Decode(&mode)

    if mode.ActiveUsers != 1 {
        t.Fatalf("expected active users 1, got %d", mode.ActiveUsers)
    }
    if len(mode.Players) != 1 || mode.Players[0] != "player1" {
        t.Fatalf("expected player1 in players, got %v", mode.Players)
    }
}
func TestLeaveModeLogic(t *testing.T) {
    collection := setupTestDB(t)
    ctx := context.Background()
    
    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }

    // Insert initial mode with a player
    collection.InsertOne(ctx, logic.ModeUsage{
        ModeName:    "TestMode",
        ActiveUsers: 1,
        Players:     []string{"player1"},
    })

    err = logic.LeaveModeLogic(ctx, collection, redisCache , "TestMode", "player1")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    var mode logic.ModeUsage
    collection.FindOne(ctx, bson.M{"mode_name": "TestMode"}).Decode(&mode)

    if mode.ActiveUsers != 0 {
        t.Fatalf("expected active users 0, got %d", mode.ActiveUsers)
    }
    if len(mode.Players) != 0 {
        t.Fatalf("expected no players, got %v", mode.Players)
    }
}

func TestGetModeDetailsLogic(t *testing.T) {
	collection := setupTestDB(t)
	ctx := context.Background()
    
    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }
	// Insert test data
	// Ensure that this struct matches what GetModeDetailsLogic expects
	collection.InsertOne(ctx, logic.ModeUsage{
		ModeName:    "TestMode",
		ActiveUsers: 5,
		AreaCode:    "123",
		LastUpdated: time.Now(),
		// Remove GameState if it's not part of ModeDetailsResponse
	})

	// Call the function under test
	modeDetails, err := logic.GetModeDetailsLogic(ctx, collection, redisCache ,"TestMode")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Validate the response
	if modeDetails.ModeName != "TestMode" {
		t.Fatalf("expected mode name 'TestMode', got %s", modeDetails.ModeName)
	}
	if modeDetails.ActiveUsers != 5 {
		t.Fatalf("expected active users 5, got %d", modeDetails.ActiveUsers)
	}
	if modeDetails.AreaCode != "123" {
		t.Fatalf("expected area code '123', got %s", modeDetails.AreaCode)
	}
	// Add more assertions based on the actual fields in ModeDetailsResponse
}
func TestGetActiveUsersByAreaCodeLogic(t *testing.T) {
    collection := setupTestDB(t)
    ctx := context.Background()

    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }

    // Insert test data
    collection.InsertOne(ctx, logic.ModeUsage{
        AreaCode:    "123",
        ActiveUsers: 5,
    })
    collection.InsertOne(ctx, logic.ModeUsage{
        AreaCode:    "123",
        ActiveUsers: 3,
    })
    collection.InsertOne(ctx, logic.ModeUsage{
        AreaCode:    "456",
        ActiveUsers: 2,
    })

    total, err := logic.GetActiveUsersByAreaCodeLogic(ctx, collection,redisCache ,"123")
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }
    if total != 8 {
        t.Fatalf("expected total active users for area code '123' to be 8, got %d", total)
    }
}

func TestUpdateGameStateLogic(t *testing.T) {
    collection := setupTestDB(t) // Assume this sets up a test MongoDB collection
    ctx := context.Background()

    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }

    // Insert initial test data
    modeName := "TestMode"
    initialGameState := "active"
    collection.InsertOne(ctx, logic.ModeUsage{
        ModeName:    modeName,
        ActiveUsers: 5,
        AreaCode:    "123",
        Players:     []string{},
        GameState:   initialGameState,
        LastUpdated: time.Now(),
    })

    // Update the game state
    newGameState := "paused"
    err = logic.UpdateGameStateLogic(ctx, collection,redisCache, modeName, newGameState)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the game state was updated
    var updatedMode logic.ModeUsage
    err = collection.FindOne(ctx, bson.M{"mode_name": modeName}).Decode(&updatedMode)
    if err != nil {
        t.Fatalf("expected to find mode %s, got error: %v", modeName, err)
    }

    if updatedMode.GameState != newGameState {
        t.Fatalf("expected game state '%s', got '%s'", newGameState, updatedMode.GameState)
    }

    // Optionally, check if LastUpdated was modified
    if updatedMode.LastUpdated.Before(time.Now().Add(-time.Second)) {
        t.Fatalf("expected LastUpdated to be recent, got %v", updatedMode.LastUpdated)
    }
}

func TestGetPlayersLogic(t *testing.T) {
    collection := setupTestDB(t) // Assume this sets up a test MongoDB collection
    ctx := context.Background()

    // Initialize Redis cache
    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }

    // Insert initial test data
    modeName := "TestMode"
    players := []string{"player1", "player2", "player3"}
    _, err = collection.InsertOne(ctx, logic.ModeUsage{
        ModeName:    modeName,
        ActiveUsers: 3,
        AreaCode:    "123",
        Players:     players,
        GameState:   "active",
        LastUpdated: time.Now(),
    })
    if err != nil {
        t.Fatalf("failed to insert test data: %v", err)
    }

    // Call the function to get players (first time should hit DB and set cache)
    retrievedPlayers, err := logic.GetPlayersLogic(ctx, collection, redisCache, modeName)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the retrieved players match the inserted players
    if len(retrievedPlayers) != len(players) {
        t.Fatalf("expected %d players, got %d", len(players), len(retrievedPlayers))
    }
    for i, player := range players {
        if retrievedPlayers[i] != player {
            t.Fatalf("expected player '%s', got '%s'", player, retrievedPlayers[i])
        }
    }

    // Call the function again (this time should hit cache)
    cachedPlayers, err := logic.GetPlayersLogic(ctx, collection, redisCache, modeName)
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the cached players match the inserted players
    if len(cachedPlayers) != len(players) {
        t.Fatalf("expected %d players from cache, got %d", len(players), len(cachedPlayers))
    }
    for i, player := range players {
        if cachedPlayers[i] != player {
            t.Fatalf("expected player '%s' from cache, got '%s'", player, cachedPlayers[i])
        }
    }
}

func TestGetGameModeStatsLogic(t *testing.T) {
    collection := setupTestDB(t) // Assume this sets up a test MongoDB collection
    ctx := context.Background()

    redisCache, err := cache.InitializeCache("localhost:6379", "", 0)
    if err != nil {
        t.Fatalf("Failed to initialize Redis cache: %v", err)
    }

    // Insert test data
    modes := []logic.ModeUsage{
        {
            ModeName:    "Mode1",
            ActiveUsers: 5,
            AreaCode:    "123",
            Players:     []string{"player1", "player2"},
            GameState:   "active",
            LastUpdated: time.Now(),
        },
        {
            ModeName:    "Mode2",
            ActiveUsers: 3,
            AreaCode:    "456",
            Players:     []string{"player3"},
            GameState:   "active",
            LastUpdated: time.Now(),
        },
        {
            ModeName:    "Mode3",
            ActiveUsers: 0,
            AreaCode:    "789",
            Players:     []string{},
            GameState:   "inactive",
            LastUpdated: time.Now(),
        },
    }

    for _, mode := range modes {
        _, err := collection.InsertOne(ctx, mode)
        if err != nil {
            t.Fatalf("failed to insert test data: %v", err)
        }
    }

    // Call the function to get game mode stats
    stats, err := logic.GetGameModeStatsLogic(ctx, collection, redisCache )
    if err != nil {
        t.Fatalf("expected no error, got %v", err)
    }

    // Verify the total modes
    expectedTotalModes := int32(len(modes))
    if stats.TotalModes != expectedTotalModes {
        t.Fatalf("expected total modes %d, got %d", expectedTotalModes, stats.TotalModes)
    }

    // Calculate expected total active users
    expectedTotalActiveUsers := int32(0)
    for _, mode := range modes {
        expectedTotalActiveUsers += int32(mode.ActiveUsers)
    }

    // Verify the total active users
    if stats.TotalActiveUsers != expectedTotalActiveUsers {
        t.Fatalf("expected total active users %d, got %d", expectedTotalActiveUsers, stats.TotalActiveUsers)
    }
}