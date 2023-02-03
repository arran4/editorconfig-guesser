package ecg

import "bytes"

type LineSurvey struct {
	NewLines         int
	WhitespacePrefix map[string]int
	WhitespaceSuffix map[string]int
	WindowNewlines   int
}

func LineSurveySample(b []byte) *LineSurvey {
	ls := &LineSurvey{
		NewLines:         0,
		WhitespacePrefix: map[string]int{},
		WhitespaceSuffix: map[string]int{},
	}
	lineLength := 0
	lastLF := -1
	lastNWS := -1
	lastCR := -1
	rns := bytes.Runes(b)
	for i, r := range rns {
		lineLength++
		switch r {
		case '\n':
			lineLength = 0
			ls.NewLines++
			end := i
			if lastCR == i-1 {
				ls.WindowNewlines++
				end = lastCR
			}
			start := lastNWS + 1
			if start < 0 {
				start = 0
			}
			if end < start {
				end = start
			}
			suffix := string(rns[start:end])
			ls.WhitespaceSuffix[suffix] = ls.WhitespaceSuffix[suffix] + 1
			lastLF = i
		case '\r':
			lastCR = i
			fallthrough
		case ' ', '\t':
		default:
			if lastNWS <= lastLF {
				start := lastLF + 1
				if start < 0 {
					start = 0
				}
				end := i
				if end < start {
					end = start
				}
				suffix := string(rns[start:end])
				ls.WhitespacePrefix[suffix] = ls.WhitespacePrefix[suffix] + 1
			}
			lastNWS = i
		}
	}
	return ls
}
