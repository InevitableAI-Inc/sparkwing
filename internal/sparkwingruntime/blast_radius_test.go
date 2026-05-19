package sparkwingruntime_test

import (
	"reflect"
	"testing"

	"github.com/sparkwing-dev/sparkwing/internal/sparkwingruntime"
)

func TestSortedUniqueRisks_Dedupes(t *testing.T) {
	got := sparkwingruntime.SortedUniqueRisks(
		[]string{"prod", "destructive"},
		[]string{"prod", "money"},
		[]string{"", "destructive"},
	)
	want := []string{"destructive", "money", "prod"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("SortedUniqueRisks = %v, want %v", got, want)
	}
}
