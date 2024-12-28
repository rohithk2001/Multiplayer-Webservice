package logic

import (
    "context"
    "encoding/json"
    "multiplayer-webservice/internal/cache"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "multiplayer-webservice/internal/proto"
     "fmt"
    "time"

)


type ModeUsage struct {
	ModeName    string `bson:"mode_name"`
	ActiveUsers int    `bson:"active_users"`
	AreaCode    string `bson:"area_code"`
	Players     []string  `bson:"players"`
    GameState   string    `bson:"game_state"`
    LastUpdated time.Time `bson:"last_updated"`
	
}




// InitializeCache initializes the Redis cache for the logic layer
func GetModeUsageLogic(ctx context.Context, collection *mongo.Collection, redisCache *cache.RedisCache) ([]*proto.ModeUsage, error) {
    // Define cache key
    cacheKey := "mode_usage"

    // Try to get data from cache
    cachedData, err := redisCache.Get(ctx, cacheKey)
    if err == nil {
        // Cache hit: Unmarshal and return data
        var modes []*proto.ModeUsage
        if jsonErr := json.Unmarshal([]byte(cachedData), &modes); jsonErr == nil {
            return modes, nil
        }
    }

    // Cache miss: Fetch data from MongoDB
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    var modes []*proto.ModeUsage
    for cursor.Next(ctx) {
        var mode ModeUsage
        if err := cursor.Decode(&mode); err != nil {
            return nil, err
        }
        modes = append(modes, &proto.ModeUsage{
            ModeName:    mode.ModeName,
            ActiveUsers: int32(mode.ActiveUsers),
            AreaCode:    mode.AreaCode,
        })
    }

    // Store the fetched data in the cache
    if jsonData, err := json.Marshal(modes); err == nil {
        redisCache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute)
    }

    return modes, nil
}

func GetTotalActiveUsersLogic(ctx context.Context, collection *mongo.Collection, redisCache *cache.RedisCache) (int32, error) {
    cacheKey := "total_active_users"

    // Check Redis Cache
    cachedData, err := redisCache.Get(ctx, cacheKey)
    if err == nil {
        var total int32
        if jsonErr := json.Unmarshal([]byte(cachedData), &total); jsonErr == nil {
            return total, nil
        }
    }

    // Query MongoDB on cache miss
    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        return 0, fmt.Errorf("failed to query MongoDB: %w", err)
    }
    defer cursor.Close(ctx)

    totalActiveUsers := int32(0)
    for cursor.Next(ctx) {
        var mode ModeUsage
        if err := cursor.Decode(&mode); err != nil {
            return 0, fmt.Errorf("failed to decode MongoDB document: %w", err)
        }
        totalActiveUsers += int32(mode.ActiveUsers)
    }

    if err := cursor.Err(); err != nil {
        return 0, fmt.Errorf("cursor error: %w", err)
    }

    // Cache the result
    if jsonData, err := json.Marshal(totalActiveUsers); err == nil {
        redisCache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute)
    }

    return totalActiveUsers, nil
}

func GetModeDetailsLogic(ctx context.Context, collection *mongo.Collection, cache *cache.RedisCache, modeName string) (*proto.ModeDetailsResponse, error) {
    // Define cache key for this mode
    cacheKey := "mode_details:" + modeName

    // Try to get data from cache
    cachedData, err := cache.Get(ctx, cacheKey)
    if err == nil {
        var modeDetails proto.ModeDetailsResponse
        if jsonErr := json.Unmarshal([]byte(cachedData), &modeDetails); jsonErr == nil {
            return &modeDetails, nil
        }
    }

    // Cache miss: Fetch data from MongoDB
    var mode ModeUsage
    err = collection.FindOne(ctx, bson.M{"mode_name": modeName}).Decode(&mode)
    if err != nil {
        return nil, err
    }

    // Prepare the response
    modeDetails := &proto.ModeDetailsResponse{
        ModeName:    mode.ModeName,
        Description: "Detailed description of the mode", // This can be customized or fetched from DB
        ActiveUsers: int32(mode.ActiveUsers),
        AreaCode:    mode.AreaCode,
    }

    // Store the result in cache
    if jsonData, err := json.Marshal(modeDetails); err == nil {
        cache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute)
    }

    return modeDetails, nil
}

func GetActiveUsersByAreaCodeLogic(ctx context.Context, collection *mongo.Collection, cache *cache.RedisCache, areaCode string) (int32, error) {
    // Define cache key
    cacheKey := "active_users_area_code_" + areaCode

    // Try to get data from cache
    cachedData, err := cache.Get(ctx, cacheKey)
    if err == nil {
        // Cache hit: Parse and return cached result
        var totalActiveUsers int32
        if jsonErr := json.Unmarshal([]byte(cachedData), &totalActiveUsers); jsonErr == nil {
            return totalActiveUsers, nil
        }
    }

    // Cache miss: Query MongoDB
    cursor, err := collection.Find(ctx, bson.M{"area_code": areaCode})
    if err != nil {
        return 0, err
    }
    defer cursor.Close(ctx)

    totalActiveUsers := int32(0)
    for cursor.Next(ctx) {
        var mode ModeUsage
        if err := cursor.Decode(&mode); err != nil {
            return 0, err
        }
        totalActiveUsers += int32(mode.ActiveUsers)
    }

    if err := cursor.Err(); err != nil {
        return 0, err
    }

    // Store the fetched data in cache
    if jsonData, err := json.Marshal(totalActiveUsers); err == nil {
        cache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute)
    }

    return totalActiveUsers, nil
}

