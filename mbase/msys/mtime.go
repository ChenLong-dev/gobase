/*
 * @Description:
 * @Author: Chen Long
 * @Date: 2020-12-16 14:52:43
 * @LastEditTime: 2020-12-16 18:02:10
 * @LastEditors: Chen Long
 * @Reference:
 */

package msys

import (
	"fmt"
	"strings"
	"time"
)

const (
	WeekMillisecond   uint64 = 7 * 24 * 60 * 60 * 1000
	DayMillisecond    uint64 = 24 * 60 * 60 * 1000
	HourMillisecond   uint64 = 60 * 60 * 1000
	MinuteMillisecond uint64 = 60 * 1000
	SecondMillisecond uint64 = 1000
)

func NowSecond() uint64 {
	return uint64(time.Now().Unix())
}
func NowMillisecond() uint64 { //	毫秒
	return uint64(time.Now().UnixNano() / 1000000)
}
func NowNanosecond() uint64 {
	return uint64(time.Now().UnixNano())
}

func NowDate() string {
	y, m, d := time.Now().Date()
	return fmt.Sprintf("%4d%02d%02d", y, m, d)
}
func NowTime() string {
	h, m, s := time.Now().Clock()
	return fmt.Sprintf("%02d%02d%02d", h, m, s)
}
func DurationToTime(t time.Duration) time.Time {
	return time.Unix(int64(t/time.Second), int64(t%time.Second))
}
func TimeToDuration(t time.Time) time.Duration {
	return time.Duration(t.UnixNano())
}

type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

func (tr *TimeRange) String() string {
	if tr == nil {
		return ""
	}
	return PrintTime(tr.StartTime) + "~" + PrintTime(tr.EndTime)
}
func TodayTimeRange() (*TimeRange, error) {
	now := time.Now()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}
	startTime, perr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d 00:00:00", now.Year(), now.Month(), now.Day()), loc)
	if perr != nil {
		return nil, perr
	}

	return &TimeRange{StartTime: startTime, EndTime: now}, nil
}
func YesterdayTimeRange() (*TimeRange, error) {
	now := time.Now()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}
	endTime, perr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d 00:00:00", now.Year(), now.Month(), now.Day()), loc)
	if perr != nil {
		return nil, perr
	}
	//startTimestamp := endTime.Unix() - 24*60*60
	startTime := endTime.Add(-time.Hour * 24)

	return &TimeRange{StartTime: startTime, EndTime: endTime}, nil
}
func ThisWeekTimeRange() (*TimeRange, error) {
	now := time.Now()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	startTime, perr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d 00:00:00", now.Year(), now.Month(), now.Day()), loc)
	if perr != nil {
		return nil, perr
	}
	weekday := now.Weekday()
	weekday = (weekday + 7 - 1) % 7
	startTime = startTime.Add(-time.Hour * 24 * time.Duration(weekday))

	return &TimeRange{StartTime: startTime, EndTime: now}, nil
}
func LastWeekTimeRange() (*TimeRange, error) {
	now := time.Now()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	endTime, perr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d 00:00:00", now.Year(), now.Month(), now.Day()), loc)
	if perr != nil {
		return nil, perr
	}
	weekday := now.Weekday()
	weekday = (weekday + 7 - 1) % 7
	endTime = endTime.Add(-time.Hour * 24 * time.Duration(weekday))

	startTime := endTime
	startTime = startTime.Add(-time.Hour * 24 * 7)

	return &TimeRange{StartTime: startTime, EndTime: endTime}, nil
}
func ThisMonthTimeRange() (*TimeRange, error) {
	now := time.Now()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	startTime, perr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d 00:00:00", now.Year(), now.Month(), 1), loc)
	if perr != nil {
		return nil, perr
	}

	return &TimeRange{StartTime: startTime, EndTime: now}, nil
}
func LastMonthTimeRange() (*TimeRange, error) {
	now := time.Now()
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	year, month := now.Year(), now.Month()
	endTime, perr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d 00:00:00", year, month, 1), loc)
	if perr != nil {
		return nil, perr
	}

	if month > 1 {
		month--
	} else {
		year--
		month = 12
	}
	startTime, perr := time.ParseInLocation("2006-01-02 15:04:05", fmt.Sprintf("%d-%02d-%02d 00:00:00", year, month, 1), loc)
	if perr != nil {
		return nil, perr
	}

	return &TimeRange{StartTime: startTime, EndTime: endTime}, nil
}

/*2020-09-10 00:00:00[~2020-09-11 00:00:00]*/
func ParseTimeRange(strTimeRange string) (*TimeRange, error) {
	loc, err := time.LoadLocation("Local")
	if err != nil {
		return nil, err
	}

	var startTime, endTime time.Time
	ss := strings.Split(strTimeRange, "~")
	if len(ss) > 0 {
		startTime, err = time.ParseInLocation("2006-01-02 15:04:05", ss[0], loc)
		if err != nil {
			return nil, err
		}
	}
	if len(ss) > 1 {
		endTime, err = time.ParseInLocation("2006-01-02 15:04:05", ss[1], loc)
		if err != nil {
			return nil, err
		}
	}

	return &TimeRange{StartTime: startTime, EndTime: endTime}, nil
}

func PrintTimeMillisecond(ms uint64) string {
	t := time.Unix(int64(ms)*1000, 0)
	Y, M, D := t.Date()
	h, m, s := t.Clock()
	return fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d", Y, M, D, h, m, s)
}
func PrintTime(t time.Time) string {
	Y, M, D := t.Date()
	h, m, s := t.Clock()
	return fmt.Sprintf("%4d-%02d-%02d %02d:%02d:%02d", Y, M, D, h, m, s)
}
