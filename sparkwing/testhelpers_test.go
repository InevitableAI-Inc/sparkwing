package sparkwing

// setDebug lives only at test build time. Production callers cannot
// flip the debug flag at runtime; SPARKWING_DEBUG at process start
// is the only supported toggle.
func setDebug(on bool) { debugEnabled.Store(on) }
