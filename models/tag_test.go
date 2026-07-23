package models

import "testing"

func TestFormatTagNames(t *testing.T) {
	tests := []struct {
		name string
		in   []Tag
		want string
	}{
		{
			name: "empty",
			in:   nil,
			want: "",
		},
		{
			name: "single",
			in:   []Tag{{Name: "go"}},
			want: "go",
		},
		{
			name: "multiple",
			in:   []Tag{{Name: "go"}, {Name: "web"}, {Name: "docker"}},
			want: "go, web, docker",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTagNames(tt.in)
			if got != tt.want {
				t.Fatalf("FormatTagNames() = %q, want %q", got, tt.want)
			}
		})
	}
}
