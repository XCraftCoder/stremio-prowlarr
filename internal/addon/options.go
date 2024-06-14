package addon

import (
	"github.com/bongnv/jackett-stremio/internal/debrid/realdebrid"
	"github.com/bongnv/jackett-stremio/internal/jackett"
)

func WithID(id string) Option {
	return func(a *Addon) {
		a.id = id
	}
}

func WithName(name string) Option {
	return func(a *Addon) {
		a.name = name
	}
}

func WithJackett(jacketUrl string, jacketApiKey string) Option {
	return func(a *Addon) {
		a.jackettClient = jackett.New(jacketUrl, jacketApiKey)
	}
}

func WithRealDebrid(realDebridToken string) Option {
	return func(a *Addon) {
		a.realDebrid = realdebrid.New(realDebridToken)
	}
}
