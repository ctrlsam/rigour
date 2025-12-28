package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/ctrlsam/rigour/internal/storage"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client manages MongoDB connections and provides methods to create repositories.
type Client struct {
	client *mongo.Client
}

// NewClient creates a new MongoDB client and connects to the server.
func NewClient(ctx context.Context, uri string, timeout time.Duration) (*Client, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	// Connect to MongoDB
	cfg := Config{URI: uri, Database: "", Timeout: timeout}
	client, _, err := Connect(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("mongodb client: failed to connect: %w", err)
	}

	return &Client{client: client}, nil
}

// Config contains connection settings for MongoDB.
type Config struct {
	URI      string
	Database string
	Timeout  time.Duration
}

// Connect creates and verifies a MongoDB client connection. It returns the
// connected client and the database (may be nil if Database is empty).
func Connect(ctx context.Context, cfg Config) (*mongo.Client, *mongo.Database, error) {
	connectCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
	defer cancel()

	clientOpts := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(connectCtx, clientOpts)
	if err != nil {
		return nil, nil, err
	}

	// Ping to ensure connection is established
	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := client.Ping(pingCtx, nil); err != nil {
		// try to disconnect on failure
		_ = client.Disconnect(context.Background())
		return nil, nil, err
	}

	var db *mongo.Database
	if cfg.Database != "" {
		db = client.Database(cfg.Database)
	}
	return client, db, nil
}

// NewHostsRepository creates a new MongoDB hosts repository with the provided configuration.
func (c *Client) NewHostsRepository(ctx context.Context, cfg storage.RepositoryConfig) (storage.HostRepository, error) {
	if c.client == nil {
		return nil, fmt.Errorf("mongodb client: client is not initialized")
	}

	db := c.client.Database(cfg.Database)
	coll := db.Collection(cfg.Collection)

	// Delegate to hosts.go to create repository with proper indexes
	return NewHostsRepository(ctx, coll)
}

// Close disconnects the MongoDB client.
func (c *Client) Close(ctx context.Context) error {
	if c.client != nil {
		return c.client.Disconnect(ctx)
	}
	return nil
}
