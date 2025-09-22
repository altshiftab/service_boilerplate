package main

import (
	"context"
	_ "embed"
	"encoding/json"
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
	motmedelJson "github.com/Motmedel/utils_go/pkg/json"
	motmedelJsonSchema "github.com/Motmedel/utils_go/pkg/json/schema"
	gcpUtilsHttp "github.com/altshiftab/gcp_utils/pkg/http"
	gcpUtilsLog "github.com/altshiftab/gcp_utils/pkg/log"
)

//go:embed config.json
var configData []byte

type Input struct {
	Token string   `json:"token,omitempty" required:"true" minLength:"1"`
	_     struct{} `additionalProperties:"false"`
}

func getConfig() (*Config, error) {
	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json unmarshal: %w", err), configData)
	}

	configJsonSchema, err := motmedelJsonSchema.New(config)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("json schema new: %w", err), config)
	}

	configMap, err := motmedelJson.ObjectToMap(config)
	if err != nil {
		return nil, motmedelErrors.NewWithTrace(fmt.Errorf("object to map: %w", err), config)
	}

	evaluationResult := configJsonSchema.Validate(configMap)
	if evaluationResult == nil {
		return nil, motmedelErrors.NewWithTrace(jsonSchemaBodyParser.ErrNilEvaluationResult)
	}

	evaluationResultList := evaluationResult.ToList()
	if evaluationResultList == nil {
		return nil, motmedelErrors.NewWithTrace(jsonSchemaBodyParser.ErrNilEvaluationResultList)
	}

	errorsMap := evaluationResultList.Errors
	if len(errorsMap) != 0 {
		return nil, motmedelErrors.ErrValidationError
	}

	return &config, nil
}

func main() {
	logger := gcpUtilsLog.DefaultFatal(context.Background())
	slog.SetDefault(logger.Logger)

	config, err := getConfig()
	if err != nil {
		logger.FatalWithExitingMessage("An error occurred when getting the config.", err)
	}
	if config == nil {
		logger.FatalWithExitingMessage("The config is nil.", nil)
	}

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
			Path:   config.XEndpoint,
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

	if err := httpServer.ListenAndServe(); err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when listening and serving.",
			motmedelErrors.NewWithTrace(fmt.Errorf("http server listen and serve: %w", err), httpServer),
		)
	}
}
