package auth

import (
	"strconv"

	"github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
)

// PruneTokensCommand is `auth:prune-tokens`: deletes used and expired email links (scheduled daily).
type PruneTokensCommand struct{}

func (r *PruneTokensCommand) Signature() string   { return "auth:prune-tokens" }
func (r *PruneTokensCommand) Description() string { return "Delete used and expired email links" }
func (r *PruneTokensCommand) Extend() command.Extend {
	return command.Extend{Category: "auth"}
}

func (r *PruneTokensCommand) Handle(ctx console.Context) error {
	removed, err := PruneTokens()
	if err != nil {
		ctx.Error(err.Error())
		return err
	}
	ctx.Success("removed " + strconv.FormatInt(removed, 10) + " link(s)")
	return nil
}
