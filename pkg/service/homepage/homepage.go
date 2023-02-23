package homepage

import (
	"context"

	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/version"
	"github.com/sirupsen/logrus"
)

// Service is the homepage Service handler. HTTP-level concerns should NOT be contained in this package,
// they should be handled and reasoned with at a higher level.
type Service struct {
	log    logrus.FieldLogger
	config *Config
}

// NewHandler returns a new Handler instance.
func NewService(log logrus.FieldLogger, namespace string, conf *Config) *Service {
	return &Service{
		log:    log.WithField("module", "service/homepage"),
		config: conf,
	}
}

func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Starting homepage service")

	return nil
}

// Status returns the status for homepage.
func (s *Service) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	response := &StatusResponse{
		BrandName:     s.config.BrandName,
		BrandImageURL: s.config.BrandImageURL,
		Version: Version{
			Full:      version.FullVWithGOOS(),
			Short:     version.Short(),
			GitCommit: version.GitCommit,
			Release:   version.Release,
		},
	}

	return response, nil
}
