package cmd

import (
	"github.com/chain710/immich-cli/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"net/http"
)

func GetAssetsCmd() *cobra.Command {
	paramsFlagSet := pflag.NewFlagSet("", pflag.ContinueOnError)
	addFlagSetByFormFields(&client.GetAllAssetsParams{}, paramsFlagSet)

	cmd := &cobra.Command{
		Use: "get_assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := newClient()
			var params client.GetAllAssetsParams
			if err := setFormFields(&params, paramsFlagSet); err != nil {
				return err
			}

			resp, err := cli.GetAllAssetsWithResponse(cmd.Context(), &params)

			if err != nil {
				log.Errorf("GetAllAssets call error: %v", err)
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				log.Errorf("Unexpected status code: %v", resp.StatusCode())
				return newUnexpectedResponse(resp.StatusCode())
			}

			if resp.JSON200 == nil {
				log.Warnf("empty data\n")
				return nil
			}

			assets := *resp.JSON200
			for _, asset := range assets {
				cmd.Printf("id: %s, updatedAt: %v\n", asset.Id, asset.UpdatedAt)
			}
			return nil
		},
	}

	cmd.Flags().AddFlagSet(paramsFlagSet)
	return cmd
}
