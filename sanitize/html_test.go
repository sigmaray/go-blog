package sanitize

import "testing"

// TestHTML_AllowsInlineCSS verifies that safe style attributes are kept
// while scripts and unsafe CSS values are stripped.
func TestHTML_AllowsInlineCSS(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "keeps color and font-size",
			in:   `<p style="color: red; font-size: 18px;">Hello</p>`,
			want: `<p style="color: red; font-size: 18px">Hello</p>`,
		},
		{
			name: "keeps text-align and margin",
			in:   `<div style="text-align: center; margin: 10px;">Centered</div>`,
			want: `<div style="text-align: center; margin: 10px">Centered</div>`,
		},
		{
			name: "strips script tags",
			in:   `<p style="color: blue;">Safe</p><script>alert(1)</script>`,
			want: `<p style="color: blue">Safe</p>`,
		},
		{
			name: "strips javascript url in background-image",
			in:   `<span style="background-image: url(javascript:alert(1)); color: green;">x</span>`,
			want: `<span style="color: green">x</span>`,
		},
		{
			name: "strips event handlers",
			in:   `<p style="color: red;" onclick="alert(1)">Click</p>`,
			want: `<p style="color: red">Click</p>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HTML(tt.in)
			if got != tt.want {
				t.Fatalf("HTML() = %q, want %q", got, tt.want)
			}
		})
	}
}
