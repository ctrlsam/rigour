package mongodb

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ctrlsam/rigour/internal/geoip"
	"github.com/ctrlsam/rigour/internal/storage"
	"github.com/ctrlsam/rigour/pkg/enricher"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// NewRepository creates a new MongoDB repository with the provided configuration.
func (c *Client) NewRepository(ctx context.Context, cfg storage.RepositoryConfig) (storage.ServiceRepository, error) {
	if c.client == nil {
		return nil, fmt.Errorf("mongodb client: client is not initialized")
	}

	// Get or create the database
	db := c.client.Database(cfg.Database)

	// Get the collection
	coll := db.Collection(cfg.Collection)

	// Ensure indexes are created
	if err := EnsureIndexes(ctx, coll); err != nil {
		return nil, fmt.Errorf("mongodb client: failed to ensure indexes: %w", err)
	}

	return &Repository{collection: coll}, nil
}

// EnsureIndexes creates the indexes required by the repository. It's safe to
// call multiple times; MongoDB will ignore duplicate index creation.
func EnsureIndexes(ctx context.Context, coll *mongo.Collection) error {
	// Index on ip and port for faster lookups
	model := mongo.IndexModel{
		Keys: bson.D{{Key: "ip", Value: 1}, {Key: "port", Value: 1}},
	}

	// Create the index with a reasonable timeout
	idxCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := coll.Indexes().CreateOne(idxCtx, model)
	if err != nil {
		return err
	}
	return nil
}

// Close disconnects the MongoDB client.
func (c *Client) Close(ctx context.Context) error {
	if c.client != nil {
		return c.client.Disconnect(ctx)
	}
	return nil
}

// Repository implements storage.ServiceRepository for MongoDB.
type Repository struct {
	collection *mongo.Collection
}

// hostDocument is the internal structure used to decode MongoDB documents
// with BSON-specific types and field names for MongoDB storage.
type hostDocument struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	IP        string             `bson:"ip"`
	Country   string             `bson:"country,omitempty"`
	ASN       string             `bson:"asn,omitempty"`
	Services  []serviceDocument  `bson:"services,omitempty"`
	GeoIP     *geoip.GeoIPRecord `bson:"geoip,omitempty"`
	Timestamp time.Time          `bson:"timestamp"`
}

// serviceDocument is the internal structure for services in MongoDB documents.
type serviceDocument struct {
	Port      int            `bson:"port"`
	Protocol  string         `bson:"protocol"`
	TLS       bool           `bson:"tls"`
	Transport string         `bson:"transport"`
	Metadata  bsonRawMessage `bson:"metadata,omitempty"`
	Timestamp time.Time      `bson:"timestamp"`
}

// hostDocumentToHost converts a hostDocument to a storage.Host.
// Since hostDocument uses properly typed fields with BSON tags,
// the conversion is straightforward.
func hostDocumentToHost(doc hostDocument) storage.Host {
	host := storage.Host{
		ID:        doc.ID.Hex(),
		IP:        doc.IP,
		Country:   doc.Country,
		ASN:       doc.ASN,
		Timestamp: doc.Timestamp,
		GeoIP:     doc.GeoIP,
	}

	// Convert services
	for _, svc := range doc.Services {
		host.Services = append(host.Services, storage.Service{
			Port:      svc.Port,
			Protocol:  svc.Protocol,
			TLS:       svc.TLS,
			Transport: svc.Transport,
			Metadata:  json.RawMessage(svc.Metadata),
			Timestamp: svc.Timestamp,
		})
	}

	return host
}

