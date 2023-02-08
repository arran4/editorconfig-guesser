package ecg

import "testing"

func TestBasicSurveyor_TabWidthLineLengthCalc(t *testing.T) {
	tests := []struct {
		name          string
		BasicSurveyor *BasicSurveyor
		wantTabWidth  string
		wantMaxDepth  string
	}{
		{
			name: "empty",
			BasicSurveyor: &BasicSurveyor{
				lineLengths: map[LineLengthDetail]int{},
			},
			wantTabWidth: "8",
			wantMaxDepth: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tabWidth, maxDepth := tt.BasicSurveyor.TabWidthLineLengthCalc()
			if tabWidth != tt.wantTabWidth {
				t.Errorf("TabWidthLineLengthCalc() tabWidth = %v, want %v", tabWidth, tt.wantTabWidth)
			}
			if maxDepth != tt.wantMaxDepth {
				t.Errorf("TabWidthLineLengthCalc() maxDepth = %v, want %v", maxDepth, tt.wantMaxDepth)
			}
		})
	}
}

func TestBasicSurveyor_IndentSizeCalc(t *testing.T) {
	tests := []struct {
		name           string
		BasicSurveyor  *BasicSurveyor
		wantindentSize string
	}{
		{
			name: "empty",
			BasicSurveyor: &BasicSurveyor{
				whitespacePrefixes: map[string]int{},
			},
			wantindentSize: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if indentSize := tt.BasicSurveyor.IndentSizeCalc(); indentSize != tt.wantindentSize {
				t.Errorf("IndentSizeCalc() = %v, want %v", indentSize, tt.wantindentSize)
			}
		})
	}
}
