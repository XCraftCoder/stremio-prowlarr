package titleparser_test

import (
	"testing"

	"github.com/bongnv/jackett-stremio/internal/titleparser"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	metaInfo := titleparser.Parse("Mad Max Fury Road 2015 2160P DV HDR10Plus Ai Enhanced H265 TrueHD Atmos 7 1 RIFE 4 15 60fps DirtyHip")
	require.Equal(t, 2160, metaInfo.Resolution)
}
