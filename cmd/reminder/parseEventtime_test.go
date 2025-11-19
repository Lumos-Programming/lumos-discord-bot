package reminder

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func Test_parseEventtime(t *testing.T) {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	thisyear := time.Now().In(jst).Year()
	nextyear := time.Now().In(jst).Year() + 1
	thismonth := int(time.Now().In(jst).Month())
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
					"thisyear 12312359 1h",
					strconv.Itoa(thisyear),
					"12312359",
					"1h",
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
					"nextyear 01051150 1h2d",
					strconv.Itoa(nextyear),
					"01051150",
					"1h2d",
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
				},
			},
			want:    time.Date(thisyear, time.Month(thismonth), thisday, thishour, thisminute, 0, 0, jst),
			want1:   time.Date(thisyear, time.Month(thismonth), thisday, thishour, thisminute, 0, 0, jst).Add(-1 * anHour),
			wantErr: false,
		},
		{
			name: "InvalidInput: len(r.eventYear) != 4",
			args: args{
				r: ReminderInfo{
					"test: len(r.eventYear) != 4",
					"23344",
					"12312359",
					"1h",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: !isAllDigits(r.eventYear)",
			args: args{
				r: ReminderInfo{
					"test: !isAllDigits(r.eventYear)",
					"a112",
					"12312359",
					"1h",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: len(r.eventTime) != 8",
			args: args{
				r: ReminderInfo{
					"test: len(r.eventTime) != 8",
					"2000",
					"1234567",
					"1h",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: !isAllDigits(r.eventTime)",
			args: args{
				r: ReminderInfo{
					"test: !isAllDigits(r.eventTime)",
					"2000",
					"abcd1234",
					"1h",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: !(monthInput < 1 || monthInput > 12)",
			args: args{
				r: ReminderInfo{
					"test: !(monthInput < 1 || monthInput > 12)",
					"2000",
					"13312359",
					"1h",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: setTime no unit",
			args: args{
				r: ReminderInfo{
					"test: setTime no unit",
					"2000",
					"12312359",
					"1h3",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: !lastWasDigit",
			args: args{
				r: ReminderInfo{
					"test: !lastWasDigit",
					"2000",
					"12312359",
					"1wd",
				},
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: setTime wrong unit",
			args: args{
				r: ReminderInfo{
					"test: setTime wrong unit",
					"2000",
					"12312359",
					"1u",
				},
			},
			wantErr: true,
		},

		{
			name: "InvalidInput: invalid date",
			args: args{
				r: ReminderInfo{
					"test: invalid date",
					"2000",
					"02311230",
					"1h",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := parseEventtime(tt.args.r)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				return
			}
			assert.Equal(t, tt.want, got, "Expected eventTime does not match")
			assert.Equal(t, tt.want1, got1, "Expected triggerTime does not match")
		})
	}
}
