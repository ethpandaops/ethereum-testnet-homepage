package homepage

import (
	"context"
	"net/http"
	"time"

	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/api"
	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/version"
	"github.com/julienschmidt/httprouter"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Server struct {
	log *logrus.Logger
	Cfg Config

	http *api.Handler
}

func NewServer(log *logrus.Logger, conf *Config) *Server {
	if err := conf.Validate(); err != nil {
		log.Fatalf("invalid config: %s", err)
	}

	s := &Server{
		Cfg: *conf,
		log: log,

		http: api.NewHandler(log, &conf.Ethereum, &conf.Homepage),
	}

	return s
}

func (s *Server) Start(ctx context.Context) error {
	// Start the api and underlying services

	if err := s.http.Start(ctx); err != nil {
		return err
	}

	s.log.Infof("Starting ethereum-testnet-homepage server (%s)", version.Short())

	router := httprouter.New()

	if err := s.http.Register(ctx, router); err != nil {
		return err
	}

	if err := s.ServeMetrics(ctx); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              s.Cfg.Global.ListenAddr,
		ReadHeaderTimeout: 15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}

	server.Handler = router

	s.log.Infof("Serving http at %s", s.Cfg.Global.ListenAddr)

	if err := server.ListenAndServe(); err != nil {
		s.log.Fatal(err)
	}

	return nil
}

func (s *Server) ServeMetrics(ctx context.Context) error {
	go func() {
		server := &http.Server{
			Addr:              s.Cfg.Global.MetricsAddr,
			ReadHeaderTimeout: 15 * time.Second,
		}

		server.Handler = promhttp.Handler()

		s.log.Infof("Serving metrics at %s", s.Cfg.Global.MetricsAddr)

		if err := server.ListenAndServe(); err != nil {
			s.log.Fatal(err)
		}
	}()

	return nil
}
