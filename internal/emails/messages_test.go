package emails

import (
	_ "embed"
	"strings"
	"testing"

	"github.com/krisch/crm-backend/internal/helpers"
)

func TestNewConfirmationMessage(t *testing.T) {
	fakeCode := helpers.FakeSentence(100)

	type args struct {
		code string
	}
	tests := []struct {
		name    string
		args    args
		look    string
		wantErr bool
	}{
		{
			name: "TestNewConfirmationMessage",
			args: args{
				code: "123456",
			},
			look:    "123456",
			wantErr: false,
		},
		{
			name: "TestNewConfirmationMessage",
			args: args{
				code: fakeCode,
			},
			look:    fakeCode,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewConfirmationMessage(tt.args.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewConfirmationMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !strings.Contains(got.GetBody(), tt.look) {
				t.Errorf("NewConfirmationMessage() = %v, want %v", got, tt.look)
				return
			}
		})
	}
}
