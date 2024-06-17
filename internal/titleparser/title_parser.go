package titleparser

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	parsers = []func(string, *MetaInfo){
		parseYear(`\b(((?:19[0-9]|20[0-9])[0-9]))\b`),
		parseResolution(`(?i)([0-9]{3,4})[pi]`),
		matchAndSetResolution(`(?i)(4k)`, 2160),
		parseSource(`\b(?:HD-?)?CAM\b`),
		matchAndSetSource(`(?i)\b(?:HD-?)?T(?:ELE)?S(?:YNC)?\b`, "telesync"),
		parseSource(`(?i)\bHD-?Rip\b`),
		parseSource(`(?i)\bBRRip\b`),
		parseSource(`(?i)\bBDRip\b`),
		parseSource(`(?i)\bDVDRip\b`),
		matchAndSetSource(`(?i)\bDVD(?:R[0-9])?\b`, "dvd"),
		parseSource(`(?i)\bDVDscr\b`),
		parseSource(`(?i)\b(?:HD-?)?TVRip\b`),
		parseSource(`\bTC\b`),
		parseSource(`(?i)\bPPVRip\b`),
		parseSource(`(?i)\bR5\b`),
		parseSource(`(?i)\bVHSSCR\b`),
		parseSource(`(?i)\bBluray\b`),
		parseSource(`(?i)\bWEB-?DL\b`),
		parseSource(`(?i)\bWEB-?Rip\b`),
		parseSource(`(?i)\b(?:DL|WEB|BD|BR)REMUX\b`),
		parseSource(`(?i)\b(DivX|XviD)\b`),
		parseSource(`(?i)HDTV`),
		parseCodec(`(?i)dvix|mpeg2|divx|xvid|[xh][-. ]?26[45]|avc|hevc`),
		parseAudio(`MD|MP3|mp3|FLAC|Atmos|DTS(?:-HD)?|TrueHD`),
		parseAudio(`(?i)Dual[- ]Audio`),
		matchAndSetAudio(`(?i)AC-?3(?:\.5\.1)?`, "ac3"),
		matchAndSetAudio(`(?i)DD5[. ]?1`, "dd5.1"),
		matchAndSetAudio(`(?i)AAC(?:[. ]?2[. ]0)?`, "aac"),
		parseContainer(`(?i)\b(MKV|AVI|MP4)\b`),
		parse3D(`(?i)\b((3D))\b`),
	}
)

type MetaInfo struct {
	Resolution int
	Year       int
	Source     string
	Codec      string
	Audio      string
	Container  string
	ThreeD     bool
}

func Parse(title string) *MetaInfo {
	m := &MetaInfo{}

	for _, parser := range parsers {
		parser(title, m)
	}
	return m
}

func findValue(value *string, title string, regex *regexp.Regexp) {
	if *value != "" {
		// don't overwrite the existing value
		return
	}

	matches := regex.FindAllString(title, -1)
	if len(matches) > 0 {
		*value = strings.ToLower(matches[len(matches)-1])
	}
}

func findAndSet(value *string, title string, regex *regexp.Regexp, target string) {
	if *value != "" {
		// don't overwrite the existing value
		return
	}

	if regex.MatchString(title) {
		*value = target
	}
}

func parseYear(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		if mi.Year > 0 {
			return
		}

		matches := compiled.FindAllString(title, -1)
		if len(matches) > 0 {
			mi.Year, _ = strconv.Atoi(matches[len(matches)-1])
		}
	}
}

func parseResolution(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		if mi.Resolution > 0 {
			return
		}

		matches := compiled.FindAllStringSubmatch(title, -1)
		if len(matches) > 0 {
			mi.Resolution, _ = strconv.Atoi(matches[len(matches)-1][1])
		}
	}
}

func matchAndSetResolution(pattern string, value int) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		if mi.Resolution > 0 {
			return
		}

		if compiled.MatchString(title) {
			mi.Resolution = value
		}
	}
}

func parseSource(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findValue(&mi.Source, title, compiled)
	}
}

func matchAndSetSource(pattern string, value string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findAndSet(&mi.Source, title, compiled, value)
	}
}

func parseCodec(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findValue(&mi.Codec, title, compiled)
		mi.Codec = strings.ReplaceAll(mi.Codec, ".", "")
		mi.Codec = strings.ReplaceAll(mi.Codec, "-", "")
		mi.Codec = strings.ReplaceAll(mi.Codec, " ", "")
	}
}

func parseAudio(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findValue(&mi.Audio, title, compiled)
	}
}

func matchAndSetAudio(pattern string, value string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findAndSet(&mi.Audio, title, compiled, value)
	}
}

func parseContainer(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findValue(&mi.Container, title, compiled)
	}
}

func parse3D(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		if mi.ThreeD {
			return
		}

		mi.ThreeD = compiled.MatchString(title)
	}
}
