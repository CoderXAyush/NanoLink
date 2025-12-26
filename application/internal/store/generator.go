package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/redis/go-redis/v9"
)

const (
	idKey     = "next_id_seq" // Redis key to track global ID
	blockSize = 1000          // Host reserves 1000 IDs at a time
)

type IDGenerator struct {
	redis *redis.Client
	mu    sync.Mutex
	minID uint64
	maxID uint64
	curID uint64
}

func NewIDGenerator(r *redis.Client) *IDGenerator {
	return &IDGenerator{
		redis: r,
	}
}

// GetID returns a unique ID. It fetches a new block from Redis if needed.
func (g *IDGenerator) GetID(ctx context.Context) (uint64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// If current block is exhausted, fetch a new one
	if g.curID >= g.maxID {
		if err := g.reserveBlock(ctx); err != nil {
			return 0, err
		}
	}

	g.curID++
	return g.curID, nil
}

func (g *IDGenerator) reserveBlock(ctx context.Context) error {
	// Atomic INCRBY in Redis
	end, err := g.redis.IncrBy(ctx, idKey, blockSize).Result()
	if err != nil {
		return fmt.Errorf("failed to reserve id block: %v", err)
	}

	g.maxID = uint64(end)
	g.minID = g.maxID - blockSize
	g.curID = g.minID

	fmt.Printf("📦 Reserved ID Block: %d - %d\n", g.minID+1, g.maxID)
	return nil
}
