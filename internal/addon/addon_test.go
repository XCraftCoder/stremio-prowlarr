package addon

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/bongnv/prowlarr-stremio/internal/debrid/realdebrid"
	"github.com/stretchr/testify/require"
)

func TestTitleCheck(t *testing.T) {
	t.Logf("Diff: %f", checkTitleSimilarity("House", "Winter House"))
	t.Logf("Diff: %f", checkTitleSimilarity("House", "House_-_"))
	t.Logf("Diff: %f", checkTitleSimilarity("Mad Max: Fury Road", "Mad Max Fury Road"))
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
		r     *streamRecord
		found bool
	}{
		"should match 708": {
			r: &streamRecord{
				ContentType: ContentTypeSeries,
				Episode:     8,
				Season:      7,
				Files: []*realdebrid.File{
					{
						FileName: "708 - Army Buddy.mkv",
					},
				},
			},
			found: true,
		},
		"should match S07E08": {
			r: &streamRecord{
				ContentType: ContentTypeSeries,
				Episode:     8,
				Season:      7,
				Files: []*realdebrid.File{
					{
						FileName: "Malcolm in the Middle (2000) - S07E08 - Army Buddy (1080p AMZN WEB-DL x265 Silence).mkv",
					},
				},
			},
			found: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.NotPanics(t, func() {
				result, err := add.locateMediaFile(tc.r)
				require.NoError(t, err)
				require.Equal(t, tc.found, len(result) == 1)
			})
		})
	}
}
