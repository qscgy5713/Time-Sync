# 部署到 GCP 免費 VM

用 `docker-compose.prod.yml`（不是 k3s）+ Caddy 自動 SSL + DuckDNS 免費網域。

## 1. 建立 GCP 免費 VM

### 前置作業

1. 用 Google 帳號登入 https://console.cloud.google.com
2. 如果還沒有專案：右上角「選取專案」→「新增專案」，取個名字（例如 `time-sync`）建立。
3. 第一次用 Compute Engine 會要求「啟用帳單」——需要綁一張信用卡，但只要用量留在免費額度內就不會被扣款。左側選單「帳單」照畫面指示設定即可。
4. 左側選單找「Compute Engine」→「VM 執行個體」，第一次進入會自動幫你啟用 Compute Engine API（等個 1 分鐘）。

### 建立 VM（Console 操作）

1. 進入「VM 執行個體」頁面，點「建立執行個體」。
2. **名稱**：`time-sync`（隨意）。
3. **區域**：選 `us-west1`、`us-central1` 或 `us-east1` 其中一個（免費方案只涵蓋這三個區域），區域（zone）任選其一，例如 `us-west1-b`。
4. **機器設定**：系列選「E2」，機器類型選 `e2-micro`（免費方案就是這一款，多選其他款會開始收費）。
5. **開機磁碟**：點「變更」，作業系統選「Debian」，版本選最新的 Debian 12，**磁碟類型務必選「標準永久磁碟」**（不要選 SSD／平衡式，那些不在免費額度內），容量預設 10GB 即可（免費額度上限 30GB）。
6. **防火牆**：勾選「允許 HTTP 流量」和「允許 HTTPS 流量」這兩個選項（會自動幫你建對應的防火牆規則）。
7. 其他保持預設，往下拉點「建立」。等個 30 秒到 1 分鐘，VM 就會啟動，列表上會看到它的**外部 IP**。

### 用 gcloud CLI 建立（如果你偏好指令列）

需要先在自己電腦裝 `gcloud` CLI 並 `gcloud init` 登入：

```
gcloud compute instances create time-sync \
  --zone=us-west1-b \
  --machine-type=e2-micro \
  --image-family=debian-12 \
  --image-project=debian-cloud \
  --boot-disk-type=pd-standard \
  --boot-disk-size=10GB \
  --tags=http-server,https-server

gcloud compute firewall-rules create allow-http-https \
  --allow=tcp:80,tcp:443 \
  --target-tags=http-server,https-server
```

### 注意事項

- **不要**額外開放 5432（Postgres），docker-compose.prod.yml 也沒有把它對外綁定。
- 免費額度是「每個帳單帳戶」一台 e2-micro，如果這個 GCP 帳號已經有其他 VM 在跑，可能會超額被收費。
- VM 建好後，直接在 VM 列表該行最右邊點「SSH」就能在瀏覽器裡直接連進去，不需要另外裝 SSH client。

## 2. 申請免費網域（DuckDNS）

1. 到 https://www.duckdns.org 用 GitHub/Google 登入。
2. 建立一個子網域，例如 `time-sync`，會得到 `time-sync.duckdns.org`。
3. 把它指到 VM 的外部 IP（GCP Console 的 VM 詳情頁可以看到）。
4. 建議在 GCP 保留一個「靜態外部 IP」，這樣重開機 IP 不會變、不用一直改 DuckDNS。

## 3. SSH 進 VM，安裝 Docker

```
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER
# 重新登入 SSH 讓群組生效
```

## 4. 把專案放上 VM

```
git clone https://github.com/qscgy5713/Time-Sync.git
cd Time-Sync
cp .env.example .env
```

編輯 `.env`：
```
POSTGRES_PASSWORD=<自己產生一組隨機密碼>
DOMAIN=timesync-tw.duckdns.org
```

## 5. 啟動

```
docker compose -f docker-compose.prod.yml up -d --build
```

Caddy 會自動跟 Let's Encrypt 要憑證（需要 80 埠能從外部連到這台 VM，DNS 也要先生效）。第一次啟動可能要等個 10-30 秒完成簽發。

## 6. 驗證

瀏覽器打開 `https://timesync-tw.duckdns.org`，確認是 HTTPS 且憑證有效。

## 之後更新程式碼

```
git pull
docker compose -f docker-compose.prod.yml up -d --build
```
