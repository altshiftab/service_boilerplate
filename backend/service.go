package main

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"

	motmedelEnv "github.com/Motmedel/utils_go/pkg/env"
	motmedelErrors "github.com/Motmedel/utils_go/pkg/errors"
	motmedelHttpErrors "github.com/Motmedel/utils_go/pkg/http/errors"
	muxErrors "github.com/Motmedel/utils_go/pkg/http/mux/errors"
	bodyParserAdapter "github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser/adapter"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	jsonSchemaBodyParser "github.com/Motmedel/utils_go/pkg/http/mux/utils/body_parser/json/schema"
	altshiftGcpUtilsHttp "github.com/altshiftab/gcp_utils/pkg/http"
	gcpUtilsHttp "github.com/altshiftab/gcp_utils/pkg/http"
	gcpUtilsLog "github.com/altshiftab/gcp_utils/pkg/log"
)

type Input struct {
	Token string   `json:"token,omitempty" required:"true" minLength:"1"`
	_     struct{} `additionalProperties:"false"`
}

func main() {
	logger := gcpUtilsLog.DefaultFatal(context.Background())
	slog.SetDefault(logger.Logger)

	bodyParser, err := jsonSchemaBodyParser.New[*Input]()
	if err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when creating the body parser.",
			fmt.Errorf("json schema body parser new: %w", err),
		)
	}

	httpServer, mux, err := gcpUtilsHttp.MakeHttpService(
		"localhost",
		motmedelEnv.GetEnvWithDefault("PORT", "8080"),
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

	mux.Add(
		&endpoint_specification.EndpointSpecification{
			Path:   xEndpoint,
			Method: http.MethodPost,
			BodyParserConfiguration: &parsing.BodyParserConfiguration{
				ContentType: "application/json",
				MaxBytes:    4096,
				Parser:      bodyParserAdapter.New(bodyParser),
			},
			Handler: func(request *http.Request, body []byte) (*response.Response, *response_error.ResponseError) {
				return nil, nil
			},
		},
	)

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
