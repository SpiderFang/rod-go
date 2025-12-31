package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

func main() {
	url := launcher.New().
		Headless(false).
		MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	// 定義要抓取的股票代號列表
	stockList := []string{"2330", "2317", "2603", "2454", "2881", "2882", "2891", "2002", "1301", "1303"}

	var wg sync.WaitGroup
	for _, stockNo := range stockList {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			scrapeStock(browser, id)
		}(stockNo)
	}
	wg.Wait()
}

func scrapeStock(browser *rod.Browser, stockNo string) {
	fmt.Printf("正在處理股票代號: %s\n", stockNo)
	page := browser.MustPage("https://www.twse.com.tw/zh/trading/historical/stock-day-avg.html")
	defer page.MustClose() // 確保每個分頁處理完後關閉

	page.MustWaitLoad()
	page.MustWaitElementsMoreThan("form", 0)
	page.MustWaitStable() // 等待頁面穩定，確保 Loading 遮罩已消失

	// 找 input（Element UI）
	input := page.
		Timeout(15 * time.Second).
		MustElement("input[name='stockNo']"). // 使用更精確的 name 屬性定位
		MustWaitVisible()
	// fmt.Println("input:", input)

	// 方法1: 用 Type
	// input.MustClick()
	// input.MustSelectAllText()
	// input.MustType('2', '3', '3', '0')
	// 方法2: 這行可以取代 MustClick, MustSelectAllText, MustType 三行
	input.MustInput(stockNo)

	time.Sleep(300 * time.Millisecond)

	// 點查詢
	page.MustElementR("button", "查詢").MustClick()

	// 等資料表
	page.MustWaitElementsMoreThan("tbody tr", 0)
	time.Sleep(time.Second)

	rows := page.MustElements("tbody tr")
	fmt.Println("rows:", len(rows))
	fmt.Printf("找到 %d 筆資料\n", len(rows))

	// 建立 CSV 檔案
	fileName := fmt.Sprintf("stock_data_%s.csv", stockNo)
	f, err := os.Create(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 寫入 UTF-8 BOM (避免 Excel 開啟時中文亂碼)
	f.WriteString("\xEF\xBB\xBF")

	w := csv.NewWriter(f)
	defer w.Flush()

	for _, row := range rows {
		var cols []string
		for _, td := range row.MustElements("td") {
			cols = append(cols, strings.TrimSpace(td.MustText()))
		}
		fmt.Println(cols)
		if err := w.Write(cols); err != nil {
			panic(err)
		}
	}
	fmt.Printf("已儲存至 %s\n\n", fileName)
}

/*
可以使用 Go 的 goroutine 搭配 sync.WaitGroup 來並發執行每個股票代號的爬取任務。這樣可以同時開啟多個分頁進行處理，而不是等待一個結束才開始下一個。
修改：
1. 引入 sync 套件。
2. 在 main 函式中建立 sync.WaitGroup。
3. 在迴圈中使用 go func 啟動 goroutine，並正確傳遞 stockNo（避免閉包變數捕獲問題）。
4. 使用 wg.Wait() 確保所有 goroutine 執行完畢後才關閉瀏覽器。
這樣寫的好處是可以大幅提升爬取多個股票代號的效率，因為多個分頁可以同時進行網頁載入和資料處理。
注意事項:
- 資源消耗：這會同時開啟 10 個分頁（因為 Headless(false)），如果電腦效能較差或網路較慢，可能會導致卡頓。
- 網站限制：如果同時發送過多請求，目標網站（證交所）可能會有 Rate Limiting（速率限制）機制導致請求失敗。如果遇到這種情況，建議限制同時執行的 goroutine 數量（例如使用 buffered channel 作為 semaphore）。
- Console 輸出：因為是並行執行，fmt.Printf 的輸出順序會交錯在一起，這是正常現象。

解釋 go func(id string) { ... }(stockNo)
這段程式碼是 Go 語言中用來實現「並發執行」（Concurrency）的經典寫法。簡單來說，它的作用是同時派出多個「分身」去執行任務，而不是一個人做完才換下一個。
讓我們把這段程式碼拆解開來解釋：
go func(id string) {
    defer wg.Done()
    scrapeStock(browser, id)
}(stockNo)
1. go 關鍵字：啟動分身
意思：告訴程式「不要等這個函式做完，直接繼續往下跑，讓這個函式在背景（新的 Goroutine）自己跑」。
比喻：就像你在餐廳點餐，服務生（主程式）把單子丟給廚房（Goroutine）後，馬上就去服務下一桌客人，而不是站在廚房門口等到菜做好才走。
2. func(id string) { ... }：匿名函式
意思：這是一個沒有名字的函式（Anonymous Function）。因為我們只需要在這裡用一次，所以不需要特地幫它取名（例如 func myTask(...)），直接當場定義當場用。
內容：大括號 { ... } 裡面就是這個分身具體要做的事情。
3. (stockNo)：馬上呼叫並傳遞參數
位置：注意看最後面的 (stockNo)。
意思：定義完上面的匿名函式後，立刻執行它，並且把目前的 stockNo（股票代號）當作參數傳進去。
為什麼要傳參數？（重要）
在迴圈中使用 go func 時，如果不把 stockNo 傳進去，而是直接在函式裡面用 stockNo，所有的分身可能會共用同一個變數。
結果可能變成：迴圈跑得比分身啟動還快，當分身真正開始做事時，stockNo 已經變成列表中的最後一個股票代號了（例如全部都抓到 1303）。
解法：透過 func(id string) 接收參數，並在最後 (stockNo) 傳入，這等於是把當下的股票代號複製一份給這個分身專用（這裡的 id 就是複製後的變數）。
4. defer wg.Done()：回報完成
意思：defer 代表「在這個函式結束前的最後一刻執行」。
作用：不管爬蟲成功還是失敗（甚至當機），只要這個分身要收工了，就會執行 wg.Done()，告訴主程式的計數器（WaitGroup）說：「我做完了，計數器減 1」。

總結流程圖
這段迴圈的運作邏輯如下：
迴圈開始：拿到第一個代號 2330。
wg.Add(1)：告訴計數器「多了一個任務要等」。
go func...：
派出 1號分身。
把 2330 複製一份交給它。
主程式不等待，馬上進入下一次迴圈。
迴圈繼續：拿到 2317。
wg.Add(1)。
派出 2號分身 (拿著 2317)。
...以此類推，瞬間派出 10 個分身。
wg.Wait()：主程式卡在這裡，直到計數器歸零（所有分身都呼叫了 Done），才結束程式。
這樣寫的好處是 10 檔股票會同時開始抓取，速度會比「抓完一檔再抓下一檔」快非常多。
*/
