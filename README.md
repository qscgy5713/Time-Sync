# Time-Sync

免登入、免註冊的多人時間對齊工具。發起人建立候選時段後，參與者透過連結圈選有空時段，系統以熱區圖呈現最大交集。

技術細節與資料庫/API 設計請見 `gemini-code-1789712746486.md`。

## 專案結構

```
backend/    Go (Gin) + PostgreSQL API
frontend/   Vue 3 + Vite + Tailwind + Pinia
```

## 開發方式

### 一鍵啟動（Docker）

```
docker compose up --build
```

- 前端：http://localhost:5173
- 後端 API：http://localhost:8080

### 本機分別啟動

```
# 1. 啟動資料庫
docker compose up -d postgres

# 2. 後端
cd backend
DATABASE_URL="postgres://timesync:timesync@localhost:5432/timesync?sslmode=disable" go run .

# 3. 前端
cd frontend
npm install
npm run dev
```

後端啟動時會自動建立資料表（`internal/db/migrations_embed.sql`），不需另外執行 migration 工具。
