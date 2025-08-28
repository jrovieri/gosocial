package store

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

type PaginatedFeedQuery struct {
	Limit  int      `json:"limit" validate:"gte=1,lte=20"`
	Offset int      `json:"offset" validate:"gte=0"`
	Sort   string   `json:"sort" validate:"oneof=asc desc"`
	Tags   []string `json:"tags" validate:"max=5"`
	Search string   `json:"search" validate:"max=100"`
	Since  string   `json:"since"`
	Until  string   `json:"until"`
}

func (q PaginatedFeedQuery) Parse(r *http.Request) (PaginatedFeedQuery, error) {

	queryStr := r.URL.Query()

	limit := queryStr.Get("limit")
	if limit != "" {
		value, err := strconv.Atoi(limit)
		if err != nil {
			return q, nil
		}
		q.Limit = value
	}

	offset := queryStr.Get("offset")
	if offset != "" {
		value, err := strconv.Atoi(offset)
		if err != nil {
			return q, err
		}
		q.Offset = value
	}

	sort := queryStr.Get("sort")
	if sort != "" {
		q.Sort = sort
	}

	tags := queryStr.Get("tags")
	if tags != "" {
		q.Tags = strings.Split(tags, ",")
	}

	search := queryStr.Get("search")
	if search != "" {
		q.Search = search
	}

	since := queryStr.Get("since")
	if since != "" {
		q.Since = parseTime(since)
	}

	until := queryStr.Get("until")
	if until != "" {
		q.Until = parseTime(until)
	}
	return q, nil
}

func parseTime(s string) string {
	t, err := time.Parse(time.DateTime, s)
	if err != nil {
		return ""
	}
	return t.Format(time.DateTime)
}
