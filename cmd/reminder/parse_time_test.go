package reminder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_parseCustomDuration(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name: "1d",
			args: args{
				s: "1d",
			},
			want:    time.Hour * 24,
			wantErr: false,
		},
		{
			name: "1w",
			args: args{
				s: "1w",
			},
			want:    time.Hour * 24 * 7,
			wantErr: false,
		},
		{
			name: "1h",
			args: args{
				s: "1h",
			},
			want:    time.Hour,
			wantErr: false,
		},
		{
			name: "1w2d3h4m",
			args: args{
				s: "1w2d3h4m",
			},
			want:    time.Hour*24*(7+2) + time.Hour*3 + time.Minute*4,
			wantErr: false,
		},
		{
			name: "100w2d3m",
			args: args{
				s: "100w2d3m",
			},
			want:    time.Hour*24*7*100 + time.Hour*24*2 + time.Minute*3,
			wantErr: false,
		},
		{
			name: "InvalidInput: h1",
			args: args{
				s: "h1",
			},
			wantErr: true,
		},
		{
			name: "InvalidInput: 1w2d-3h4m",
			args: args{
				s: "1w2d-3h4m",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCustomDuration(tt.args.s)
			if tt.wantErr {
				assert.Error(t, err, "Expected an error but got none")
				return
			}
			assert.Equal(t, tt.want, got, "Expected duration does not match")
		})
	}
}
