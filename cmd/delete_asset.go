package cmd

import (
	"errors"
	"github.com/chain710/immich-cli/client"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
)

func strSliceToIds(args []string) ([]openapi_types.UUID, error) {
	var errs []error
	var ids []openapi_types.UUID
	for _, arg := range args {
		if id, err := uuid.Parse(arg); err != nil {
			errs = append(errs, err)
		} else {
			ids = append(ids, id)
		}
	}

	return ids, errors.Join(errs...)
}

func DeleteAssetCmd() *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "delete_asset",
		Short: "delete_asset [id1] [id2] ...",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ids, err := strSliceToIds(args)
			if err != nil {
				log.Errorf("malformed ids: %v", err)
				return err
			}

			cli := newClient()
			body := client.DeleteAssetsJSONRequestBody{
				Force: &force,
				Ids:   ids,
			}

			log.Debugf("ready to delete %d assets", len(ids))
			response, err := cli.DeleteAssetsWithResponse(cmd.Context(), body)
			if err != nil {
				log.Errorf("delete asset error: %v", err)
				return err
			}
			if response.StatusCode() != http.StatusNoContent {
				return newUnexpectedResponse(response.StatusCode())
			}

			log.Debugf("delete ok")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "force delete")
	return cmd
}
