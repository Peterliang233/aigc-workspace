package settings

// Store provides persisted runtime configuration for the app.
// Implementations must be concurrency-safe.
type Store interface {
	Get() (Settings, error)
	Update(fn func(*Settings) error) (Settings, error)
	Close() error
}
