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

	page := browser.MustPage("https://www.twse.com.tw/zh/trading/historical/stock-day-avg.html")

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
	input.MustInput("2330")

	time.Sleep(300 * time.Millisecond)

	// 點查詢
	page.MustElementR("button", "查詢").MustClick()

	// 等資料表
	page.MustWaitElementsMoreThan("tbody tr", 0)
	time.Sleep(time.Second)

	rows := page.MustElements("tbody tr")
	fmt.Println("rows:", len(rows))

	// 建立 CSV 檔案
	f, err := os.Create("stock_data.csv")
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
}
