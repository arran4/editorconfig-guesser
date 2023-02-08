package ecg

import "bytes"

type LineLengthDetail struct{ length, tabIndentation int }

type LineSurvey struct {
	NewLines         int
	WhitespacePrefix map[string]int
	WhitespaceSuffix map[string]int
	WindowNewlines   int
	LineLengths      map[LineLengthDetail]int
}

func (survey *LineSurvey) TrailingWhitespaceCommon() bool {
	// TODO this is too simplistic -- I don't know what circumstance this would be for.... Request PR / case study
	return len(survey.WhitespaceSuffix) > 5
}

func (survey *LineSurvey) LinuxNewlinesPercent() float64 {
	if survey.NewLines == 0 {
		return 0
	}
	return float64(survey.NewLines-survey.WindowNewlines) / float64(survey.NewLines)
}

func (survey *LineSurvey) WindowNewlinesPercent() float64 {
	if survey.NewLines == 0 {
		return 0
	}
	return float64(survey.WindowNewlines) / float64(survey.NewLines)
}

func LineSurveySample(b []byte) *LineSurvey {
	ls := &LineSurvey{
		NewLines:         0,
		WhitespacePrefix: map[string]int{},
		WhitespaceSuffix: map[string]int{},
		LineLengths:      map[LineLengthDetail]int{},
	}
	lineLength := 0
	lineTabCount := 0
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
			count := lineTabCount
			if count == -1 {
				count = 0
			}
			ls.LineLengths[LineLengthDetail{length: (end) - (lastLF + 1), tabIndentation: count}] += 1
			lineTabCount = 0
			lastLF = i
		case '\t':
			if lastNWS <= lastLF {
				lineTabCount++
			}
		case '\r':
			lastCR = i
			if lastNWS <= lastLF {
				lineTabCount = 0
			}
		case ' ':
			if lastNWS <= lastLF {
				lineTabCount = 0
			}
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
				if lineTabCount != end-start {
					lineTabCount = -1
				}
				prefix := string(rns[start:end])
				ls.WhitespacePrefix[prefix] = ls.WhitespacePrefix[prefix] + 1
			}
			lastNWS = i
		}
	}
	return ls
}
