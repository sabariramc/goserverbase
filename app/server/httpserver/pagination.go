package httpserver

import (
	"fmt"

	"github.com/sabariramc/goserverbase/v5/utils"
)

type Filter struct {
	PageNo int64  `json:"pageNo" schema:"pageNo"`
	Limit  int64  `json:"limit" schema:"limit"`
	SortBy string `json:"sortBy" schema:"sortBy"`
	Asc    bool   `json:"asc" schema:"asc"`
}

func SetDefaultPagination(filter interface{}, defaultSortBy string) error {
	var defaultFilter Filter
	err := utils.StrictJsonTransformer(filter, &defaultFilter)
	if err != nil {
		return fmt.Errorf("baseapp.SetDefaultPagination: format mismatch: %w", err)
	}
	if defaultFilter.PageNo <= 0 {
		defaultFilter.PageNo = 1
	}
	if defaultFilter.Limit <= 0 {
		defaultFilter.Limit = 10
	}
	if defaultFilter.SortBy == "" {
		defaultFilter.SortBy = defaultSortBy
	}
	err = utils.StrictJsonTransformer(&defaultFilter, filter)
	if err != nil {
		return fmt.Errorf("baseapp.SetDefaultPagination: error transforming filter: %w", err)
	}
	return nil
}
