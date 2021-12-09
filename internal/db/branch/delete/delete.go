package delete

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/supabase/cli/internal/utils"
)

func Run(branch string) error {
	if err := utils.AssertSupabaseStartIsRunning(); err != nil {
		return err
	}

	if utils.IsBranchNameReserved(branch) {
		return errors.New("Cannot delete branch " + utils.Aqua(branch) + ": branch name is reserved.")
	}

	if _, err := os.ReadDir("supabase/.branches/" + branch); err != nil {
		return errors.New("Branch " + utils.Aqua(branch) + " does not exist.")
	}

	if err := os.RemoveAll("supabase/.branches/" + branch); err != nil {
		return fmt.Errorf("Failed deleting branch %s: %w", utils.Aqua(branch), err)
	}

	{
		// https://dba.stackexchange.com/a/11895
		out, err := utils.DockerExec(context.Background(), utils.DbId, []string{
			"sh", "-c", "psql --username postgres --host localhost <<'EOSQL' " +
				"&& dropdb --force --username postgres --host localhost '" + branch + `'
BEGIN;
` + fmt.Sprintf(utils.TerminateDbSqlFmt, branch) + `
COMMIT;
EOSQL
`,
		})
		if err != nil {
			return err
		}
		var errBuf bytes.Buffer
		if _, err := stdcopy.StdCopy(io.Discard, &errBuf, out); err != nil {
			return err
		}
		if errBuf.Len() > 0 {
			return errors.New("Error dropping database: " + errBuf.String())
		}
	}

	fmt.Println("Deleted branch " + utils.Aqua(branch) + ".")
	return nil
}