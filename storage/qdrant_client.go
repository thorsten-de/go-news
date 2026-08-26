package storage

import (
	"context"
	"fmt"

	"github.com/qdrant/go-client/qdrant"
)

const vectorDimension = 768

// QdrantClient is a wrapper around the Qdrant vector operations.
type QdrantClient struct {
	// Client for managing collections.
	collectionsClient qdrant.CollectionsClient
	// Client for managing points (vectors) inside collections.
	pointsClient qdrant.PointsClient
	// Name of the collection storing article vectors.
	collectionName string
}

// NewQdrantClient creates a new QdrantClient with the given collection name.
// Connects via gRPC port 6334 by default on localhost.
func NewQdrantClient(ctx context.Context, address, collectionName string) (*QdrantClient, error) {
	if address == "" {
		address = "localhost"
	}
	if collectionName == "" {
		collectionName = "articles"
	}

	conn, err := qdrant.NewClient(&qdrant.Config{
		Host: address,
	})
	if err != nil {
		return nil, fmt.Errorf("connection to qdrant failed: %w", err)
	}

	client := &QdrantClient{
		collectionsClient: conn.GetCollectionsClient(),
		pointsClient:      conn.GetPointsClient(),
		collectionName:    collectionName,
	}

	if err := client.ensureCollection(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

// ensureCollection ensures the collection exists, creating it if necessary with the
// constant vector dimension and cosine distance.
func (qc *QdrantClient) ensureCollection(ctx context.Context) error {
	exists, err := qc.collectionsClient.CollectionExists(ctx, &qdrant.CollectionExistsRequest{
		CollectionName: qc.collectionName})
	if err != nil {
		return fmt.Errorf("failed to check if collection %q exists: %w", qc.collectionName, err)
	}

	if exists.Result.Exists {
		return nil
	}

	_, err = qc.collectionsClient.Create(ctx, &qdrant.CreateCollection{
		CollectionName: qc.collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorDimension,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return fmt.Errorf("failed to create collection %q: %w", qc.collectionName, err)
	}
	return nil
}

// Stores a vector in the collection with the given ID and metadata.
func (qc *QdrantClient) Store(ctx context.Context, id string, vector []float32, metadata map[string]any) error {
	payload := make(map[string]*qdrant.Value)
	for k, v := range metadata {
		payload[k] = valueToQrant(v)
	}

	// A point represents a vector with an ID and optional metadata.
	point := &qdrant.PointStruct{
		Id:      qdrant.NewID(id),
		Vectors: qdrant.NewVectors(vector...),
		Payload: payload,
	}
	_, err := qc.pointsClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: qc.collectionName,
		Points:         []*qdrant.PointStruct{point},
	})
	return err
}

// Search performs a similarity search on the collection using the given vector and returns the top matches.
func (qc *QdrantClient) Search(ctx context.Context, vector []float32, limit int) ([]SearchMatch, error) {
	response, err := qc.pointsClient.Query(ctx, &qdrant.QueryPoints{
		CollectionName: qc.collectionName,
		Query:          qdrant.NewQuery(vector...),
		Limit:          qdrant.PtrOf(uint64(limit)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	matches := make([]SearchMatch, len(response.Result))
	for i, hit := range response.Result {
		matches[i] = SearchMatch{
			ID:    hit.Id.GetUuid(),
			Score: hit.Score,
		}
	}
	return matches, nil
}

// SearchMatch represents a search result match.
type SearchMatch struct {
	// ID is the unique identifier of the match, linking to the article ID in primary storage.
	ID string
	// Score is the similarity score of the match.
	Score float32
}

// valueToQrant converts a value to a Qdrant Value.
func valueToQrant(v any) *qdrant.Value {
	switch val := v.(type) {
	case string:
		return qdrant.NewValueString(val)
	case int64:
		return qdrant.NewValueInt(val)
	case float64:
		return qdrant.NewValueDouble(val)
	case bool:
		return qdrant.NewValueBool(val)
	default:
		return qdrant.NewValueString(fmt.Sprintf("%v", v))
	}
}
