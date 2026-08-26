package messages

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestasPathToStringByVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		asPath  string
		want    string
	}{
		{
			name:    "legacy flat array (4.x)",
			version: "4.2.21",
			asPath:  `[30740, 30740, 30740]`,
			want:    "30740 30740 30740",
		},
		{
			name:    "segmented object (5.x)",
			version: "5.0.0",
			asPath:  `{"0": {"element": "as-sequence", "value": [64542]}}`,
			want:    "64542",
		},
		{
			name:    "segmented multi-segment ordered by index",
			version: "5.0.0",
			asPath:  `{"1": {"element": "as-set", "value": [100, 200]}, "0": {"element": "as-sequence", "value": [64542]}}`,
			want:    "64542 100 200",
		},
		{
			name:    "empty as-path",
			version: "5.0.0",
			asPath:  ``,
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := Attribute{}
			if tt.asPath != "" {
				a.ASPath = json.RawMessage(tt.asPath)
			}
			require.Equal(t, tt.want, a.asPathToString(tt.version))
		})
	}
}
