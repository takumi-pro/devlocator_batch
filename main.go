package main

import (
	"database/sql"
	"devlocator_batch/database"
	"devlocator_batch/models"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	MAX_EVENTS_PER_REQUEST   = 10
	EVENT_INDEX_INCREMENT    = 10
	REQUEST_INTERVAL_SECONDS = 6
)

func main() {
	dates := getDatesUntilNextMonth(time.Now())
	stringDates := strings.Join(dates, ",")

	db, err := database.DBConnect()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		} else if err != nil {
			tx.Rollback()
		} else {
			err = tx.Commit()
		}
	}()

	if err := deleteEvents(tx); err != nil {
		fmt.Println(err.Error())
		return
	}

	startIndex := 1
	for {
		eventsResponse, err := getEvents(stringDates, startIndex, MAX_EVENTS_PER_REQUEST)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		if len(eventsResponse.Events) == 0 {
			break
		}

		if err := insertEvents(tx, eventsResponse.Events); err != nil {
			fmt.Println(err.Error())
			return
		}

		startIndex += EVENT_INDEX_INCREMENT
		time.Sleep(REQUEST_INTERVAL_SECONDS * time.Second)
	}
}

// 引数に指定した日付から1ヶ月後の日付までの日付の配列を返却する
func getDatesUntilNextMonth(today time.Time) []string {
	currentYear, currentMonth, currentDay := today.Date()

	// 翌月の最終日を求める
	nextMonthTime := time.Date(currentYear, currentMonth+2, 1, 0, 0, 0, 0, today.Location()).AddDate(0, 0, -1)
	nextYear := nextMonthTime.Year()
	nextMonth := nextMonthTime.Month()

	// 今日の日にちと翌月の最終日を比較して
	// 翌月の方が大きい場合には終了日は今日の日にちと同じ値
	// 今日の日にちの方が大きい場合には翌月の最終日が終了日となる
	endDay := nextMonthTime.Day()
	if nextMonthTime.Day() > currentDay {
		endDay = currentDay
	}

	var dates []string
	month := time.Date(nextYear, nextMonth, endDay, 0, 0, 0, 0, today.Location())
	for !today.After(month) {
		dates = append(dates, today.Format("20060102"))
		today = today.AddDate(0, 0, 1)
	}
	return dates
}

// 引数
// count - 取得件数
// date - イベント日付
// start - 検索開始位置
func getEvents(date string, startIndex int, count int) (models.EventsResponse, error) {
	url := fmt.Sprintf("https://connpass.com/api/v1/event?count=%d&start=%d&ymd=%s", count, startIndex, date)
	res, err := http.Get(url)
	if err != nil {
		return models.EventsResponse{}, err
	}
	defer res.Body.Close()

	var connpassApi models.EventsResponse
	if err := json.NewDecoder(res.Body).Decode(&connpassApi); err != nil {
		return models.EventsResponse{}, err
	}

	// オフラインのみのイベント情報を取得するため
	// 経度と緯度を判定
	var filteredEvents []models.Event
	for _, event := range connpassApi.Events {
		if event.Lat != "" && event.Lon != "" {
			filteredEvents = append(filteredEvents, event)
		}
	}
	connpassApi.Events = filteredEvents

	return connpassApi, nil
}

func deleteEvents(tx *sql.Tx) error {
	sqlStr := `DELETE FROM events;`
	if _, err := tx.Exec(sqlStr); err != nil {
		return err
	}
	return nil
}

func insertEvents(tx *sql.Tx, events []models.Event) error {
	sqlStr := `
		INSERT INTO events (
			event_id,
			title,
			catch,
			description,
			event_url,
			started_at,
			ended_at,
			` + "`limit`" + `,
			hash_tag,
			event_type,
			accepted,
			waiting,
			owner_id,
			owner_nickname,
			owner_display_name,
			place,
			address,
			lat,
			lon
		) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`
	stmt, err := tx.Prepare(sqlStr)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range events {
		if _, err := stmt.Exec(
			e.EventId, e.Title, e.Catch, e.Description, e.EventUrl, e.StartedAt,
			e.EndedAt, e.Limit, e.HashTag, e.EventType, e.Accepted, e.Waiting,
			e.OwnerId, e.OwnerNickname, e.OwnerDisplayName, e.Place, e.Address, e.Lat, e.Lon,
		); err != nil {
			return err
		}
	}

	return nil
}
