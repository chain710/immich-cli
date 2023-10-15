package cmd

import (
	"github.com/chain710/immich-cli/client"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
)

func GetAssetsCmd() *cobra.Command {
	var uuid UUID
	var isFavorite, isArchived bool
	var skip float32
	var updatedAfter Time
	cmd := &cobra.Command{
		Use: "get_assets",
		RunE: func(cmd *cobra.Command, args []string) error {
			cli := newClient(cmd)
			// ifflag set then set field
			var params client.GetAllAssetsParams
			var zeroId UUID
			if uuid != zeroId {
				params.UserId = &uuid.UUID
			}

			if !updatedAfter.IsZero() {
				params.UpdatedAfter = &updatedAfter.Time
			}
			resp, err := cli.GetAllAssetsWithResponse(cmd.Context(), &params)

			if err != nil {
				log.Errorf("GetAllAssets call error: %v", err)
				return err
			}

			if resp.StatusCode() != http.StatusOK {
				log.Errorf("Unexpected status code: %v", resp.StatusCode())
				return err
			}

			if resp.JSON200 == nil {
				return newUnexpectedResponse(resp.StatusCode())
			}

			assets := *resp.JSON200
			for _, asset := range assets {
				cmd.Printf("id: %s, updatedAt: %v\n", asset.Id, asset.UpdatedAt)
			}
			return nil
		},
	}

	cmd.Flags().VarP(&uuid, "uid", "u", "user's uuid")
	cmd.Flags().BoolVar(&isFavorite, "favorite", false, "is favorite")
	cmd.Flags().BoolVar(&isArchived, "archived", false, "is archived")
	cmd.Flags().Float32Var(&skip, "skip", skip, "skip")
	cmd.Flags().Var(&updatedAfter, "updated-after", "time in RFC3339: 2006-01-02T15:04:05Z07:00")
	return cmd
}
