package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/service/ethereum"
	"github.com/ethpandaops/ethereum-testnet-homepage/pkg/service/homepage"
	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
)

// Handler is an API handler that is responsible for negotiating with a HTTP api.
// All http-level concerns should be handled in this package, with the "namespaces" (eth/ethereum-testnet-homepage)
// handling all business logic and dealing with concrete types.
type Handler struct {
	log logrus.FieldLogger

	ethereum *ethereum.Service
	homepage *homepage.Service

	metrics Metrics
}

func NewHandler(log logrus.FieldLogger, ethConfig *ethereum.Config, homepageConfig *homepage.Config) *Handler {
	namespace := "ethereum_testnet_homepage"

	return &Handler{
		log: log.WithField("module", "api"),

		ethereum: ethereum.NewService(log, namespace, ethConfig),
		homepage: homepage.NewService(log, namespace, homepageConfig),

		metrics: NewMetrics("http"),
	}
}

func (h *Handler) Start(ctx context.Context) error {
	h.log.Info("Starting API handler")

	if err := h.ethereum.Start(ctx); err != nil {
		return err
	}

	if err := h.homepage.Start(ctx); err != nil {
		return err
	}

	return nil
}

func (h *Handler) Register(ctx context.Context, router *httprouter.Router) error {
	router.GET("/api/v1/homepage/status", h.wrappedHandler(h.handleHomepageV1Status))

	router.GET("/api/v1/ethereum/nodes", h.wrappedHandler(h.handleEthereumV1Nodes))

	return nil
}

func deriveRegisteredPath(request *http.Request, ps httprouter.Params) string {
	registeredPath := request.URL.Path
	for _, param := range ps {
		registeredPath = strings.Replace(registeredPath, param.Value, fmt.Sprintf(":%s", param.Key), 1)
	}

	return registeredPath
}

func (h *Handler) wrappedHandler(handler func(ctx context.Context, r *http.Request, p httprouter.Params, contentType ContentType) (*HTTPResponse, error)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		start := time.Now()

		contentType := NewContentTypeFromRequest(r)
		ctx := r.Context()
		registeredPath := deriveRegisteredPath(r, p)

		h.log.WithFields(logrus.Fields{
			"method":       r.Method,
			"path":         r.URL.Path,
			"content_type": contentType,
			"accept":       r.Header.Get("Accept"),
		}).Debug("Handling request")

		h.metrics.ObserveRequest(r.Method, registeredPath)

		response := &HTTPResponse{}

		var err error

		defer func() {
			h.metrics.ObserveResponse(r.Method, registeredPath, fmt.Sprintf("%v", response.StatusCode), contentType.String(), time.Since(start))
		}()

		response, err = handler(ctx, r, p, contentType)
		if err != nil {
			if writeErr := WriteErrorResponse(w, err.Error(), response.StatusCode); writeErr != nil {
				h.log.WithError(writeErr).Error("Failed to write error response")
			}

			return
		}

		data, err := response.MarshalAs(contentType)
		if err != nil {
			if writeErr := WriteErrorResponse(w, err.Error(), http.StatusInternalServerError); writeErr != nil {
				h.log.WithError(writeErr).Error("Failed to write error response")
			}

			return
		}

		for header, value := range response.Headers {
			w.Header().Set(header, value)
		}

		if err := WriteContentAwareResponse(w, data, contentType); err != nil {
			h.log.WithError(err).Error("Failed to write response")
		}
	}
}

func (h *Handler) handleHomepageV1Status(ctx context.Context, r *http.Request, p httprouter.Params, contentType ContentType) (*HTTPResponse, error) {
	if err := ValidateContentType(contentType, []ContentType{ContentTypeJSON}); err != nil {
		return NewUnsupportedMediaTypeResponse(nil), err
	}

	status, err := h.homepage.Status(ctx, &homepage.StatusRequest{})
	if err != nil {
		return NewInternalServerErrorResponse(nil), err
	}

	rsp := NewSuccessResponse(ContentTypeResolvers{
		ContentTypeJSON: func() ([]byte, error) {
			return json.Marshal(status)
		},
	})

	rsp.SetCacheControl("public, s-max-age=30")

	return rsp, nil
}

func (h *Handler) handleEthereumV1Nodes(ctx context.Context, r *http.Request, p httprouter.Params, contentType ContentType) (*HTTPResponse, error) {
	if err := ValidateContentType(contentType, []ContentType{ContentTypeJSON}); err != nil {
		return NewUnsupportedMediaTypeResponse(nil), err
	}

	nodes, err := h.ethereum.Nodes(ctx, &ethereum.NodesRequest{})
	if err != nil {
		return NewInternalServerErrorResponse(nil), err
	}

	rsp := NewSuccessResponse(ContentTypeResolvers{
		ContentTypeJSON: func() ([]byte, error) {
			return json.Marshal(nodes)
		},
	})

	rsp.SetCacheControl("public, s-max-age=30")

	return rsp, nil
}
