// Package pool manages a Kubernetes-backed pool of warm Docker
// cache PVCs. Each PVC is pre-pulled with the images jobs are likely
// to need; checking one out gives a Docker build a fully primed
// layer cache, eliminating the cold-pull tax that otherwise dominates
// CI build time.
//
// # State machine
//
// Each PVC moves through:
//
//	clean    - warmed, ready for checkout
//	in-use   - checked out by a job
//	dirty    - job returned it, needs rewarming
//	warming  - warmer pod is actively pulling images into it
//	unknown  - brand new, no state yet (treated as dirty)
//
// State is tracked via PVC annotations so the pool survives
// controller restarts. The PVC itself is the source of truth; no
// in-memory state matters.
//
// # Surface
//
// [Pool] is the runtime handle: construct via [NewPool], drive
// [Pool.Checkout] / [Pool.Return] / [Pool.Heartbeat] from the
// controller's request path, and run [ReconcileLoop] +
// [WarmingLoop] as background goroutines to keep the pool topped
// up. [Config] tunes pool size, image list, and timing; load it via
// [LoadConfig] from a ConfigMap in the controller's namespace.
package pool
