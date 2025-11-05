package reminder

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
	"time"
)

func Test_validateEventYear(t *testing.T) {
	thisYear := strconv.Itoa(time.Now().Year())
	lastYear := strconv.Itoa(time.Now().Year() - 1)
	type args struct {
		eventYear string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "this year",
			args: args{
				eventYear: thisYear,
			},
			wantErr: false,
		},
		{
			name: "invalidInput: 20250 (invalid format)",
			args: args{
				eventYear: "20250",
			},
			wantErr: true,
		},
		{
			name: "invalidInput: 02025 (invalid format)",
			args: args{
				eventYear: "02025",
			},
			wantErr: true,
		},
		{
			name: "invalidInput: last year (invalid value: event is in the past)",
			args: args{
				eventYear: lastYear,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCustomDuration(tt.args.eventYear)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				return
			}
		})
	}
}
