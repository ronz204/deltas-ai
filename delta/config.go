package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/ronz204/deltas-ai/package/domain/conffile"
	"github.com/ronz204/deltas-ai/package/gateway/diskbase"
)

func loadConf(root string) conffile.Conf {
	vault := diskbase.New(root)
	var problems []string
	read := func(name string) string {
		text, err := vault.ReadFile(context.Background(), name)
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			problems = append(problems, fmt.Sprintf("%s could not be read: %v", name, err))
		}
		return string(text)
	}
	c := conffile.Parse(read(conffile.BudgetName), read(conffile.TestingName))
	c.Problems = append(problems, c.Problems...)
	return c
}
