package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/pkg/errors"
)

func link(ctx context.Context, path string, args model.LinkArgs) (*model.Link, model.Obj, error) {
	account, actualPath, err := operations.GetAccountAndActualPath(path)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed get account")
	}
	return operations.Link(ctx, account, actualPath, args)
}
