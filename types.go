package ecg

import (
	"errors"
	"fmt"
	"github.com/alecthomas/units"
	"github.com/gabriel-vasile/mimetype"
	"github.com/saintfish/chardet"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
	"io"
	"io/fs"
	"log"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
)

// Contraster Provides contrasts two SummaryResults from the same file format for the purpose of consolidation -- future -- maybe
type Contraster func(s1, s2 *SummaryResult) int

type ErrorStringer interface {
	String() (string, error)
}

// SummaryResult results from a file format
type SummaryResult struct {
	// Impacted globs, used for reconstructing.
	FileGlobs []string
	// How confident we are not used atm
	Confidence float64
	// Future use to compare how different two file paths (same file format) are
	Contaster Contraster
	// The template
	Template ErrorStringer
	// The path, this is for future versions where it will suggest sub-directory variants based on confidence and contrast -- maybe
	Path string
	// Internal data, probably going to be used by Contraster
	Data any
}

// File reference, could also be a cache
type File struct {
	Filename   string
	FileOpener fs.FS
	size       *int64
	sync.Mutex
}

// Size of file + cache
func (fd *File) Size() int64 {
	fd.Lock()
	defer fd.Unlock()
	if fd.size != nil {
		return *fd.size
	}
	st, err := os.Stat(fd.Filename)
	if err != nil {
		return -1
	}
	s := st.Size()
	fd.size = &s
	return *fd.size
}

// Open abstracter eventually might cache, perhaps checking file size first - or only caching the first 256kb
func (fd *File) Open() (io.ReadSeekCloser, error) {
	if fd.FileOpener != nil {
		f, err := fd.FileOpener.Open(fd.Filename)
		rsc, ok := f.(io.ReadSeekCloser)
		if !ok {
			return nil, fmt.Errorf("file isn't readable, seakable or closable")
		}
		return rsc, err
	}
	return os.Open(fd.Filename)
}

func (fd *File) IsBinary() bool {
	fh, err := fd.Open()
	if err != nil {
		return false // So something else can generate the error TODO
	}
	defer func() {
		if cerr := fh.Close(); cerr != nil {
			log.Printf("Error: closing %s: %s", fd.Filename, cerr)
		}
	}()
	testBytes := make([]byte, 1*units.KiB)
	n, err := fh.Read(testBytes)
	if err != nil {
		return false
	}
	testBytes = testBytes[:n]
	detectedMIME := mimetype.Detect(testBytes)

	isBinary := true
	for mtype := detectedMIME; mtype != nil; mtype = mtype.Parent() {
		if mtype.Is("text/plain") {
			isBinary = false
		}
	}
	return isBinary
}

// FileFormat a file format
type FileFormat interface {
	// Name The display name for errors etc
	Name() string
	// Start starts reading files sent to it on the channel, will close on receiving a nil
	Start() chan *File
	// Done waits until Start() is complete, then returns the SummaryResults and/or an error
	Done() ([]*SummaryResult, error)
}

type BasicSurveyor struct {
	InsertFinalNewline     string
	Charset                string
	Charsets               string
	TrimTrailingWhitespace string
	EndOfLine              string
	Files                  int
	CharacterSets          *CharSetSummary
	finalNewLineBalance    struct {
		True  int
		False int
	}
	trailingSpaceOkay struct {
		True  int
		False int
	}
	lineEndings struct {
		Windows int
		Unix    int
	}
	lineLengths        map[LineLengthDetail]int
	IndentStyle        string
	IndentSize         string
	MaxLineLength      string
	TabWidth           string
	whitespacePrefixes map[string]int
}

func NewBasicSurveyor() *BasicSurveyor {
	return &BasicSurveyor{
		CharacterSets: &CharSetSummary{
			Sets: map[string]int{},
		},
		whitespacePrefixes: map[string]int{},
		lineLengths:        map[LineLengthDetail]int{},
	}
}

