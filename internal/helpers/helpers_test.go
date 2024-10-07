package helpers

import (
	"reflect"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "invalid email abcde",
			args: args{
				s: "abcde",
			},
			wantErr: true,
		},
		{
			name: "invalid email @",
			args: args{
				s: "@",
			},
			wantErr: true,
		},
		{
			name: "valid email tim@gmail.com",
			args: args{
				s: "tim@gmail.com",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateEmail(tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPatchPath(t *testing.T) {
	type args struct {
		parent  string
		current string
		path    []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "immutable path",
			args: args{
				parent:  "a",
				current: "b",
				path:    []string{"a", "b", "c"},
			},
			want: []string{"a", "b", "c"},
		},
		{
			name: "middle path",
			args: args{
				parent:  "x",
				current: "c",
				path:    []string{"a", "b", "c", "d"},
			},
			want: []string{"a", "b", "x", "c", "d"},
		},
		{
			name: "middle path",
			args: args{
				parent:  "x",
				current: "a",
				path:    []string{"a", "b"},
			},
			want: []string{"x", "a", "b"},
		},
		{
			name: "empty path",
			args: args{
				parent:  "x",
				current: "a",
				path:    []string{},
			},
			want: []string{"x", "a"},
		},
		{
			name: "not unique path",
			args: args{
				parent:  "a",
				current: "a",
				path:    []string{},
			},
			want: []string{"a"},
		},
		{
			name: "move parent",
			args: args{
				parent:  "z",
				current: "b",
				path:    []string{"a", "b", "z", "c"},
			},
			want: []string{"a", "z", "b", "c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PatchPath(tt.args.parent, tt.args.current, tt.args.path); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PatchPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
