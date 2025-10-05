package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/wkchen007/nftweb-back/internal/ethcli"
	"github.com/wkchen007/nftweb-back/internal/nft"
	"github.com/wkchen007/nftweb-back/internal/repository"
	"github.com/wkchen007/nftweb-back/internal/repository/dbrepo"
)

type application struct {
	httpAddr  string
	DSN       string
	auth      Auth
	DB        repository.DatabaseRepo
	ethClient *ethcli.Client
	nft       *nft.Handlers
}

func main() {
	var app application

	// 載入 .env 檔案
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, skip loading")
	}

	// 讀取 HTTP 監聽位址，預設 :8080
	defaultAddr := os.Getenv("HTTP_ADDR")
	if defaultAddr == "" {
		defaultAddr = ":8080"
	}
	flag.StringVar(&app.httpAddr, "httpAddr", defaultAddr, "HTTP network address")
	flag.StringVar(&app.DSN, "dsn", "host=postgres port=5432 user=postgres password=postgres dbname=nftweb sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection string")
	flag.Parse()

	// connect to the databases
	connPostgres, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}
	app.DB = &dbrepo.PostgresDBRepo{DB: connPostgres}
	defer app.DB.Connection().Close()

	// 讀取 JWT 相關設定
	app.auth = Auth{
		Issuer:        os.Getenv("JWT_ISSUER"),
		Audience:      os.Getenv("JWT_AUDIENCE"),
		Secret:        os.Getenv("JWT_SECRET"),
		TokenExpiry:   time.Minute * 5,
		RefreshExpiry: time.Hour * 24,
		CookiePath:    "/",
		CookieName:    "refresh_token",
		CookieDomain:  os.Getenv("COOKIE_DOMAIN"),
	}
	log.Printf("JWT config: issuer=%s, audience=%s, cookie_domain=%s", app.auth.Issuer, app.auth.Audience, app.auth.CookieDomain)

	// 從環境變數讀設定檔路徑，預設用 configs/config.yaml
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "configs/config.yaml"
	}

	// 載入NFT設定檔
	cfg, err := nft.LoadConfig(configPath)
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	// 建立以太連線(封裝在 internal/ethcli)
	ethc, err := ethcli.New(os.Getenv("RPC_URL"))
	if err != nil {
		log.Fatalf("cannot create eth client: %v", err)
	}
	defer ethc.Close()
	app.ethClient = ethc

	// 建立 NFT 服務(封裝在 internal/nft)
	svc, err := nft.NewServiceFromConfig(ethc, cfg)
	if err != nil {
		log.Fatal("failed to create nft service:", err)
	}
	svc.DB = app.DB
	app.nft = nft.NewHandlers(svc)
	log.Print("[nft] service created")

	// 啟動 HTTP server
	log.Printf("Listening on http://0.0.0.0%s\n", app.httpAddr)
	if err := http.ListenAndServe(app.httpAddr, app.routes()); err != nil {
		log.Fatal("failed to start server:", err)
	}
}
