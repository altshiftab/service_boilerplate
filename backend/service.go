package main

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"slices"

	motmedelEnv "github.com/Motmedel/utils_go/pkg/env"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	altshiftGcpUtilsHttp "github.com/altshiftab/gcp_utils/pkg/http"
	gcpUtilsHttp "github.com/altshiftab/gcp_utils/pkg/http"
	gcpUtilsLog "github.com/altshiftab/gcp_utils/pkg/log"
	"github.com/altshiftab/x/endpoints"
)

func main() {
	logger := gcpUtilsLog.DefaultFatal(context.Background())
	slog.SetDefault(logger.Logger)

	databaseUrl := motmedelEnv.GetEnvWithDefault("DATABASE_URL", "postgres://vph:change-this@127.0.0.1:5432/highlighter?sslmode=disable")

	var err error
	endpoints.Database, err = sql.Open("pgx", databaseUrl)
	if err != nil {
		logger.FatalWithExitingMessage("An error occurred when opening the database.", err)
	}
	defer func() {
		if err := endpoints.Database.Close(); err != nil {
			logger.FatalWithExitingMessage("An error occurred when closing the database.", err, endpoints.Database)
		}
	}()

	httpServer, mux, err := gcpUtilsHttp.MakeHttpService(
		endpoints.Domain,
		endpoints.Port,
		staticContentEndpointSpecifications,
	)
	if err != nil {
		logger.FatalWithExitingMessage("An error occurred when making the mux.", fmt.Errorf("make mux: %w", err))
	}
	if httpServer == nil {
		logger.FatalWithExitingMessage("The HTTP server is nil.", motmedelHttpErrors.ErrNilServer)
	}
	if mux == nil {
		logger.FatalWithExitingMessage("The mux is nil", muxErrors.ErrNilMux)
	}

	for _, endpointSpecificationGetter := range endpoints.EndpointSpecificationGetters {
		mux.Add(endpointSpecificationGetter.GetEndpointSpecification())
	}

	indexEndpointSpecification := mux.Get("/", http.MethodGet)
	indexRoutes := slices.Collect(maps.Values(endpoints.Routes))
	if err := mux.DuplicateEndpointSpecification(indexEndpointSpecification, indexRoutes...); err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when duplicating the monitor endpoint specification.",
			motmedelErrors.New(
				fmt.Errorf("duplicate endpoint specification: %w", err),
				indexEndpointSpecification,
				indexRoutes,
			),
		)
	}

	if err := altshiftGcpUtilsHttp.PatchTrustedTypes(mux, litHtmlTrustedTypesPolicy, webpackTrustedTypesPolicy); err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when patching trusted types.",
			motmedelErrors.NewWithTrace(
				fmt.Errorf("patch trusted types: %w", err),
				mux, litHtmlTrustedTypesPolicy, webpackTrustedTypesPolicy,
			),
		)
	}

	if err := httpServer.ListenAndServe(); err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when listening and serving.",
			motmedelErrors.NewWithTrace(fmt.Errorf("http server listen and serve: %w", err), httpServer),
		)
	}
}
