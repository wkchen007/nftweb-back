
# nftweb-back

Go 語言開發的 NFT Web API 後端專案，使用 Docker 部署。


## 環境說明

- Go 1.24.0
- Docker Compose 2.0
- Sepolia 測試鏈（預設連線，可於 .env 設定 RPC）

## API 功能

### 錢包功能

- 查詢錢包餘額
	- `GET /wallet/balance` 取得指定地址的 ETH 餘額
- 轉帳功能
	- `POST /wallet/transfer` 發送 ETH 至指定地址

## 快速開始

### 1. 建立映像檔並啟動服務

```bash
docker compose up -d
```

### 2. API 服務預設監聽 8080 port

- http://localhost:8080

### 3. 健康檢查

- http://localhost:8080/healthz

## 授權

MIT License
