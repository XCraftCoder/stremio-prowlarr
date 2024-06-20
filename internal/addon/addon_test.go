package addon

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/bongnv/prowlarr-stremio/internal/debrid/realdebrid"
	"github.com/stretchr/testify/require"
)

func TestTitleCheck(t *testing.T) {
	t.Logf("Diff: %d", checkTitleSimilarity("House", "House M D"))
	require.True(t, checkTitleSimilarity("House", "Winter House") > maxTitleDistance)
	require.True(t, checkTitleSimilarity("House", "House_-_") < maxTitleDistance)
	require.True(t, checkTitleSimilarity("Mad Max: Fury Road", "Mad Max Fury Road") < maxTitleDistance)
	require.True(t, checkTitleSimilarity("House", "House M D") < maxTitleDistance)
}

func Test_Encoding(t *testing.T) {
	userData := &UserData{
		RDAPIKey: "RD API TOKEN",
	}

	jsonEncoded, err := json.Marshal(userData)
	require.NoError(t, err)
	base64Encoded := base64.RawURLEncoding.EncodeToString(jsonEncoded)
	t.Logf("Encoded: %s", base64Encoded)
}

func Test_FindingEpisodeFile(t *testing.T) {
	add := &Addon{}

	testCases := map[string]struct {
		r *streamRecord
	}{
		"should match 708": {
			r: &streamRecord{
				ContentType: ContentTypeSeries,
				Episode:     8,
				Season:      7,
				Files: []*realdebrid.File{
					{
						ID:       "match",
						FileName: "708 - Army Buddy.mkv",
					},
				},
			},
		},
		"should match S07E08": {
			r: &streamRecord{
				ContentType: ContentTypeSeries,
				Episode:     8,
				Season:      7,
				Files: []*realdebrid.File{
					{
						ID:       "match",
						FileName: "Malcolm in the Middle (2000) - S07E08 - Army Buddy (1080p AMZN WEB-DL x265 Silence).mkv",
					},
				},
			},
		},

		"should match Season 3 Episode 05": {
			r: &streamRecord{
				ContentType: ContentTypeSeries,
				Season:      3,
				Episode:     5,
				Files: []*realdebrid.File{
					{
						ID:       "match",
						FileName: "House MD Season 3 Episode 05 - Else.avi",
					},
					{
						FileName: "House MD Season 8 Episode 05 - The Confession.avi",
					},
				},
			},
		},

		"should match _S01E05_": {
			r: &streamRecord{
				ContentType: ContentTypeSeries,
				Episode:     5,
				Season:      1,
				Files: []*realdebrid.File{
					{
						ID:       "match",
						FileName: "Dr_ House_S01E05_K čertu s tebou, jestli to uděláš.mkv",
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.NotPanics(t, func() {
				result, err := add.locateMediaFile(tc.r)
				require.NoError(t, err)
				require.Len(t, result, 1)
				require.Equal(t, "match", result[0].MediaFile.ID)
			})
		})
	}
}
