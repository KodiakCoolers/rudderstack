package prebackup

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	backendconfig "github.com/rudderlabs/rudder-server/config/backend-config"
	"github.com/rudderlabs/rudder-server/utils/pubsub"
)

// Doer does something before a jobsdb table backup happens
type Doer interface {
	Do(ctx context.Context, txn *sql.Tx, jobsTable, jobStatusTable string) error
}

// DropSourceIds provides a pre-backup doer who is responsible for removing events
// from the tables which belong to a specified list of source ids.
// The list of source ids is dynamically provided by the sourceIdsProvider
func DropSourceIds(sourceIdsProvider func() []string) Doer {
	return &dropSourceIds{
		sourceIdsProvider: sourceIdsProvider,
	}
}

type dropSourceIds struct {
	sourceIdsProvider func() []string
}

func (r *dropSourceIds) Do(ctx context.Context, txn *sql.Tx, jobsTable, jobStatusTable string) error {
	sourceIds := r.sourceIdsProvider()
	if len(sourceIds) == 0 {
		// skip
		return nil
	}
	sourceIdsParam := pq.Array(r.sourceIdsProvider())

	// First cleanup events from the job status table since it relies on the jobs table
	jsSql := fmt.Sprintf(`DELETE FROM "%[1]s" WHERE job_id IN (SELECT job_id FROM "%[2]s" WHERE parameters->>'source_id' = ANY ($1))`, jobStatusTable, jobsTable)
	jsStmt, err := txn.Prepare(jsSql)
	if err != nil {
		return err
	}
	_, err = jsStmt.ExecContext(ctx, sourceIdsParam)
	if err != nil {
		return err
	}

	// Last cleanup events from the jobs table
	jSql := fmt.Sprintf(`DELETE FROM "%[1]s" WHERE parameters->>'source_id' = ANY($1)`, jobsTable)
	jStmt, err := txn.Prepare(jSql)
	if err != nil {
		return err
	}
	_, err = jStmt.ExecContext(ctx, sourceIdsParam)
	if err != nil {
		return err
	}
	return nil
}

func BackendConfigSourceIdsProvider(ctx context.Context, config backendconfig.BackendConfig) func() []string {

	sourceIds := make([]string, 0)

	go func() {
		ch := make(chan pubsub.DataEvent)
		config.Subscribe(ch, backendconfig.TopicBackendConfig)

		for {
			select {
			case ev := <-ch:
				c := ev.Data.(backendconfig.ConfigT)
				newSourceIds := backendconfig.ExtractExcludedBackupSourceIds(c)
				sourceIds = newSourceIds
			case <-ctx.Done():
				return
			}
		}
	}()

	return func() []string {
		return sourceIds
	}
}
