package ecg

import "testing"

func TestBasicSurveyor_TabWidthLineLengthCalc(t *testing.T) {
	tests := []struct {
		name          string
		BasicSurveyor *BasicSurveyor
		want          string
		want1         string
	}{
		{
			name: "empty",
			BasicSurveyor: &BasicSurveyor{
				lineLengths: map[LineLengthDetail]int{},
			},
			want:  "",
			want1: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := tt.BasicSurveyor.TabWidthLineLengthCalc()
			if got != tt.want {
				t.Errorf("TabWidthLineLengthCalc() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("TabWidthLineLengthCalc() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestBasicSurveyor_IndentSizeCalc(t *testing.T) {
	tests := []struct {
		name          string
		BasicSurveyor *BasicSurveyor
		want          string
	}{
		{
			name: "empty",
			BasicSurveyor: &BasicSurveyor{
				whitespacePrefixes: map[string]int{},
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.BasicSurveyor.IndentSizeCalc(); got != tt.want {
				t.Errorf("IndentSizeCalc() = %v, want %v", got, tt.want)
			}
		})
	}
}
