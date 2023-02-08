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
			wantTabWidth: "",
			wantMaxDepth: "",
		},
		{
			name: "Short lines below min",
			BasicSurveyor: &BasicSurveyor{
				lineLengths: map[LineLengthDetail]int{
					LineLengthDetail{length: 30}: 1,
					LineLengthDetail{length: 59}: 1,
					LineLengthDetail{length: 50}: 1,
				},
			},
			wantTabWidth: "8",
			wantMaxDepth: "",
		},
		{
			name: "Short lines one in min min",
			BasicSurveyor: &BasicSurveyor{
				lineLengths: map[LineLengthDetail]int{
					LineLengthDetail{length: 30}: 1,
					LineLengthDetail{length: 61}: 1,
					LineLengthDetail{length: 50}: 1,
				},
			},
			wantTabWidth: "8",
			wantMaxDepth: "80",
		},
		{
			name: "Tab depth pushes line to min",
			BasicSurveyor: &BasicSurveyor{
				lineLengths: map[LineLengthDetail]int{
					LineLengthDetail{length: 84}: 1,
				},
			},
			wantTabWidth: "6",
			wantMaxDepth: "80",
		},
		{
			name: "A really long line", // TODO eliminate based on standard deviation
			BasicSurveyor: &BasicSurveyor{
				lineLengths: map[LineLengthDetail]int{
					LineLengthDetail{length: 30}:  1,
					LineLengthDetail{length: 168}: 1,
					LineLengthDetail{length: 50}:  1,
				},
			},
			wantTabWidth: "8",
			wantMaxDepth: "180",
		},
		{
			name: "A really long line tab pushes it under",
			BasicSurveyor: &BasicSurveyor{
				lineLengths: map[LineLengthDetail]int{
					LineLengthDetail{length: 30, tabIndentation: 2}:  1,
					LineLengthDetail{length: 168, tabIndentation: 2}: 1,
					LineLengthDetail{length: 50, tabIndentation: 2}:  1,
				},
			},
			wantTabWidth: "6",
			wantMaxDepth: "160",
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
		{
			name: "Double space",
			BasicSurveyor: &BasicSurveyor{
				whitespacePrefixes: map[string]int{
					"":         10,
					"  ":       3,
					"    ":     7,
					"      ":   3,
					"        ": 7,
				},
			},
			wantindentSize: "2",
		},
		{
			name: "Double space - misleading single",
			BasicSurveyor: &BasicSurveyor{
				whitespacePrefixes: map[string]int{
					" ":        1,
					"":         10,
					"  ":       3,
					"    ":     7,
					"      ":   3,
					"        ": 7,
				},
			},
			wantindentSize: "2",
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
