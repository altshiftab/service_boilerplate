package types

type Input struct {
	Token string   `json:"token,omitempty" required:"true" minLength:"1"`
	_     struct{} `additionalProperties:"false"`
}
