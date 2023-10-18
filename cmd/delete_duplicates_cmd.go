package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chain710/immich-cli/client"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
)

type deleteDuplicatesCmd struct {
	database   string
	dryRun     bool
	archive    bool
	concurrent int
	force      bool

	client client.ClientWithResponsesInterface
	queue  chan []string
}

type assetQuality struct {
	id    openapi_types.UUID
	score int64 // score of quality, higher is better
}

func (c *deleteDuplicatesCmd) processGroup(ctx context.Context, group []string) error {
	var errs []error
	var assets []assetQuality
	for _, assetId := range group {
		aq, err := c.getAssetQuality(ctx, assetId)
		if err != nil {
			log.Warnf("get asset %s error: %v", assetId, err)
			errs = append(errs, err)
			continue
		}
		assets = append(assets, *aq)
	}

	if len(assets) == 0 {
		return errors.Join(errs...)
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].score > assets[j].score
	})

	log.DebugFn(func() []interface{} {
		var values []any
		for _, asset := range assets {
			values = append(values, fmt.Sprintf("%s: %d,", asset.id.String(), asset.score))
		}

		return values
	})
	// keep: assets[0], delete others
	var ids []openapi_types.UUID
	for _, asset := range assets[1:] {
		ids = append(ids, asset.id)
	}

	return c.deleteAsset(ctx, ids)
}

func (c *deleteDuplicatesCmd) getAssetQuality(ctx context.Context,
	stringId string) (*assetQuality, error) {
	assetUUID, err := uuid.Parse(stringId)
	if err != nil {
		return nil, fmt.Errorf("malform uuid: `%s`", stringId)
	}
	response, err := c.client.GetAssetByIdWithResponse(ctx, assetUUID, &client.GetAssetByIdParams{})
	if err != nil {
		return nil, fmt.Errorf("get asset `%s` error: %w", stringId, err)
	}

	if response.JSON200 == nil {
		return nil, newUnexpectedResponse(response.StatusCode())
	}

	var score int64
	if response.JSON200.ExifInfo != nil && response.JSON200.ExifInfo.FileSizeInByte != nil {
		score = *response.JSON200.ExifInfo.FileSizeInByte
	}

	// prefer heic
	if strings.HasSuffix(strings.ToLower(response.JSON200.OriginalPath), ".heic") {
		score = score * 10
	}

	return &assetQuality{id: assetUUID, score: score}, nil
}

func (c *deleteDuplicatesCmd) deleteAsset(ctx context.Context, ids []openapi_types.UUID) error {
	defer func() { log.Debugf("Done!") }()
	for _, id := range ids {
		log.Infof("Should delete asset: `%s`, dryRun: %v", id.String(), c.dryRun)
	}
	if c.archive {
		log.Debugf("ready to archive assets...")
		body := client.UpdateAssetsJSONRequestBody{
			Ids:        ids,
			IsArchived: &c.archive,
		}
		response, err := c.client.UpdateAssetsWithResponse(ctx, body)
		if err != nil {
			log.Errorf("archive assets error: %v", err)
			return err
		}

		if response.StatusCode() != http.StatusNoContent {
			return newUnexpectedResponse(response.StatusCode())
		}

		return nil
	} else if !c.dryRun {
		log.Debugf("ready to delete assets...")
		body := client.DeleteAssetsJSONRequestBody{
			Force: &c.force,
			Ids:   ids,
		}
		response, err := c.client.DeleteAssetsWithResponse(ctx, body)
		if err != nil {
			log.Errorf("delete assets error: %v", err)
			return err
		}
		if response.StatusCode() != http.StatusNoContent {
			return newUnexpectedResponse(response.StatusCode())
		}

		return nil
	}

	return nil
}

func (c *deleteDuplicatesCmd) run(cmd *cobra.Command, _ []string) error {
	c.queue = make(chan []string, c.concurrent)
	c.client = newClient()

	file, err := os.Open(c.database)
	if err != nil {
		log.Errorf("open `%s` error: %v", c.database, err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	var duplicates [][]string
	if err := decoder.Decode(&duplicates); err != nil {
		log.Errorf("decode duplicates error: %v", err)
		return err
	}

	var mu sync.Mutex // protect errs
	var errs []error
	var wg sync.WaitGroup
	for i := 0; i < c.concurrent; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for group := range c.queue {
				if err := c.processGroup(cmd.Context(), group); err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			}
		}()
	}

	for _, group := range duplicates {
		c.queue <- group
	}

	close(c.queue)
	wg.Wait()
	if len(errs) > 0 {
		log.Warnf("%d error(s) occured during process, see log for more details", len(errs))
	}
	return nil
}

func DeleteDuplicatesCmd() *cobra.Command {
	impl := &deleteDuplicatesCmd{}
	cmd := &cobra.Command{
		Use:  "delete_duplicates",
		RunE: impl.run,
	}

	cmd.Flags().StringVar(&impl.database, "database", "", "duplicate database json file")
	cobra.CheckErr(cmd.MarkFlagRequired("database"))
	cmd.Flags().BoolVar(&impl.dryRun, "dry-run", false, "don't actually delete")
	cmd.Flags().BoolVar(&impl.archive, "archive", false, "archive photo instead of delete")
	cmd.Flags().IntVar(&impl.concurrent, "concurrent", 4, "num of concurrent workers")
	cmd.Flags().BoolVar(&impl.force, "force", false, "force delete")
	return cmd
}