// InsertServiceRecord inserts or updates a service record for a host.
// It creates a host document if it doesn't exist, or adds/updates a service for an existing host.
func (r *Repository) InsertServiceRecord(ctx context.Context, rec enricher.EnrichedServiceEvent) error {
	if r == nil || r.collection == nil {
		return fmt.Errorf("mongodb: repository is nil")
	}

	// Build the service document using the structured type
	service := serviceDocument{
		Port:      rec.Port,
		Protocol:  rec.Protocol,
		TLS:       rec.TLS,
		Transport: rec.Transport,
		Timestamp: rec.Timestamp,
		Metadata:  bsonRawMessage(rec.Metadata),
	}

	// Upsert host with service
	filter := bson.M{"ip": rec.IP}
	update := bson.M{
		"$set": bson.M{
			"ip":        rec.IP,
			"geoip":     rec.GeoIP,
			"timestamp": rec.Timestamp,
		},
		"$addToSet": bson.M{
			"services": service,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("mongodb: upsert service record: %w", err)
	}
	return nil
}

// Search queries hosts with filter and pagination support.
func (r *Repository) Search(ctx context.Context, filter map[string]interface{}, lastID string, limit int) ([]storage.Host, string, error) {
	if r == nil || r.collection == nil {
		return nil, "", fmt.Errorf("mongodb: repository is nil")
	}

	// Build the query filter
	query := bson.M{}
	for key, value := range filter {
		query[key] = value
	}

	// Add pagination constraint if lastID is provided
	if lastID != "" {
		query["ip"] = bson.M{"$gte": lastID}
	}

	// Set up find options with limit and sort
	findOpts := options.Find().
		SetLimit(int64(limit + 1)). // Fetch one extra to determine if there are more results
		SetSort(bson.D{{Key: "ip", Value: 1}})

	// Execute the query
	cursor, err := r.collection.Find(ctx, query, findOpts)
	if err != nil {
		return nil, "", fmt.Errorf("mongodb: failed to search: %w", err)
	}
	defer cursor.Close(ctx)

	// Decode results into hostDocument structs
	var docs []hostDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, "", fmt.Errorf("mongodb: failed to decode results: %w", err)
	}

	// Convert host documents to storage.Host, handling metadata conversion
	var hosts []storage.Host
	for _, doc := range docs {
		host := hostDocumentToHost(doc)
		hosts = append(hosts, host)
	}

	// Determine if there are more results
	var nextID string
	if len(hosts) > limit {
		// Remove the extra document
		hosts = hosts[:limit]
		nextID = hosts[len(hosts)-1].IP
	}

	return hosts, nextID, nil
}

// Facets performs aggregation for facet counts.
func (r *Repository) Facets(ctx context.Context, filter map[string]interface{}) (*storage.FacetCounts, error) {
	if r == nil || r.collection == nil {
		return nil, fmt.Errorf("mongodb: repository is nil")
	}

	agg := &storage.FacetCounts{
		Services:  make(map[string]int),
		Countries: make(map[string]int),
		ASNs:      make(map[string]int),
	}

	// Aggregate service counts
	servicePipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$unwind", Value: "$services"}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$services.protocol"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, servicePipeline)
	if err != nil {
		return nil, fmt.Errorf("mongodb: failed to aggregate services: %w", err)
	}
	defer cursor.Close(ctx)

	type result struct {
		ID    string `bson:"_id"`
		Count int    `bson:"count"`
	}

	var results []result
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("mongodb: failed to decode service results: %w", err)
	}

	for _, res := range results {
		if res.ID != "" {
			agg.Services[res.ID] = res.Count
		}
	}

	// Aggregate country counts - extract from geoip if available
	countryPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$geoip.country"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor2, err := r.collection.Aggregate(ctx, countryPipeline)
	if err == nil {
		defer cursor2.Close(ctx)

		type countResult struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}

		var countResults []countResult
		if err := cursor2.All(ctx, &countResults); err == nil {
			for _, res := range countResults {
				if res.ID != "" {
					agg.Countries[res.ID] = res.Count
				}
			}
		}
	}

	// Aggregate ASN counts - extract ASN from geoip if available
	asnPipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$geoip.organization"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor3, err := r.collection.Aggregate(ctx, asnPipeline)
	if err == nil {
		defer cursor3.Close(ctx)

		type asnResult struct {
			ID    interface{} `bson:"_id"`
			Count int         `bson:"count"`
		}

		var asnResults []asnResult
		if err := cursor3.All(ctx, &asnResults); err == nil {
			for _, res := range asnResults {
				if res.ID != nil {
					agg.ASNs[fmt.Sprintf("%v", res.ID)] = res.Count
				}
			}
		}
	}

	return agg, nil
}

var _ storage.ServiceRepository = (*Repository)(nil)
