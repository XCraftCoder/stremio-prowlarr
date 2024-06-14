package jackett

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-bittorrent/magneturi"
	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2/log"
)

const (
	btihPrefix = "urn:btih:"
)

func NotFollowMagnet() resty.RedirectPolicy {
	return resty.RedirectPolicyFunc(func(r1 *http.Request, _ []*http.Request) error {
		if r1.URL.Scheme == "magnet" {
			return http.ErrUseLastResponse
		}

		return nil
	})
}

// parseMagnetUri parses Magnet-formatted URIs into infoHash
func parseMagnetUri(uri string) (infoHash string, err error) {
	magnet, err := magneturi.Parse(uri)
	if err != nil {
		log.Errorf("Failed to parse magnet uri for %s due to: %v", uri, err)
		return "", err
	}

	for _, et := range magnet.ExactTopics {
		if strings.HasPrefix(et, btihPrefix) {
			return strings.TrimPrefix(et, btihPrefix), nil
		}
	}

	return "", errors.New("no info hash is found")
}
