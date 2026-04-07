package backends

import (
	"context"
	"database/sql"
	"fmt"
)

type contextKey string
const txHooksKey = contextKey("tx_hooks")
const txSavepointKey = contextKey("tx_savepoints")

// txState holds the transaction state inside the context.
type txState struct {
	tx       *sql.Tx
	hooks    []func()
	isNested bool
}

// Atomic executes a function within a database transaction.
// If the function returns an error, the transaction is rolled back.
// If it succeeds, the transaction is committed, and any on_commit hooks are executed.
// It supports nested atomics via savepoints.
func Atomic(ctx context.Context, dbAlias string, fn func(context.Context, *sql.Tx) error) error {
	backend, err := GetBackend(dbAlias)
	if err != nil {
		return err
	}

	state, _ := ctx.Value(txHooksKey).(*txState)

	if state != nil && state.tx != nil {
		// Nested transaction: use Savepoints
		spLevel, _ := ctx.Value(txSavepointKey).(int)
		spName := fmt.Sprintf("s%d", spLevel)

		if backend.Features().SupportsSavepoints {
			_, err = state.tx.Exec("SAVEPOINT " + spName)
			if err != nil {
				return err
			}
		}

		// Create a new context with incremented savepoint level
		newCtx := context.WithValue(ctx, txSavepointKey, spLevel+1)

		err = fn(newCtx, state.tx)
		if err != nil {
			if backend.Features().SupportsSavepoints {
				state.tx.Exec("ROLLBACK TO SAVEPOINT " + spName)
			}
			return err
		}

		if backend.Features().SupportsSavepoints {
			_, err = state.tx.Exec("RELEASE SAVEPOINT " + spName)
			if err != nil {
				return err
			}
		}

		return nil
	}

	// New transaction
	tx, err := backend.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	newState := &txState{
		tx:       tx,
		hooks:    make([]func(), 0),
		isNested: false,
	}

	newCtx := context.WithValue(ctx, txHooksKey, newState)
	newCtx = context.WithValue(newCtx, txSavepointKey, 1)

	err = fn(newCtx, tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	// Execute on_commit hooks
	for _, hook := range newState.hooks {
		hook()
	}

	return nil
}

// OnCommit registers a hook to be executed after the current transaction successfully commits.
// If there is no active transaction, the hook runs immediately.
func OnCommit(ctx context.Context, hook func()) {
	state, ok := ctx.Value(txHooksKey).(*txState)
	if ok && state != nil {
		state.hooks = append(state.hooks, hook)
	} else {
		// No active transaction, run immediately
		hook()
	}
}
