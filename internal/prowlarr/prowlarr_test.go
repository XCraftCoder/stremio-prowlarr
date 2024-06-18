package prowlarr_test

import (
	"testing"

	"github.com/bongnv/prowlarr-stremio/internal/prowlarr"
	"github.com/stretchr/testify/require"
)

func TestProwlarr_FetchMagnetURI(t *testing.T) {
	t.Run("should parse infoHash properly", func(t *testing.T) {
		var err error
		torrent := &prowlarr.Torrent{
			MagnetUri: "magnet:?xt=urn:btih:9b4c1489bfccd8205d152345f7a8aad52d9a1f57&dn=archlinux-2022.05.01-x86_64.iso",
		}
		client := prowlarr.New("", "")
		torrent, err = client.FetchMagnetURI(torrent)
		require.NoError(t, err)
		require.Equal(t, "9b4c1489bfccd8205d152345f7a8aad52d9a1f57", torrent.InfoHash)
	})
}
