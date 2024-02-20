package test

import (
	"github.com/dnwsilver/tld/internal/pkg/widgets"
	"testing"
)

func TestRenderResult_Crop(t *testing.T) {
	type args struct {
		length int
	}
	var tests = []struct {
		name string
		r    widgets.RenderResult
		args args
		want widgets.RenderResult
	}{
		{"Simple crop", widgets.RenderResult("1234567890"), args{length: 5}, widgets.RenderResult("1234#")},
		{"Big length", widgets.RenderResult("1234567890"), args{length: 20}, widgets.RenderResult("1234567890")},
		{"Butt length", widgets.RenderResult("1234567890"), args{length: 10}, widgets.RenderResult("1234567890")},
		{"Single unicode", widgets.RenderResult("\uF1F9 1234567890"), args{length: 5}, widgets.RenderResult("\uF1F9 12#")},
		{"Single emoji", widgets.RenderResult("💅 1234567890"), args{length: 5}, widgets.RenderResult("💅 12#")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.r.Crop(tt.args.length); got != tt.want {
				t.Errorf("Crop() = %v, want %v", got, tt.want)
			}
		})
	}
}
