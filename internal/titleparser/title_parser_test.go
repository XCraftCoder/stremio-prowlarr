package titleparser_test

import (
	"testing"

	"github.com/bongnv/prowlarr-stremio/internal/titleparser"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	metaInfo := titleparser.Parse("Mad Max Fury Road 2015 2160P DV HDR10Plus Ai Enhanced H265 TrueHD Atmos 7 1 RIFE 4 15 60fps DirtyHip")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 2160, metaInfo.Resolution)

	metaInfo = titleparser.Parse("Mad Max: Fury Road (2015) 4K UHD BDRemux 2160p Dolby Vision-Rja")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, "bdremux", metaInfo.Quality)

	metaInfo = titleparser.Parse("Cloud Atlas 2012 1080p USA Blu-ray Remux AVC DTS-HD MA 5.1 -KRa")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, "brremux", metaInfo.Quality)

	metaInfo = titleparser.Parse("Summer House S08E06 Start Your Engines 720p AMZN WEB-DL DDP 2.0 H 264-NTb[TGx]")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 8, metaInfo.FromSeason)
	require.Equal(t, 6, metaInfo.Episode)

	metaInfo = titleparser.Parse("Mind Your Language - S01 to S03 - Sitcom - Xvid -Slimoo")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 1, metaInfo.FromSeason)
	require.Equal(t, 3, metaInfo.ToSeason)

	metaInfo = titleparser.Parse("House of Cards S02-E06 (2013) XviD Custom NLsubs NLtoppers")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 2, metaInfo.FromSeason)

	metaInfo = titleparser.Parse("The Great House Revival S02 COMPLETE 720p RTE WEBRip x264 GalaxyTV")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 2, metaInfo.FromSeason)

	metaInfo = titleparser.Parse("House.S02.1080p.BluRay.REMUX.AVC.DTS.5.1-NOGRP")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 2, metaInfo.FromSeason)
	require.Equal(t, "brremux", metaInfo.Quality)

	metaInfo = titleparser.Parse("House.Season-02.DvDrip.Xvid.Aquintesce")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 2, metaInfo.FromSeason)

	metaInfo = titleparser.Parse("House (2004) S01-8 S01-S08 (1080p BluRay x265 HEVC 10bit A")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 1, metaInfo.FromSeason)
	require.Equal(t, 8, metaInfo.ToSeason)

	metaInfo = titleparser.Parse("Malcolm in the Middle 2000 2006 Season 2 1080p x264 AVC")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 2, metaInfo.FromSeason)
	require.Equal(t, 2, metaInfo.ToSeason)

	metaInfo = titleparser.Parse("Malcolm in the Middle 2000 Season 1 7 S01 S07 1080p AMZN WEB-DL x265 HEVC 10bit EAC3 5.1 Silence-QxR")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 1, metaInfo.FromSeason)
	require.Equal(t, 7, metaInfo.ToSeason)

	metaInfo = titleparser.Parse("Malcolm in the Middle 2000 Season 1 7 1080p AMZN WEB-DL x265 HEVC 10bit EAC3 5.1 Silence-QxR")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, 1, metaInfo.FromSeason)
	require.Equal(t, 7, metaInfo.ToSeason)

	metaInfo = titleparser.Parse("Furiosa A Mad Max Saga 2024 1080p Cam X264 COLLECTIVE")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, "cam", metaInfo.Quality)

	metaInfo = titleparser.Parse("Furiosa A Mad Max Saga (2024) 1080p CAMRip Hindi x264  1XBET")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, "cam", metaInfo.Quality)

	metaInfo = titleparser.Parse("Mad Max Fury Road 2015 UHD BluRay 2160p TrueHD Atmos 7 1 DV HEVC HYBRID REMUX FraMeSToR")
	t.Logf("Info: %v", metaInfo)
	require.Equal(t, "brremux", metaInfo.Quality)
}
