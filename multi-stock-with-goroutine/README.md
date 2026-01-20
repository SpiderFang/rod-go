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

## 如果想抓取同一股票不同年月的資料

將「股票代號」、「年份」與「月份」都定義在 CSV 檔案中，可以讓你更靈活地安排抓取任務（例如：針對台積電抓取 11 月和 12 月的資料，而針對鴻海只抓取 12 月的資料）。

以下是修改步驟：

1. 修改 CSV 檔案 (stocks.csv)：將原本只有代號的列表，格式改為 股票代號,年份,月份的任務清單。
   原：
   2330,台積電
   2317,鴻海
   改為：
   2330,114,11
   2330,114,12
   2317,114,12
2. 修改 main.go：
    * 定義一個 SearchTask 結構來存放每一列的任務資訊。
    * 修改 readCSV 函式，讓它一次讀取三個欄位。
    * 修改 main 迴圈，從 CSV 讀取任務並傳遞給 scrapeStock。
    * 重要：修改輸出的檔名格式，加入年月（如 stock_data_2330_114_12.csv），避免不同月份的資料寫入同一個檔案造成覆蓋。

修改重點說明： 
1. 定義了一個新的結構 SearchTask: 建立了一個結構來包裝 StockNo, Year, Month，讓程式碼更具可讀性。
2. 修改readCSV 更新: 現在它會檢查 CSV 是否至少有 3 個欄位，並回傳 []SearchTask 而不是單純的字串切片。
3. 檔名唯一性: 在 scrapeStock 中，將輸出檔名改為 stock_data_{代號}_{年}_{月}.csv。這點非常重要，因為如果同時抓取同一支股票的不同月份（例如並發執行 11 月和 12 月），若檔名只有股票代號，會導致檔案被覆蓋或寫入衝突。


## 為這個專案添加單元測試（Unit Test）

為這個專案添加單元測試（Unit Test），最適合的切入點是測試 邏輯處理 的部分，也就是 readCSV 函式。這部分的邏輯獨立且單純（讀取檔案、解析 CSV、去除空白），非常適合編寫單元測試。

至於 scrapeStock 函式，因為它涉及瀏覽器操作與網路請求，通常屬於 整合測試（Integration Test） 的範疇，測試成本較高（需要 Mock 瀏覽器或架設測試伺服器），建議先從 readCSV 開始建立測試基礎。

要為這個專案添加單元測試，可以使用 Go 的內建測試框架 `testing`。以下是一些步驟和範例，說明如何為主要功能撰寫單元測試。
1. 建立測試檔案
在 main.go 同一層目錄下（/Users/spider/Workspace/GoProjects/rod-go/multi-stock-with-goroutine/），建立一個新檔案 main_test.go。
Go 語言的測試檔案必須以 _test.go 結尾，並且與被測試的程式碼位於同一個 package（這裡是 package main）。

2. 如何執行測試
開啟終端機（Terminal），切換到該目錄，然後執行以下指令：

執行所有測試：go test -v
-v 參數會顯示詳細的測試過程（Verbose）。

執行特定測試：go test -v -run TestReadCSV

測試程式碼重點解析:
1. Table-Driven Tests: 這是 Go 社群最推薦的測試寫法。我們定義了一個 tests slice，裡面包含了各種情境（正常、有空白、格式錯誤等）。這樣如果要加新測試案例，只要在 slice 裡加一行即可，不用複製貼上程式碼。

2. 暫存檔案 (os.CreateTemp): 因為 readCSV 需要讀取實體檔案，我們在測試中使用 os.CreateTemp 建立臨時檔案，並在測試結束後用 defer os.Remove 清除，這樣可以確保測試環境乾淨，不會依賴外部特定的 stocks.csv。

3. 邊界條件測試:
   - 空白處理: 測試了 2330 這種帶有空白的情況，確保 strings.TrimSpace 有正常運作。
   - 欄位不足: 測試了只有兩欄的資料，確保程式不會 Crash 且正確忽略該行。
   - 格式錯誤: 測試了 CSV 格式損壞的情況，確保函式會回傳 error。
