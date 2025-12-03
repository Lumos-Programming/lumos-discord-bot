package reminder

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func TestReminderInfo_validate(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	thisyear := time.Now().In(jst).Year()
	nextyear := thisyear + 1
	thismonth := int(time.Now().In(jst).Month())
	lastmonth := thismonth - 1
	thisday := time.Now().In(jst).Day()
	thishour := time.Now().In(jst).Hour()
	thisminute := time.Now().In(jst).Minute()
	type args struct {
		r ReminderInfo
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		want1   time.Time
		wantErr bool
	}{
		{
			name: "ValidInput: thisyear 12312359 1h",
			args: args{
				r: ReminderInfo{
					"test: thisyear 12312359 1h",
					strconv.Itoa(thisyear),
					"12312359",
					"1h",
					[]int{0, 0, 0},
					"",
				},
			},
			want:    time.Date(thisyear, time.December, 31, 23, 59, 0, 0, jst),
			want1:   time.Date(thisyear, time.December, 31, 22, 59, 0, 0, jst),
			wantErr: false,
		},
		{
			name: "ValidInput: nextyear 01051150 1h2d",
			args: args{
				r: ReminderInfo{
					"test: nextyear 01051150 1h2d",
					strconv.Itoa(nextyear),
					"01051150",
					"1h2d",
					[]int{0, 0, 0},
					"",
				},
			},
			want:    time.Date(nextyear, time.January, 5, 11, 50, 0, 0, jst),
			want1:   time.Date(nextyear, time.January, 3, 10, 50, 0, 0, jst),
			wantErr: false,
		},
		{
			name: "ValidInput: now",
			args: args{
				r: ReminderInfo{
					"test: now",
					strconv.Itoa(thisyear),
					strconv.Itoa(thismonth) + strconv.Itoa(thisday) + strconv.Itoa(thishour) + strconv.Itoa(thisminute),
					"1h",
					[]int{0, 0, 0},
					"",
				},
			},
			//want:    time.Date(thisyear, time.Month(thismonth), thisday, thishour, thisminute, 0, 0, jst),
			//want1:   time.Date(thisyear, time.Month(thismonth), thisday, thishour, thisminute, 0, 0, jst).Add(-1 * anHour),
			wantErr: true, //プログラム上では完全な同時刻ならエラーなし．モーダルからの入力だと，分までしか入力できず，秒以下が0に設定されるので，イベントの日時=現在の日時の入力にしても基本的にエラーが出る
		},
		{
			name: "InvalidInput: year is in past",
			args: args{
				r: ReminderInfo{
					"test: year is in past",
					"2023",
					"12312359",
					"1h",
					[]int{0, 0, 0},
					"",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: event is in past",
			args: args{
				r: ReminderInfo{
					"test: event is in past",
					strconv.Itoa(thisyear),
					strconv.Itoa(lastmonth) + "122300",
					"1h",
					[]int{0, 0, 0},
					"",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := tt.args.r.validate()
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				return
			}
			assert.Equal(t, tt.want, got, "Expected eventTime does not match")
			assert.Equal(t, tt.want1, got1, "Expected triggerTime does not match")
		})
	}
}
