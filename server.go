package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/tasks/v1"
)

var AuthConfig *oauth2.Config
var DbClient *DynamoDbClient

var DynamoDbTableName string
var GoogleActionStore sync.Map
var NotiClient *NotificationClient

const CONFIG_BASE_PATH = "./google_service_configuration"

func main() {
	e := echo.New()
	if err := InitConfig(); err != nil {
		e.Logger.Fatalf("Load config failed: %v", err)
	}
	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "https://oism.app"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	DynamoDbTableName = viper.GetString("dynamoDb_table_name")
	b, err := os.ReadFile(fmt.Sprintf("%v/credentials.json", CONFIG_BASE_PATH))
	if err != nil {
		e.Logger.Fatalf("Unable to read client secret file: %v", err)
	}
	AuthConfig, err = google.ConfigFromJSON(b, tasks.TasksScope)
	if err != nil {
		e.Logger.Fatalf("unable to load SDK config, %v", err)
	}
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-1"))
	if err != nil {
		e.Logger.Fatalf("unable to load aws default config, %v", err)
	}
	DbClient = &DynamoDbClient{inner: dynamodb.NewFromConfig(cfg)}
	NotiClient = NewNotificationClient(viper.GetString("webhooks_url"))
	e.GET("/google/health_check", HealthCheck)
	e.GET("/google/auth_url", PublishGoogleAuthUrl)
	e.GET("google/check_google_action_status/:ticket", CheckGoogleActionStatus)
	e.POST("/google/code_exchange", CodeExchange)
	e.POST("/google/insert_task", InsertOneTask)
	e.Logger.Fatal(e.Start(":" + viper.GetString("app_port")))
}
