package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
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

	for _, stockNo := range stockList {
		scrapeStock(browser, stockNo)
	}
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
		// fmt.Println(cols)
		if err := w.Write(cols); err != nil {
			panic(err)
		}
	}
	fmt.Printf("已儲存至 %s\n\n", fileName)
}
