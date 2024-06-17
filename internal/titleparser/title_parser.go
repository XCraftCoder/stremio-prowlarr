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
		parseQuality(`\b(?:HD-?)?CAM\b`),
		matchAndSetQuality(`(?i)\b(?:HD-?)?T(?:ELE)?S(?:YNC)?\b`, "telesync"),
		parseQuality(`(?i)\bHD-?Rip\b`),
		parseQuality(`(?i)\bBRRip\b`),
		parseQuality(`(?i)\bBDRip\b`),
		parseQuality(`(?i)\bDVDRip\b`),
		matchAndSetQuality(`(?i)\bDVD(?:R[0-9])?\b`, "dvd"),
		parseQuality(`(?i)\bDVDscr\b`),
		parseQuality(`(?i)\b(?:HD-?)?TVRip\b`),
		parseQuality(`\bTC\b`),
		parseQuality(`(?i)\bPPVRip\b`),
		parseQuality(`(?i)\bR5\b`),
		parseQuality(`(?i)\bVHSSCR\b`),
		matchAndSetQuality(`(?i)\bBlu-?ray Remux\b`, "brremux"),
		matchAndSetQuality(`(?i)\bBlu-?ray\b`, "bluray"),
		parseQuality(`(?i)\bWEB-?DL\b`),
		parseQuality(`(?i)\bWEB-?Rip\b`),
		parseQuality(`(?i)\b(?:DL|WEB|BD|BR)REMUX\b`),
		parseQuality(`(?i)\b(DivX|XviD)\b`),
		parseQuality(`(?i)HDTV`),
		parseCodec(`(?i)dvix|mpeg2|divx|xvid|[xh][-. ]?26[45]|avc|hevc`),
		parseAudio(`MD|MP3|mp3|FLAC|Atmos|DTS(?:-HD)?|TrueHD`),
		parseAudio(`(?i)Dual[- ]Audio`),
		matchAndSetAudio(`(?i)AC-?3(?:\.5\.1)?`, "ac3"),
		matchAndSetAudio(`(?i)DD5[. ]?1`, "dd5.1"),
		matchAndSetAudio(`(?i)AAC(?:[. ]?2[. ]0)?`, "aac"),
		parseContainer(`(?i)\b(MKV|AVI|MP4)\b`),
		parse3D(`(?i)\b((3D))\b`),
		parseSeasonAndEpisode(`(?i)S(\d{2})E(\d{2})`),
	}
)

type MetaInfo struct {
	Resolution int
	Year       int
	Quality    string
	Codec      string
	Audio      string
	Container  string
	ThreeD     bool
	Season     int
	Episode    int
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
		if len(matches) > 0 && len(matches[len(matches)-1]) > 1 {
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

func parseQuality(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findValue(&mi.Quality, title, compiled)
	}
}

func matchAndSetQuality(pattern string, value string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		findAndSet(&mi.Quality, title, compiled, value)
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

func parseSeasonAndEpisode(pattern string) func(string, *MetaInfo) {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) {
		if mi.Season > 0 {
			return
		}

		matches := compiled.FindAllStringSubmatch(title, -1)
		if len(matches) > 0 && len(matches[len(matches)-1]) > 2 {
			mi.Season, _ = strconv.Atoi(matches[len(matches)-1][1])
			mi.Episode, _ = strconv.Atoi(matches[len(matches)-1][2])
		}
	}
}
