package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"crypto-analytics/internal/models"

	"github.com/redis/go-redis/v9"
)

type AnalysisTempRStorage struct {
	rdb *redis.Client
}

func NewAnalysisTempStorage(client *redis.Client) *AnalysisTempRStorage {
	return &AnalysisTempRStorage{
		rdb: client,
	}
}

func (a *AnalysisTempRStorage) generateKey(pair, timeframe string) string {
	return fmt.Sprintf("%s:%s", pair, timeframe)
}

func (a *AnalysisTempRStorage) SaveAnalysisData(data models.AnalysisData) error {
	key := a.generateKey(data.Pair, data.Timeframe)
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = a.rdb.Set(ctx, key, jsonData, time.Hour).Err()
	if err != nil {
		return fmt.Errorf("failed to save to redis: %w", err)
	}

	return nil
}
func (a *AnalysisTempRStorage) SavePairs(pairs models.PairsCrypto) error {
	for _, data := range pairs {
		err := a.SaveAnalysisData(data)
		if err != nil {
			return fmt.Errorf("failed to save pair %s:%s: %w",
				data.Pair, data.Timeframe, err)
		}
	}
	return nil
}

func (a *AnalysisTempRStorage) GetAnalysisData(pair, timeframe string) (*models.AnalysisData, error) {
	key := a.generateKey(pair, timeframe)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	jsonData, err := a.rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get data from redis: %w", err)
	}

	var data models.AnalysisData
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", err)
	}

	return &data, nil
}
func (a *AnalysisTempRStorage) GetStats() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	keys, err := a.rdb.Keys(ctx, "*:*").Result()
	if err != nil {
		return fmt.Sprintf("Error: %v", err)
	}

	if len(keys) == 0 {
		return "Redis is empty"
	}

	return fmt.Sprintf("%s %d", strings.Join(keys, " "), len(keys))
}

func (a *AnalysisTempRStorage) Close(client *redis.Client) {
	client.Close()
}