func (l *BasicSurveyor) ReadFile(fd *File) (string, bool, *LineSurvey, error) {
	l.Files += 1
	f, err := fd.Open()
	if err != nil {
		return "", false, nil, fmt.Errorf("opening %s: %w", fd.Filename, err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("Error: closing %s: %s", fd.Filename, err)
		}
	}()
	b := make([]byte, ReadSize)
	if n, err := f.Read(b); err != nil {
		return "", false, nil, fmt.Errorf("read %d (of %d) from %s: %w", n, len(b), fd.Filename, err)
	} else {
		b = b[:n]
	}
	detector := chardet.NewTextDetector()
	result, err := detector.DetectBest(b)
	if err != nil {
		if !errors.Is(err, chardet.NotDetectedError) {
			return "", false, nil, fmt.Errorf("detect character encoding from %s: %w", fd.Filename, err)
		}
	}
	var charset string
	if result != nil && result.Confidence > 80 {
		// TODO use with Summary's Confidence
		charset = result.Charset
	}
	var finalNewLine bool
	survey := LineSurveySample(b)
	if n, err := f.Seek(-1, io.SeekEnd); err != nil {
		return "", false, nil, fmt.Errorf("seak %d from %s: %w", n, fd.Filename, err)
	} else if n > 0 {
		b = make([]byte, 1)
		if n, err := f.Read(b); err != nil {
			return "", false, nil, fmt.Errorf("read %d (of %d) from %s: %w", n, len(b), fd.Filename, err)
		}
		finalNewLine = n > 0 && b[0] == '\n'
	}
	switch charset {
	case "UTF-8":
		l.CharacterSets.Utf8 += 1
	case "UTF-16BE":
		l.CharacterSets.Utf16Be += 1
	case "UTF-16LE":
		l.CharacterSets.Utf16Le += 1
	case "ISO-8859-1":
		l.CharacterSets.Latin1 += 1
	case "":
	default:
		l.CharacterSets.OtherTotal += 1
	}
	l.CharacterSets.Sets[charset] += 1
	if finalNewLine {
		l.finalNewLineBalance.True += 1
	} else {
		l.finalNewLineBalance.False += 1
	}
	if survey.LinuxNewlinesPercent() >= .80 { // 20% threshold
		l.lineEndings.Unix += 1
	} else if survey.WindowNewlinesPercent() >= .8 { // 80% threshold
		l.lineEndings.Windows += 1
	}
	if !survey.TrailingWhitespaceCommon() {
		l.trailingSpaceOkay.True += 1
	} else {
		l.trailingSpaceOkay.False += 1
	}
	for k, v := range survey.LineLengths {
		l.lineLengths[k] += v
	}
	for k, v := range survey.WhitespacePrefix {
		l.whitespacePrefixes[k] += v
	}
	return charset, finalNewLine, survey, nil
}

func (l *BasicSurveyor) Summarize() {
	if l.FinalNewLineBalanceTruePercent() > .80 {
		l.InsertFinalNewline = "true"
	} else if l.FinalNewLineBalanceFalsePercent() > .80 {
		l.InsertFinalNewline = "false"
	}
	l.Charset = l.CharacterSets.BestFit()
	l.Charsets = l.CharacterSets.Distribution(l.Files)
	if l.TrailingSpaceOkayPercent() >= .80 { // 20% threshold
		l.TrimTrailingWhitespace = "true"
	} else if l.TrailingSpaceOkayPercent() <= .20 { // 20% threshold
		l.TrimTrailingWhitespace = "false"
	}
	if l.UnixLineEndingPercent() >= .80 {
		l.EndOfLine = "lf"
	} else if l.WindowsLineEndingPercent() >= .80 {
		l.EndOfLine = "crlf"
	}
	// TODO think about mixed cases
	if v := l.TabPercent(); v >= .8 {
		l.IndentStyle = "tabs"
		l.TabWidth, l.MaxLineLength = l.TabWidthLineLengthCalc()
	} else if v <= .2 {
		l.IndentStyle = "spaces"
		l.MaxLineLength = l.SpaceMaxLineLengthCalc()
	}
	l.IndentSize = l.IndentSizeCalc()
}

func (l *BasicSurveyor) WindowsLineEndingPercent() float64 {
	if l.Files > 0 {
		return float64(l.lineEndings.Windows) / float64(l.Files)
	}
	return 0
}

func (l *BasicSurveyor) UnixLineEndingPercent() float64 {
	if l.Files > 0 {
		return float64(l.lineEndings.Unix) / float64(l.Files)
	}
	return 0
}

func (l *BasicSurveyor) FinalNewLineBalanceTruePercent() float64 {
	if l.Files > 0 {
		return float64(l.finalNewLineBalance.True) / float64(l.Files)
	}
	return 0
}

