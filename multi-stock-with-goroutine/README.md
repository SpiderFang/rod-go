# TWSE Stock Scraper (Multi-Stock with Goroutine)

這是一個使用 Go 語言與 [go-rod](https://github.com/go-rod/rod) 開發的網頁爬蟲工具，用於自動化抓取台灣證券交易所 (TWSE) 的個股歷史交易資料（個股日成交資訊）。

本專案展示了如何使用 Goroutine 進行並發抓取，並透過 Buffered Channel 實作 Semaphore (信號量) 來限制並發數量，是學習 Go 並發程式設計與網頁爬蟲的實用範例。

## 功能特色

*   **自動化瀏覽器操作**: 使用 DevTools Protocol 控制瀏覽器，模擬真實使用者行為，可處理動態渲染的網頁。
*   **並發執行 (Concurrency)**: 支援同時抓取多檔股票，大幅縮短總執行時間。
*   **流量控制 (Semaphore)**: 透過 Buffered Channel 限制同時執行的瀏覽器分頁數量（預設為 3），避免因開啟過多 Chrome 分頁導致資源耗盡或觸發網站的 Rate Limiting。
*   **錯誤恢復 (Panic Recover)**: 內建 `recover` 機制，確保單一股票抓取失敗（如超時或元素未找到）時，不會導致整個程式崩潰，其餘任務仍可繼續執行。
*   **CSV 輸出**: 自動將抓取到的資料儲存為 CSV 檔案，並加入 UTF-8 BOM 檔頭，確保使用 Microsoft Excel 開啟時中文顯示正常。

## 前置需求

*   Go 1.16 或更高版本
*   Google Chrome 或 Chromium 瀏覽器 (go-rod 執行時會自動尋找或下載)

## 安裝與執行

1.  **進入專案目錄**

    ```bash
    cd /Users/spider/Workspace/GoProjects/rod-go/multi-stock-with-goroutine
    ```

2.  **安裝相依套件**

    ```bash
    go mod tidy
    ```

3.  **執行程式**

    ```bash
    go run main.go
    ```

## 程式碼結構重點

*   **`main.go`**:
    *   定義 `stockList`：欲抓取的股票代號列表。
    *   **Semaphore 實作**：使用 `sem := make(chan struct{}, 3)` 建立容量為 3 的通道，控制同時運行的 Goroutine 數量。
    *   **`scrapeStock` 函數**：負責開啟分頁、輸入股票代號、點擊查詢、解析表格資料並寫入 CSV。
    *   **錯誤處理**：使用 `defer` 配合 `recover()` 捕捉執行期間的 Panic。

## 輸出結果

程式執行完畢後，會在當前目錄下生成對應的 CSV 檔案，檔名格式為 `stock_data_{股票代號}.csv`，例如：

*   `stock_data_2330.csv`
*   `stock_data_2317.csv`
*   ...

## 注意事項

*   **Context deadline exceeded**: 若網路環境較慢或電腦負載過高，可能會發生等待元素超時的錯誤。程式碼中已將關鍵操作的超時時間設定為 40 秒。

*   **網站結構變動**: 若 TWSE 網站的 HTML 結構發生變化，可能需要更新程式碼中的選擇器 (Selector)。

## 自訂查詢年份與月份

如果想要查詢特定年份與月份的個股日成交資訊，可以透過操作網頁上的下拉式選單來達成。
要控制網頁上的下拉式選單（Dropdown），我們可以使用 go-rod 的 MustSelect 方法。

針對台灣證券交易所 (TWSE) 的「個股日成交資訊」頁面，年份的下拉選單 name 屬性通常是 yy，月份則是 mm。

以下是修改步驟：

1. 修改 scrapeStock 函式：增加 year 和 month 兩個參數，並在輸入股票代號前，先選取好年份與月份。
2. 修改 main 函式：定義你想要查詢的年份（需注意證交所通常使用民國年）與月份，並傳遞給 scrapeStock。

修改重點說明：

1. MustSelect(value): 這是 go-rod 用來操作 <select> 元素的關鍵方法。它會根據你傳入的值（例如 "114" 或 "12"）自動去匹配下拉選單中的 <option> 選項。
2. 定位器 (Selector): 我們使用 select[name='yy'] 來定位年份選單，select[name='mm'] 來定位月份選單，這是該網頁表單的標準屬性。
3. 民國年: 請注意 targetYear 必須設為民國年（例如 2025 年要輸入 "114"），因為該網頁的選項值是使用民國紀年。