package endpoints

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"

	motmedelEnv "github.com/Motmedel/utils_go/pkg/env"
	bodyParserAdapter "github.com/Motmedel/utils_go/pkg/http/mux/interfaces/body_parser/adapter"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/endpoint_specification"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/parsing"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response"
	"github.com/Motmedel/utils_go/pkg/http/mux/types/response_error"
	jsonSchemaBodyParser "github.com/Motmedel/utils_go/pkg/http/mux/utils/json/schema"
	clientCodeGenerationTypes "github.com/altshiftab/gcp_utils/pkg/http/client_code_generation/types"
	gcpUtilsLog "github.com/altshiftab/gcp_utils/pkg/log"
	"github.com/altshiftab/x/types"
)

var Routes = routes
var TrustedTypesPolicies = []string{litHtmlTrustedTypesPolicy, webpackTrustedTypesPolicy}

var Domain = motmedelEnv.GetEnvWithDefault("DOMAIN", domain)
var Port = motmedelEnv.GetEnvWithDefault("PORT", port)
var Database *sql.DB

var EndpointSpecificationGetters []clientCodeGenerationTypes.EndpointSpecificationGetter

func init() {
	logger := gcpUtilsLog.DefaultFatal(context.Background())
	slog.SetDefault(logger.Logger)

	bodyParser, err := jsonSchemaBodyParser.New[*types.Input]()
	if err != nil {
		logger.FatalWithExitingMessage(
			"An error occurred when creating the body parser.",
			fmt.Errorf("json schema body parser new: %w", err),
		)
	}

	EndpointSpecificationGetters = []clientCodeGenerationTypes.EndpointSpecificationGetter{
		&clientCodeGenerationTypes.TypedEndpointSpecification[types.Input, any]{
			EndpointSpecification: &endpoint_specification.EndpointSpecification{
				Path:   "/api/x",
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
		},
	}
}
