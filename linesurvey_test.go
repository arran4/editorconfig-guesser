package ecg

import (
	"github.com/google/go-cmp/cmp"
	"testing"
)

func TestLineSurveySample(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		want *LineSurvey
	}{
		{name: "Empty string", b: []byte{}, want: &LineSurvey{
			NewLines:         0,
			WhitespacePrefix: map[string]int{},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "Just some text nothing interesting", b: []byte("Just some text nothing interesting"), want: &LineSurvey{
			NewLines: 0,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "One word per line, unix, no ws", b: []byte("one\nword\nper\nline"), want: &LineSurvey{
			NewLines: 3,
			WhitespacePrefix: map[string]int{
				"": 4,
			},
			WhitespaceSuffix: map[string]int{
				"": 3,
			},
			WindowNewlines: 0,
			LineLengths: map[LineLengthDetail]int{
				LineLengthDetail{length: 3}: 2,
				LineLengthDetail{length: 4}: 1,
			},
		}},
		{name: "One word per line, windows, no ws", b: []byte("one\r\nword\r\nper\r\nline"), want: &LineSurvey{
			NewLines: 3,
			WhitespacePrefix: map[string]int{
				"": 4,
			},
			WhitespaceSuffix: map[string]int{
				"": 3,
			},
			WindowNewlines: 3,
			LineLengths: map[LineLengthDetail]int{
				LineLengthDetail{length: 3}: 2,
				LineLengthDetail{length: 4}: 1,
			},
		}},
		{name: "A couple spaces then a token", b: []byte("    token"), want: &LineSurvey{
			NewLines: 0,
			WhitespacePrefix: map[string]int{
				"    ": 1,
			},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "A couple tabs then a token", b: []byte("\t\ttoken"), want: &LineSurvey{
			NewLines: 0,
			WhitespacePrefix: map[string]int{
				"\t\t": 1,
			},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "A couple tabs and spaces then a token", b: []byte("\t  \ttoken"), want: &LineSurvey{
			NewLines: 0,
			WhitespacePrefix: map[string]int{
				"\t  \t": 1,
			},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "A token then a couple spaces", b: []byte("token  "), want: &LineSurvey{
			NewLines: 0,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "A token then a couple spaces then a new line", b: []byte("token  \n"), want: &LineSurvey{
			NewLines: 1,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{
				"  ": 1,
			},
			WindowNewlines: 0,
			LineLengths: map[LineLengthDetail]int{
				LineLengthDetail{length: 7}: 1,
			},
		}},
		{name: "A token then a couple tabs", b: []byte("token\t\t"), want: &LineSurvey{
			NewLines: 0,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "A token then a couple tabs then a new line", b: []byte("token\t\t\n"), want: &LineSurvey{
			NewLines: 1,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{
				"\t\t": 1,
			},
			WindowNewlines: 0,
			LineLengths: map[LineLengthDetail]int{
				LineLengthDetail{length: 7}: 1,
			},
		}},
		{name: "A token then a couple tabs then a windows new line", b: []byte("token\t\t\r\n"), want: &LineSurvey{
			NewLines: 1,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{
				"\t\t": 1,
			},
			WindowNewlines: 1,
			LineLengths: map[LineLengthDetail]int{
				LineLengthDetail{length: 7}: 1,
			},
		}},
		{name: "A token then a couple spaces and tabs", b: []byte("token\t  \t"), want: &LineSurvey{
			NewLines: 0,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{},
			WindowNewlines:   0,
			LineLengths:      map[LineLengthDetail]int{},
		}},
		{name: "A token then a couple spaces and tabs then a new line", b: []byte("token\t  \t\n"), want: &LineSurvey{
			NewLines: 1,
			WhitespacePrefix: map[string]int{
				"": 1,
			},
			WhitespaceSuffix: map[string]int{
				"\t  \t": 1,
			},
			WindowNewlines: 0,
			LineLengths: map[LineLengthDetail]int{
				LineLengthDetail{length: 9}: 1,
			},
		}},
		{name: "One word per line, mixed", b: []byte("one\r\n\tword\t\n\tper \r\n line \t"), want: &LineSurvey{
			NewLines: 3,
			WhitespacePrefix: map[string]int{
				"":   1,
				" ":  1,
				"\t": 2,
			},
			WhitespaceSuffix: map[string]int{
				"":   1,
				" ":  1,
				"\t": 1,
			},
			WindowNewlines: 2,
			LineLengths: map[LineLengthDetail]int{
				LineLengthDetail{length: 3}:                    1,
				LineLengthDetail{length: 6, tabIndentation: 1}: 1,
				LineLengthDetail{length: 5, tabIndentation: 1}: 1,
			},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LineSurveySample(tt.b)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("LineSurveySample() = \n%s", diff)
			}
		})
	}
}