func GetGameModeStatsLogic(ctx context.Context, collection *mongo.Collection, cache *cache.RedisCache) (*proto.GameModeStatsResponse, error) {
    // Define cache key
    cacheKey := "game_mode_stats"

    // Try to get data from cache
    cachedData, err := cache.Get(ctx, cacheKey)
    if err == nil {
        // Cache hit: Parse and return cached result
        var stats proto.GameModeStatsResponse
        if jsonErr := json.Unmarshal([]byte(cachedData), &stats); jsonErr == nil {
            return &stats, nil
        }
    }

    // Cache miss: Query MongoDB to compute statistics
    totalModes, err := collection.CountDocuments(ctx, bson.M{})
    if err != nil {
        return nil, err
    }

    cursor, err := collection.Find(ctx, bson.M{})
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)

    totalActiveUsers := int32(0)
    for cursor.Next(ctx) {
        var mode ModeUsage
        if err := cursor.Decode(&mode); err != nil {
            return nil, err
        }
        totalActiveUsers += int32(mode.ActiveUsers)
    }

    if err := cursor.Err(); err != nil {
        return nil, err
    }

    stats := &proto.GameModeStatsResponse{
        TotalModes:       int32(totalModes),
        TotalActiveUsers: totalActiveUsers,
    }

    // Store the fetched statistics in cache
    if jsonData, err := json.Marshal(stats); err == nil {
        cache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute)
    }

    return stats, nil
}

func JoinModeLogic(ctx context.Context, collection *mongo.Collection, cache *cache.RedisCache, modeName, playerId string) error {
    // Update MongoDB: Add the player and increment active users
    filter := bson.M{"mode_name": modeName}
    update := bson.M{
        "$inc": bson.M{"active_users": 1},
        "$push": bson.M{"players": playerId},
        "$set": bson.M{"last_updated": time.Now()},
    }

    _, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }

    // Cache Invalidation: Remove cache entries related to this mode
    modeCacheKey := "mode_details_" + modeName
    statsCacheKey := "game_mode_stats"

    cache.Delete(ctx, modeCacheKey) // Invalidate mode details cache
    cache.Delete(ctx, statsCacheKey) // Invalidate game statistics cache

    return nil
}

func LeaveModeLogic(ctx context.Context, collection *mongo.Collection, redisCache *cache.RedisCache, modeName, playerId string) error {
    // Update MongoDB to remove the player and decrement active users
    filter := bson.M{"mode_name": modeName}
    update := bson.M{
        "$inc": bson.M{"active_users": -1},       // Decrement active users
        "$pull": bson.M{"players": playerId},     // Remove player from players array
        "$set": bson.M{"last_updated": time.Now()}, // Update timestamp
    }

    _, err := collection.UpdateOne(ctx, filter, update)
    if err != nil {
        return err
    }

    // Invalidate cache for the mode and related data
    modeCacheKey := "mode_details_" + modeName
    statsCacheKey := "game_mode_stats"
    redisCache.Delete(ctx, modeCacheKey) // Invalidate mode details cache
    redisCache.Delete(ctx, statsCacheKey) // Invalidate game statistics cache

    return nil
}

func GetPlayersLogic(ctx context.Context, collection *mongo.Collection, redisCache *cache.RedisCache, modeName string) ([]string, error) {
    // Define cache key for the players list
    cacheKey := "players_list_" + modeName

    // Try to fetch the players list from the cache
    cachedData, err := redisCache.Get(ctx, cacheKey)
    if err == nil {
        // Cache hit: Unmarshal and return players list
        var players []string
        if jsonErr := json.Unmarshal([]byte(cachedData), &players); jsonErr == nil {
            return players, nil
        }
    }

    // Cache miss: Fetch players from MongoDB
    filter := bson.M{"mode_name": modeName}
    var result struct {
        Players []string `bson:"players"`
    }
    err = collection.FindOne(ctx, filter).Decode(&result)
    if err != nil {
        return nil, err
    }

    // Store the players list in the cache
    if jsonData, err := json.Marshal(result.Players); err == nil {
        redisCache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute)
    }

    return result.Players, nil
}

// UpdateGameStateLogic updates the game state of a mode and handles cache synchronization
func UpdateGameStateLogic(ctx context.Context, collection *mongo.Collection, cache *cache.RedisCache, modeName, gameState string) error {
	// Define the cache key based on the mode name
	cacheKey := "mode_" + modeName

	// Try to get the data from the cache
	cachedData, err := cache.Get(ctx, cacheKey)
	if err == nil {
		// Cache hit: Unmarshal the cached data and update the game state
		var mode ModeUsage
		if jsonErr := json.Unmarshal([]byte(cachedData), &mode); jsonErr == nil {
			mode.GameState = gameState

			// Store the updated mode in the cache
			if jsonData, err := json.Marshal(mode); err == nil {
				cache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute) // Cache TTL set to 10 minutes
			}
		}
	}

	// Update the game state in the MongoDB collection
	filter := bson.M{"mode_name": modeName}
	update := bson.M{
		"$set": bson.M{"game_state": gameState, "last_updated": time.Now()},
	}

	_, err = collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// After MongoDB update, also update the cache with the latest data
	var updatedMode ModeUsage
	err = collection.FindOne(ctx, bson.M{"mode_name": modeName}).Decode(&updatedMode)
	if err != nil {
		return err
	}

	// Store the updated mode in the cache
	if jsonData, err := json.Marshal(updatedMode); err == nil {
		cache.Set(ctx, cacheKey, string(jsonData), 10*time.Minute) // Cache TTL set to 10 minutes
	}

	return nil
}
