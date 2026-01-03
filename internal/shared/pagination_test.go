package shared

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNewListFilterFromRequest(t *testing.T) {
	tests := []struct {
		name       string
		query      url.Values
		wantErr    bool
		wantLimit  int
		wantOffset int
		wantSort   SortOrder
	}{
		{"defaults", url.Values{}, false, DefaultLimit, 0, SortAsc},
		{"custom limit", url.Values{"$limit": {"50"}}, false, 50, 0, SortAsc},
		{"custom offset", url.Values{"$offset": {"10"}}, false, DefaultLimit, 10, SortAsc},
		{"sort desc", url.Values{"sort": {"desc"}}, false, DefaultLimit, 0, SortDesc},
		{"sort asc", url.Values{"sort": {"asc"}}, false, DefaultLimit, 0, SortAsc},
		{"invalid sort", url.Values{"sort": {"invalid"}}, true, 0, 0, ""},
		{"limit too high", url.Values{"$limit": {"200"}}, true, 0, 0, ""},
		{"limit zero", url.Values{"$limit": {"0"}}, true, 0, 0, ""},
		{"limit negative", url.Values{"$limit": {"-1"}}, true, 0, 0, ""},
		{"offset negative", url.Values{"$offset": {"-1"}}, true, 0, 0, ""},
		{"limit not integer", url.Values{"$limit": {"abc"}}, true, 0, 0, ""},
		{"offset not integer", url.Values{"$offset": {"abc"}}, true, 0, 0, ""},
		{"nome too long", url.Values{"nome": {string(make([]byte, 101))}}, true, 0, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &url.URL{RawQuery: tt.query.Encode()}
			r := &http.Request{URL: u}

			filter, err := NewListFilterFromRequest(r)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewListFilterFromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if filter.Limit != tt.wantLimit {
					t.Errorf("Limit = %v, want %v", filter.Limit, tt.wantLimit)
				}
				if filter.Offset != tt.wantOffset {
					t.Errorf("Offset = %v, want %v", filter.Offset, tt.wantOffset)
				}
				if filter.Sort != tt.wantSort {
					t.Errorf("Sort = %v, want %v", filter.Sort, tt.wantSort)
				}
			}
		})
	}
}
