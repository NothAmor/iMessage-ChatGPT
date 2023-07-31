package global

import (
	"context"

	"github.com/peacecwz/go-mac-imessage/config"
	"github.com/sashabaranov/go-openai"
)

var (
	Config   config.Config
	Context  context.Context
	AIClient *openai.Client
)
