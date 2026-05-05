// Package snapshot persists the last known port state to a JSON file on disk.
//
// On startup, portwatch loads the snapshot to restore prior knowledge of each
// port's state, preventing spurious "state changed" notifications on restart.
//
// Usage:
//
//	snap, err := snapshot.New("/var/lib/portwatch/state.json")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	// Record a state change:
//	_ = snap.Set("localhost", 8080, "open")
//	// Retrieve last known state:
//	entry, ok := snap.Get("localhost", 8080)
package snapshot
