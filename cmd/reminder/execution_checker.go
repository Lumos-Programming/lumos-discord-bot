package reminder

import (
	"log"
	"strings"
	"time"
)

// RemindChecker リマインダーチェックのバックグラウンドプロセス
func (n *ReminderCmd) RemindChecker() {
	log.Printf("Started reminder executing checker")
	go func() {
		// JSTタイムゾーン
		jst, _ := time.LoadLocation("Asia/Tokyo")
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now().In(jst)
				reminderStatus.Range(func(key, value interface{}) bool {
					confirmedID := key.(string)
					status := value.(bool)
					if status {
						return true // 実行済みはスキップ
					}

					// CustomIDからYYYYMMDDHHMMを抽出
					parts := strings.Split(confirmedID, "-")
					if len(parts) != 3 || parts[0] != "reminder" {
						log.Printf("無効なCustomID形式: %s", confirmedID)
						return true
					}
					timeStr := parts[1] // YYYYMMDDHHMM
					if len(timeStr) != 12 {
						log.Printf("無効な時間形式: %s", timeStr)
						return true
					}

					// 時間パース
					triggerTime, err := time.ParseInLocation("200601021504", timeStr, jst)
					if err != nil {
						log.Printf("時間パースエラー: %v", err)
						return true
					}

					// 発火条件チェック
					if now.After(triggerTime) || now.Equal(triggerTime) {
						// 発火処理（現時点ではCustomIDをログ出力）
						log.Printf("リマインダー発火: CustomID=%s", confirmedID)

						// 実行状況をtrueに更新
						reminderStatus.Store(confirmedID, true)
					}
					return true
				})
			}
		}
	}()
}
