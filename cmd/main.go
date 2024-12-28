package main

import (
	"context"
	"fmt"
	"log"
	"net"


	"github.com/gin-gonic/gin"
	"multiplayer-webservice/internal/cache"
	"multiplayer-webservice/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"multiplayer-webservice/internal/handlers"
	"multiplayer-webservice/internal/proto"
)

var collection *mongo.Collection

func main() {
	err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	err = connectToMongoDB()
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}

	redisCache, err := cache.InitializeCache(config.AppConfig.RedisAddr, config.AppConfig.RedisPass, config.AppConfig.RedisDB)
	if err != nil {
		log.Fatalf("failed to initialize Redis cache: %v", err)
	}
	log.Println("Redis cache initialized successfully")

	go startGRPCServer(redisCache)

	router := gin.Default()
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Multiplayer Web Service is running!"})
	})
	router.GET("/total-active-users", getTotalActiveUsers)

	port := config.AppConfig.ServerPort
	fmt.Printf("Starting HTTP server on port %s\n", port)
	log.Fatal(router.Run(":" + port))
}

func connectToMongoDB() error {
	uri := config.AppConfig.MongoDBURI
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return err
	}
	collection = client.Database("multiplayer").Collection("modes")
	fmt.Println("Successfully connected to MongoDB!")
	return nil
}

func startGRPCServer(redisCache *cache.RedisCache) {
	lis, err := net.Listen("tcp", ":"+config.AppConfig.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	multiplayerHandler := &handlers.MultiplayerService{
		Collection: collection,
		RedisCache: redisCache,
	}
	proto.RegisterMultiplayerServiceServer(grpcServer, multiplayerHandler)
	reflection.Register(grpcServer)

	fmt.Printf("Starting gRPC server on port %s\n", config.AppConfig.GRPCPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

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
