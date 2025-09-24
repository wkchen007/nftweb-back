package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/wkchen007/nftweb-back/internal/ethcli"
	"github.com/wkchen007/nftweb-back/internal/nft"
)

type application struct {
	httpAddr  string
	ethClient *ethcli.Client
	nft       *nft.Handlers
}

func main() {
	var app application
	var err error
	flag.StringVar(&app.httpAddr, "httpAddr", ":8080", "HTTP network address")
	flag.Parse()

	if err = godotenv.Load(); err != nil {
		log.Println("No .env file found, skip loading")
	}

	// 建立以太連線(封裝在 internal/ethcli)
	ethc, err := ethcli.NewWithSigner(os.Getenv("RPC_URL"), os.Getenv("privKey"))
	if err != nil {
		log.Fatalf("cannot create eth client: %v", err)
	}
	defer ethc.Close()

	app.ethClient = ethc

	// 建立 NFT 服務(封裝在 internal/nft)
	svc, err := nft.NewService(ethc, "internal/nft/nftABI.json", os.Getenv("CONTRACT_ADDRESS"))
	if err != nil {
		log.Fatalf("cannot create nft service: %v", err)
	}
	app.nft = nft.NewHandlers(svc)

	// 啟動 HTTP server
	log.Printf("HTTP server listening on %s ...", app.httpAddr)
	if err = http.ListenAndServe(app.httpAddr, app.routes()); err != nil {
		log.Fatal(err)
	}
}
