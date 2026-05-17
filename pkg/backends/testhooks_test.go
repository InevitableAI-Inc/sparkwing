package backends

import "sync"

func resetShimWarnedForTest() {
	shimLogStoreWarned = sync.Once{}
	shimArtStoreWarned = sync.Once{}
}
