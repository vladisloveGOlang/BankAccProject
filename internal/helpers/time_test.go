package helpers

import (
	"testing"
	"time"
)

func TestIsTheSameDay(t *testing.T) {
	type args struct {
		a time.Time
		b time.Time
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test 1",
			args: args{
				a: time.Now(),
				b: time.Now(),
			},
			want: true,
		},
		{
			name: "Test 1",
			args: args{
				a: time.Now().Add(time.Hour * 24),
				b: time.Now(),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTheSameDay(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("IsTheSameDay() = %v, want %v", got, tt.want)
			}
		})
	}
}
