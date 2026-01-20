package main

import (
	"os"
	"reflect"
	"testing"
)

// TestReadCSV 測試讀取 CSV 的邏輯
// 使用 Table-Driven Tests (表格驅動測試) 模式
func TestReadCSV(t *testing.T) {
	// 定義測試案例結構
	tests := []struct {
		name    string       // 測試案例名稱
		content string       // 模擬的 CSV 檔案內容
		want    []SearchTask // 預期的結果
		wantErr bool         // 是否預期會發生錯誤
	}{
		{
			name: "正常讀取",
			content: `2330,112,01
2317,112,02`,
			want: []SearchTask{
				{StockNo: "2330", Year: "112", Month: "01"},
				{StockNo: "2317", Year: "112", Month: "02"},
			},
			wantErr: false,
		},
		{
			name:    "包含空白需自動去除",
			content: ` 2330 , 112 , 01 `,
			want: []SearchTask{
				{StockNo: "2330", Year: "112", Month: "01"},
			},
			wantErr: false,
		},
		{
			name:    "欄位不足應忽略",
			content: `2330,112`, // 只有兩欄，應該被忽略
			want:    nil,        // 根據程式邏輯，如果沒有 append 任何東西，slice 會是 nil
			wantErr: false,
		},
		{
			name:    "CSV 格式錯誤",
			content: `2330,"112`, // 引號未閉合
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 1. 建立暫存檔案來模擬真實檔案
			tmpfile, err := os.CreateTemp("", "test_stocks_*.csv")
			if err != nil {
				t.Fatal(err)
			}
			// 測試結束後刪除暫存檔
			defer os.Remove(tmpfile.Name())

			// 2. 寫入測試內容
			if _, err := tmpfile.Write([]byte(tt.content)); err != nil {
				t.Fatal(err)
			}
			if err := tmpfile.Close(); err != nil {
				t.Fatal(err)
			}

			// 3. 執行被測試的函式
			got, err := readCSV(tmpfile.Name())

			// 4. 驗證錯誤狀態
			if (err != nil) != tt.wantErr {
				t.Errorf("readCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// 5. 驗證回傳資料
			// 使用 DeepEqual 比較兩個 Slice 是否內容一致
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("readCSV() = %v, want %v", got, tt.want)
			}
		})
	}
}
