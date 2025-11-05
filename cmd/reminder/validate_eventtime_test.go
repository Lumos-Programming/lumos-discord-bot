package reminder

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_validateEventTime(t *testing.T) {
	type args struct {
		eventTime string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "11022323",
			args: args{
				eventTime: "11022323",
			},
			wantErr: false,
		},
		{
			name: "invalidInput: 13020202 (invalid value: month)",
			args: args{
				eventTime: "13020202",
			},
			wantErr: true,
		},
		{
			name: "invalidInput: 12001212 (invalid value: date)",
			args: args{
				eventTime: "12001212",
			},
			wantErr: true,
		},
		{
			name: "invalidInput: 11026002 (invalid value: hour)",
			args: args{
				eventTime: "11026002",
			},
			wantErr: true,
		},
		{
			name: "invalidInput: 11021272 (invalid value: minute)",
			args: args{
				eventTime: "11021272",
			},
			wantErr: true,
		},
		{
			name: "invalidInput: 110212723 (invalid format)",
			args: args{
				eventTime: "110212723",
			},
			wantErr: true,
		},
		{
			name: "invalidInput: 1a021212 (invalid format)",
			args: args{
				eventTime: "1a021212",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseCustomDuration(tt.args.eventTime)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				return
			}
		})
	}
}
