# GIDO Guardian Bot

## 專案概述

GIDO Guardian Bot 是一個 Discord bot，用於自動查詢和監控 **吉哆火鍋百匯** 的等候資訊，幫助使用者更便利地了解目前的候位情況。

## 主要功能

- 🔍 **即時等候查詢**: 透過 Discord 指令快速查詢餐廳等候時間
- 📊 **等候資訊追蹤**: 記錄和追蹤等候人數的變化
- 👥 **使用者追蹤**: 記錄使用者的查詢歷史和互動
- 🤖 **Discord 整合**: 原生支援 Discord 斜線指令和互動

## 技術棧

- **語言**: Go 1.x
- **依賴**: 
  - `github.com/bwmarrin/discordgo` - Discord API 包裝器
  - 標準庫中的 HTTP 和時間處理

## 專案結構

```
gido-guardian-bot/
├── main.go              # 應用程式入口
├── bot/
│   ├── bot.go          # Bot 核心邏輯和初始化
│   ├── interaction.go   # Discord 互動處理
│   └── user_tracker.go  # 使用者追蹤功能
├── gido/
│   ├── gido.go         # GIDO API 核心
│   ├── request.go      # HTTP 請求處理
│   ├── ticket_tracker.go # 票券/等候追蹤
│   └── wait_info.go    # 等候資訊資料結構
├── go.mod              # Go 模組定義
├── go.sum              # Go 依賴校驗
├── LICENSE             # 授權條款
└── README.md           # 此檔案
```

## 快速開始

### 前置需求

- Go 1.x 或更新版本
- 有效的 Discord Bot Token
- 有 [吉哆火鍋百匯](https://www.weshine.com.tw/) 等候系統的存取權限

### 安裝

1. **複製專案**
   ```bash
   git clone https://github.com/SDxBacon/gido-guardian-bot.git
   cd gido-guardian-bot
   ```

2. **安裝依賴**
   ```bash
   go mod download
   ```

3. **設定環境變數**
   
   建立 `.env` 檔案或設定系統環境變數：
   ```bash
   export BOT_TOKEN="your_bot_token_here"
   ```

4. **執行應用程式**
   ```bash
   go run main.go
   ```

## 使用方式

### Discord 指令

#### 1. 查詢等候資訊
```
/wait-info
```
即時查詢吉哆火鍋百匯目前的等候資訊，包括當前叫號和總共等待組數。

**回應範例：**
```
當前叫號: 123，總共等待組數: 45
```

#### 2. 開始追蹤票號
```
/watching <ticket_number>
```
開始監控指定的票號，機器人會定期檢查目前的叫號進度。當您的票號即將到達或已經超過時，機器人會通知您。

**參數：**
- `ticket_number` (必填): 您的票號（整數）

**範例：**
```
/watching 150
```

**bot 會回應：**
- 開始追蹤時: `開始追蹤 Ticket: 150`
- 定期更新: `@user 當前票號: 145，總共等待組數: 5`
- 到達時: `@user 您的票號: 150 已經到達或已經過號！`

#### 3. 停止追蹤票號
```
/stop-watching
```
停止監控您目前正在追蹤的票號。

**回應：**
- 成功停止: `正在停止追蹤 Ticket: 150`
- 未追蹤任何票號: `您沒有正在追蹤的 Ticket`

### 開發者 API

#### 查詢等候資訊 (Go 套件)
```go
import "gido"

waitInfo, err := gido.GetCurrentWaitInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("當前票號: %s, 等待組數: %s\n", waitInfo.CurrentNumber.String(), waitInfo.TotalWaiting.String())
```

## API 端點

本專案與以下 API 互動：

- **GIDO 等候系統**: `https://vpn.weshine.com.tw:8089/WaitInfoWeb/WaitInfo_GIDOHandler.ashx`
  - 參數:
    - `act`: 操作類型 (預設: `WaitInfo`)
    - `DEP_CODE`: 部門代碼 (預設: `吉哆火鍋百匯`)
    - `Kind`: 類別代碼 (預設: `a1`)
    - `date`: 查詢日期
    - `_`: 時間戳 (防快取)

## 設定

### 環境變數

| 變數 | 說明 | 必須 |
|------|------|------|
| `BOT_TOKEN` | Discord Bot Token | ✅ 是 |

### 應用程式設定

在 `bot/bot.go` 中修改以下設定：
- Discord 指令前綴
- 等候查詢時間間隔
- 其他 bot 行為

## 錯誤處理

- ✅ Discord 連線錯誤: 應用程式會記錄詳細的連線失敗資訊
- ✅ GIDO API 錯誤: 自動重試機制 (待實裝)
- ✅ HTTP 超時: 預設超時時間為 2 秒

## 資料追蹤

### 使用者追蹤 (`bot/user_tracker.go`)
記錄以下資訊：
- 使用者 ID
- 查詢次數
- 最後查詢時間

### 票券追蹤 (`gido/ticket_tracker.go`)
追蹤等候資訊的變化：
- 當前等候人數
- 歷史等候記錄

## 常見問題

### Q: 如何取得 Discord Bot Token？
A: 請造訪 [Discord Developer Portal](https://discord.com/developers/applications)，建立應用程式並產生 Bot Token。

### Q: 為什麼等候資訊無法查詢？
A: 請檢查：
1. 網路連線是否正常
2. GIDO API 端點是否可用 (`https://vpn.weshine.com.tw:8089`)
3. 是否正確的日期進行查詢

另外，吉哆的侯位系統供應商(weshine)伺服器並不是非常穩定，時不時會 return falsy value 的情況發生。

### Q: 機器人無法連線到 Discord？
A: 檢查以下項目：
1. `BOT_TOKEN` 環境變數是否正確設定
2. Bot 是否已邀請到 Discord 伺服器
3. Bot 是否具有必要的權限

### 開發規範

- 使用 Go 標準編碼規範 (`gofmt`, `go vet`)
- 新功能必須包含適當的錯誤處理
- 提交前運行 `go test` 進行測試

## 授權

此專案採用 MIT 授權。詳見 [LICENSE](./LICENSE) 檔案。

---

**最後更新**: 2025 年 12 月 3 日