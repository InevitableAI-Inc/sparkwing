package jobs

import "testing"

func TestUnreleasedEntries(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{
			name: "empty section, only blank lines",
			body: "# Changelog\n\n## [Unreleased]\n\n## [v1.0.0]\n\n- old entry\n",
			want: 0,
		},
		{
			name: "empty section, only sub-headings",
			body: "# Changelog\n\n## [Unreleased]\n\n### Added\n\n### Changed\n\n## [v1.0.0]\n",
			want: 0,
		},
		{
			name: "one entry under Added",
			body: "# Changelog\n\n## [Unreleased]\n\n### Added\n\n- new thing\n\n## [v1.0.0]\n",
			want: 1,
		},
		{
			name: "entries across multiple sub-sections",
			body: "## [Unreleased]\n### Added\n- a\n- b\n### Fixed\n- c\n## [v1.0.0]\n- old\n",
			want: 3,
		},
		{
			name: "no [Unreleased] section",
			body: "# Changelog\n\n## [v1.0.0]\n\n- something\n",
			want: 0,
		},
		{
			name: "bare 'Unreleased' (no brackets) also recognized",
			body: "## Unreleased\n\n- entry\n\n## [v1.0.0]\n",
			want: 1,
		},
		{
			name: "entries in old version do not count",
			body: "## [Unreleased]\n\n## [v1.0.0]\n- old\n- older\n",
			want: 0,
		},
		{
			name: "stops at next top-level heading, ignores indented dashes",
			body: "## [Unreleased]\n\n### Added\n  -- this is not a bullet, dashed prose\n- real bullet\n## [v1.0.0]\n",
			want: 1,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := unreleasedEntries(c.body)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != c.want {
				t.Fatalf("got %d entries, want %d", got, c.want)
			}
		})
	}
}
