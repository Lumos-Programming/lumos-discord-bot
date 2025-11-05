package reminder

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func Test_invertEventTime(t *testing.T) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("JST location not found: %v", err)
	}

	type args struct {
		eventYear string
		eventTime string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		{
			name: "2025+11062300 → 2025-11-06 23:00 JST",
			args: args{
				eventYear: "2025",
				eventTime: "11062300",
			},
			want:    time.Date(2025, 11, 6, 23, 0, 0, 0, jst),
			wantErr: false,
		},
		{
			name: "InvalidInput: 2023+01140202 (event is in the past)",
			args: args{
				eventYear: "2023",
				eventTime: "01140202",
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: 2025+00121212 (invalid value: month)",
			args: args{
				eventYear: "2025",
				eventTime: "00121212",
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: 2025+001212 (invalid format)",
			args: args{
				eventYear: "2025",
				eventTime: "001212",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := invertEventTime(tt.args.eventYear, tt.args.eventTime)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				return
			}
			assert.Equal(t, tt.want, got, "Expected time does not match")
		})
	}
}
