package domain

import (
	"testing"
)

func TestNewStatusGraph(t *testing.T) {
	json := `{ "0" : ["10"], "10": ["2","3","11"], "11": ["12"], "3": ["12"] }`

	graph, err := NewStatusGraphFromJSON(json)
	if err != nil {
		t.Errorf("Error on NewStatusGraphFromJson: %v", err)
	}

	tests := []struct {
		name        string
		root        string
		destination string
		want        bool
	}{
		{
			name:        "Move to 5",
			root:        "0",
			destination: "5",
			want:        false,
		},
		{
			name:        "Move to 99",
			root:        "0",
			destination: "99",
			want:        false,
		},
		{
			name:        "Move to 12",
			root:        "0",
			destination: "12",
			want:        true,
		},
		{
			name:        "Move to x",
			root:        "0",
			destination: "xxxx",
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := CheckPathByValue(graph, tt.root, tt.destination); got != tt.want {
				t.Errorf("StatusGraph (%v) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}
