package images

import (
	"testing"
)

func TestParseDimension(t *testing.T) {
	tests := []struct {
		name    string
		dimStr  string
		wantW   int
		wantH   int
		wantErr bool
	}{
		{
			name:    "valid normal size",
			dimStr:  "1024x768",
			wantW:   1024,
			wantH:   768,
			wantErr: false,
		},
		{
			name:    "valid small size",
			dimStr:  "1x1",
			wantW:   1,
			wantH:   1,
			wantErr: false,
		},
		{
			name:    "valid with spaces",
			dimStr:  " 1024 x 768 ",
			wantErr: true,
		},
		{
			name:    "invalid missing x",
			dimStr:  "1024",
			wantErr: true,
		},
		{
			name:    "invalid empty string",
			dimStr:  "",
			wantErr: true,
		},
		{
			name:    "invalid multiple x",
			dimStr:  "1024x768x24",
			wantErr: true,
		},
		{
			name:    "invalid width not a number",
			dimStr:  "abcx768",
			wantErr: true,
		},
		{
			name:    "invalid height not a number",
			dimStr:  "1024xabc",
			wantErr: true,
		},
		{
			name:    "invalid missing height",
			dimStr:  "1024x",
			wantErr: true,
		},
		{
			name:    "invalid missing width",
			dimStr:  "x768",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, h, err := ParseDimension(tt.dimStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDimension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if w != tt.wantW {
					t.Errorf("ParseDimension() w = %v, want %v", w, tt.wantW)
				}
				if h != tt.wantH {
					t.Errorf("ParseDimension() h = %v, want %v", h, tt.wantH)
				}
			}
		})
	}
}
