package main

import (
	"context"
	"fmt"
	"log/slog"

	motmedelEnv "github.com/Motmedel/utils_go/pkg/env"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	gcpUtilsHttp "github.com/altshiftab/gcp_utils/pkg/http"
	gcpUtilsLog "github.com/altshiftab/gcp_utils/pkg/log"
)

func main() {
	logger := gcpUtilsLog.DefaultFatal(context.Background())
	slog.SetDefault(logger.Logger)

	httpServer, _, err := gcpUtilsHttp.MakePublicHttpService(
		"localhost",
		motmedelEnv.GetEnvWithDefault("PORT", "8080"),
		staticContentEndpointSpecifications,
	)
	if err != nil {
		logger.FatalWithExitingMessage("An error occurred when making the mux.", fmt.Errorf("make mux: %w", err))
	}

	if err := httpServer.ListenAndServe(); err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when listening and serving.",
			motmedelErrors.NewWithTrace(fmt.Errorf("http server listen and serve: %w", err), httpServer),
		)
	}
}
