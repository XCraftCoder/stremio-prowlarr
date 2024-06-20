package addon

import (
	"github.com/bongnv/prowlarr-stremio/internal/prowlarr"
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

func WithProwlarr(jacketUrl string, jacketApiKey string) Option {
	return func(a *Addon) {
		a.prowlarrClient = prowlarr.New(jacketUrl, jacketApiKey)
	}
}

func WithDevelopment(development bool) Option {
	return func(a *Addon) {
		a.development = development
	}
}

func WithVersion(version string) Option {
	return func(a *Addon) {
		a.version = version
	}
}
