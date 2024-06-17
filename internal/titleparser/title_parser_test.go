package titleparser_test

import (
	"testing"

	"github.com/bongnv/jackett-stremio/internal/titleparser"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	metaInfo := titleparser.Parse("Mad Max Fury Road 2015 2160P DV HDR10Plus Ai Enhanced H265 TrueHD Atmos 7 1 RIFE 4 15 60fps DirtyHip")
	require.Equal(t, 2160, metaInfo.Resolution)

	metaInfo = titleparser.Parse("Mad Max: Fury Road (2015) 4K UHD BDRemux 2160p Dolby Vision-Rja")
	require.Equal(t, "bdremux", metaInfo.Quality)

	metaInfo = titleparser.Parse("Cloud Atlas 2012 1080p USA Blu-ray Remux AVC DTS-HD MA 5.1 -KRa")
	require.Equal(t, "brremux", metaInfo.Quality)

	metaInfo = titleparser.Parse("Summer House S08E06 Start Your Engines 720p AMZN WEB-DL DDP 2.0 H 264-NTb[TGx]")
	require.Equal(t, 8, metaInfo.Season)
	require.Equal(t, 6, metaInfo.Episode)

	// metaInfo = titleparser.Parse("Mind Your Language - S01 to S03 - Sitcom - Xvid -Slimoo")
	// require.Equal(t, 1, metaInfo.Season)
}
