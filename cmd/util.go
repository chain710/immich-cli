package cmd

import (
	"context"
	"github.com/chain710/immich-cli/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"net/http"
)

func newClient(cmd *cobra.Command) client.ClientWithResponsesInterface {
	api := viper.GetString(ViperKey_API)
	key := viper.GetString(ViperKey_APIKey)

	options := []client.ClientOption{
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-Api-Key", key)
			return nil
		}),
		client.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			log.Debugf("%s: `%s` headers: %v", req.Method, req.URL.String(), req.Header)
			return nil
		}),
	}

	cli, err := client.NewClientWithResponses(api, options...)

	if err != nil {
		log.Fatalf("create immich client error: %v", err)
	}

	return cli
}
