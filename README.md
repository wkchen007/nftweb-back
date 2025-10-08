
# nftweb-back

Go 語言開發的 NFT Web API 後端專案，使用 Docker 部署。


## 環境說明

- Go 1.24.0
- go-ethereum：以太坊鏈上互動（合約、錢包、交易）
- Solidity NFT ERC-721智能合約
- Sepolia 測試鏈（預設連線，可於 .env 設定 RPC）
- Postgres：儲存 NFT 與用戶資料
- RabbitMQ：發送登入通知信和寫入 log
- Redis：儲存 JWT 認證資料
- MongoDB：儲存 log 資料
- Docker Compose 2.0

## API 功能

### 商品查詢

- 查詢抽獎商品
	- `GET /demo` 回傳所有 NFT 抽獎商品資訊

### NFT 抽獎功能

- 開盲盒（抽獎）
    - `POST /nft/mint` 花費ETH，隨機抽獎獲得NFT
- 查詢抽獎結果
    - `GET /nft/tokensOfOwner` 查詢抽中的NFT

### 錢包功能

- 查詢錢包餘額
	- `GET /wallet/balance` 取得指定地址的 ETH 餘額
- 轉帳功能
	- `POST /wallet/transfer` 發送 ETH 至指定地址

## 快速開始

### 1. 建立 .env 檔案

請參考 `.env.example` 檔案，複製為 `.env` 並根據需求修改Ethereum RPC、JWT金鑰等設定。

```bash
cp .env.example .env
```

### 2. 建立 config.yaml 檔案

部署 NFT 合約，ABI 存入 json ，請參考 `config.yaml`，根據需求修改設定。

### 3. 修改 DB sql 檔案

登入有作用戶錢包地址檢查，要在 sql 目錄下 `create_tables.sql` 檔案，新增用戶的錢包地址。
登入預設密碼是 `secret`，可自行修改。

### 4. 建立映像檔並啟動服務

```bash
docker compose up -d
```

### 5. API 服務預設監聽 8080 port

- http://localhost:8080

### 6. 健康檢查

- http://localhost:8080/healthz

## 相關專案

- 前端專案 [nftweb-front](https://github.com/wkchen007/nftweb-front)

## 授權

MIT License