func (l *BasicSurveyor) FinalNewLineBalanceFalsePercent() float64 {
	if l.Files > 0 {
		return float64(l.finalNewLineBalance.False) / float64(l.Files)
	}
	return 0
}

func (l *BasicSurveyor) TrailingSpaceOkayPercent() float64 {
	if l.Files > 0 {
		return float64(l.trailingSpaceOkay.True) / float64(l.Files)
	}
	return 0
}

func (l *BasicSurveyor) TabPercent() float64 {
	count := 0
	total := 0
	for k, v := range l.lineLengths {
		if k.tabIndentation > 0 {
			count += v
		}
		total += v
	}
	if total == 0 {
		return 0
	}
	return float64(count) / float64(total)
}

func MinMax(array []int) (int, int) {
	max := array[0]
	min := array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

func (l *BasicSurveyor) TabWidthLineLengthCalc() (string, string) {
	if len(l.lineLengths) == 0 {
		return "", ""
	}
	type TabWidthDetail struct {
		DepthCount map[int]int
		MaxStep    int
	}
	tabWidths := map[int]*TabWidthDetail{}
	var depthKeys []int
	for i := 1; i <= 8; i++ {
		depthKeys = append(depthKeys, i)
		tabWidths[i] = &TabWidthDetail{
			DepthCount: map[int]int{},
		}
	}
	const step = 20
	const firstGoal = 80
	const minimum = firstGoal - step
	const minimumDepth = minimum / step
	for k, v := range l.lineLengths {
		for _, dk := range depthKeys {
			l := k.length + k.tabIndentation*dk
			depth := int(math.Ceil(float64(l)/float64(step))) - 1
			tabWidths[dk].DepthCount[depth] += v
			if depth > tabWidths[dk].MaxStep {
				tabWidths[dk].MaxStep = depth
			}
		}
	}
	slices.SortFunc(depthKeys, func(a, b int) bool {
		al := len(tabWidths[a].DepthCount)
		bl := len(tabWidths[b].DepthCount)
		if al != bl {
			return al < bl
		}
		if tabWidths[b].MaxStep != tabWidths[a].MaxStep {
			return tabWidths[b].MaxStep < tabWidths[a].MaxStep
		}
		return a >= b
	})
	lengths := maps.Keys(tabWidths[depthKeys[0]].DepthCount)
	sort.Sort(sort.Reverse(sort.IntSlice(lengths)))
	p := 0
	//for ;p < len(lengths) && ; p++ {
	//
	//}
	if len(lengths) > 0 && p < len(lengths) && minimumDepth <= lengths[p] {
		return fmt.Sprintf("%d", depthKeys[0]), fmt.Sprintf("%d", (lengths[p]+1)*step)
	}
	return "8", ""
}

func (l *BasicSurveyor) SpaceMaxLineLengthCalc() string {
	// TODO skip tab stuff
	d, _ := l.TabWidthLineLengthCalc()
	return d
}

func (l *BasicSurveyor) IndentSizeCalc() string {
	all := maps.Keys(l.whitespacePrefixes)
	sort.Strings(all)
	longest := 0
	for _, e := range all {
		if len(e) > longest {
			longest = len(e)
		}
	}
	var longestRun struct {
		RunLength int
		RunStr    string
		RunSize   int
	}
	for _, e := range all {
		runLength := 0
		runSize := 0
		if len(e) == 0 {
			continue
		}
		for i := len(e); i <= longest/len(e); i++ {
			k := strings.Repeat(e, i)
			if v, ok := l.whitespacePrefixes[k]; ok {
				runLength++
				runSize += v
			} else {
				runLength = 0
				runSize = 0
			}
			if runLength > longestRun.RunLength {
				longestRun.RunStr = e
				longestRun.RunLength = runLength
				longestRun.RunSize = runSize
			} else if runLength == longestRun.RunLength && runSize > longestRun.RunSize {
				longestRun.RunStr = e
				longestRun.RunLength = runLength
				longestRun.RunSize = runSize
			}
		}
	}
	if len(longestRun.RunStr) == 0 {
		return ""
	}
	return fmt.Sprintf("%d", len(longestRun.RunStr))
}

type BasicSurveyorGetter interface {
	BasicSurveyor() *BasicSurveyor
}

type BasicSurveyorSetter interface {
	SetBasicSurveyor(af *BasicSurveyor)
}
