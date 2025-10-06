NFT_BINARY=nftServiceApp

## build_nft: builds the nft binary as a linux executable
build_nft:
	@echo "Building nft binary..."
	@env GOOS=linux CGO_ENABLED=0 go build -o ${NFT_BINARY} ./cmd/api
	@cp ${NFT_BINARY} ../web-microservice/nft-service/
	@echo "Done!"