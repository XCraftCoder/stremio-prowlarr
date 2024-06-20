package titleparser

import (
	"regexp"
	"strconv"
	"strings"
)

var (
	parsers = []func(string, *MetaInfo) int{
		parseYear(`(?:\b((?:19[0-9]|20[0-9])[0-9])\b)|(?:\(((?:19[0-9]|20[0-9])[0-9])\))`),
		parseResolution(`(?i)([0-9]{3,4})[pi]`),
		matchAndSetResolution(`(?i)(4k)`, 2160),
		matchAndSetQuality(`(?i)\b(?:HD-?)?CAM(?:rip)?\b`, "cam"),
		matchAndSetQuality(`(?i)\b(?:HD-?)?T(?:ELE)?S(?:YNC)?\b`, "telesync"),
		matchAndSetQuality(`(?i)\bTS-?Rip\b`, "telesync"),
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
		matchAndSetQuality(`(?i)\bBlu-?ray(?:[\s\.]|.+\b)Remux\b`, "brremux"),
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
		parseSeasonAndEpisode(`(?i)S(\d{2})-?E(\d{2})`),
		parseMultiSeason(`(?i)S(\d{2})\s*(?:to|-)?\s*S(\d{2})`),
		parseMultiSeason(`(?i)\bseason\s+(\d{1,2})[\s-]+(\d{1,2})\b`),
		parseSingleSeason(`(?i)\bs(\d{2})\b`),
		parseSingleSeason(`(?i)\bseason[- ]?(\d{1,2})\b`),
		parseLanguage(`\bFR(?:ENCH)?\b`),
		parseFillerWords(`(?i)[-\s\.\(]+\b(?:TV|Complete|Full) series\b`),
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
	FromSeason int
	ToSeason   int
	Episode    int
	Title      string
	Language   string
}

func Parse(title string) *MetaInfo {
	m := &MetaInfo{}
	index := len(title)

	for _, parser := range parsers {
		nextIndex := parser(title, m)
		if nextIndex >= 0 && nextIndex < index {
			index = nextIndex
		}
	}

	m.Title = title[0:index]

	return m
}

func findValue(value *string, title string, regex *regexp.Regexp) int {
	if *value != "" {
		// don't overwrite the existing value
		return -1
	}

	matches := regex.FindAllStringIndex(title, -1)
	if len(matches) > 0 {
		loc := matches[len(matches)-1]
		*value = strings.ToLower(title[loc[len(loc)-2]:loc[len(loc)-1]])
		return loc[0]
	}

	return -1
}

func findSubValue(value *string, title string, regex *regexp.Regexp) int {
	if *value != "" {
		// don't overwrite the existing value
		return -1
	}

	matches := regex.FindAllStringSubmatchIndex(title, -1)
	if len(matches) > 0 && len(matches[len(matches)-1]) > 3 {
		loc := matches[len(matches)-1]
		*value = strings.ToLower(title[loc[2]:loc[3]])
		return loc[0]
	}

	return -1
}

func findAndSet(value *string, title string, regex *regexp.Regexp, target string) int {
	if *value != "" {
		// don't overwrite the existing value
		return -1
	}

	matches := regex.FindAllStringIndex(title, -1)
	if len(matches) > 0 {
		loc := matches[len(matches)-1]
		*value = target
		return loc[0]
	}

	return -1
}

func parseYear(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		if mi.Year > 0 {
			return -1
		}

		var year string
		index := findValue(&year, title, compiled)
		if index != -1 {
			mi.Year, _ = strconv.Atoi(year)
		}

		return index
	}
}

func parseResolution(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		if mi.Resolution > 0 {
			return -1
		}

		var resolution string
		index := findSubValue(&resolution, title, compiled)
		if index != -1 {
			mi.Resolution, _ = strconv.Atoi(resolution)
		}

		return index
	}
}

func matchAndSetResolution(pattern string, value int) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		if mi.Resolution > 0 {
			return -1
		}

		var resolution string
		index := findValue(&resolution, title, compiled)
		if index != -1 {
			mi.Resolution = value
		}

		return index
	}
}

func parseQuality(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		return findValue(&mi.Quality, title, compiled)
	}
}

func matchAndSetQuality(pattern string, value string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		return findAndSet(&mi.Quality, title, compiled, value)
	}
}

func parseCodec(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		index := findValue(&mi.Codec, title, compiled)
		if index != -1 {
			mi.Codec = strings.ReplaceAll(mi.Codec, ".", "")
			mi.Codec = strings.ReplaceAll(mi.Codec, "-", "")
			mi.Codec = strings.ReplaceAll(mi.Codec, " ", "")
		}
		return index
	}
}

func parseAudio(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		return findValue(&mi.Audio, title, compiled)
	}
}

func matchAndSetAudio(pattern string, value string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		return findAndSet(&mi.Audio, title, compiled, value)
	}
}

func parseContainer(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		return findValue(&mi.Container, title, compiled)
	}
}

func parse3D(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		if mi.ThreeD {
			return -1
		}

		var threeD string
		index := findValue(&threeD, title, compiled)
		mi.ThreeD = index != -1
		return index
	}
}

func parseSeasonAndEpisode(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		if mi.FromSeason > 0 {
			return -1
		}

		matches := compiled.FindAllStringSubmatchIndex(title, -1)
		if len(matches) > 0 && len(matches[len(matches)-1]) > 5 {
			loc := matches[len(matches)-1]
			mi.FromSeason, _ = strconv.Atoi(title[loc[2]:loc[3]])
			mi.ToSeason = mi.FromSeason
			mi.Episode, _ = strconv.Atoi(title[loc[4]:loc[5]])
			return loc[0]
		}

		return -1
	}
}

func parseMultiSeason(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		if mi.FromSeason > 0 {
			return -1
		}

		matches := compiled.FindAllStringSubmatchIndex(title, -1)
		if len(matches) > 0 && len(matches[len(matches)-1]) > 5 {
			loc := matches[len(matches)-1]
			mi.FromSeason, _ = strconv.Atoi(title[loc[2]:loc[3]])
			mi.ToSeason, _ = strconv.Atoi(title[loc[4]:loc[5]])
			return loc[0]
		}

		return -1
	}
}

func parseSingleSeason(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		if mi.FromSeason > 0 {
			return -1
		}

		matches := compiled.FindAllStringSubmatchIndex(title, -1)
		if len(matches) > 0 && len(matches[len(matches)-1]) > 3 {
			loc := matches[len(matches)-1]
			mi.FromSeason, _ = strconv.Atoi(title[loc[2]:loc[3]])
			mi.ToSeason = mi.FromSeason
			return loc[0]
		}

		return -1
	}
}

func parseLanguage(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		return findValue(&mi.Quality, title, compiled)
	}
}

func parseFillerWords(pattern string) func(string, *MetaInfo) int {
	compiled := regexp.MustCompile(pattern)
	return func(title string, mi *MetaInfo) int {
		var filler string
		return findValue(&filler, title, compiled)
	}
}
