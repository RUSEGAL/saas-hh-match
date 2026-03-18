package user

import (
	"fmt"
	"time"

	"go-api/internal/cache"
	dbuser "go-api/internal/repository/user"
)

const statsCacheTTL = 5 * time.Minute

func GetUserStats(userID int64) (*dbuser.UserStats, error) {
	if cache.Client != nil {
		cacheKey := fmt.Sprintf("user:stats:%d", userID)
		var stats dbuser.UserStats
		if err := cache.Get(cacheKey, &stats); err == nil {
			return &stats, nil
		}
	}

	stats, err := dbuser.GetUserStats(userID)
	if err != nil {
		return nil, err
	}

	if cache.Client != nil {
		cache.Set(fmt.Sprintf("user:stats:%d", userID), stats, statsCacheTTL)
	}

	return stats, nil
}

func GetPaymentStatus(userID int64) (*dbuser.PaymentStatus, error) {
	cacheKey := fmt.Sprintf("user:payment:%d", userID)

	if cache.Client != nil {
		var status dbuser.PaymentStatus
		if err := cache.Get(cacheKey, &status); err == nil {
			return &status, nil
		}
	}

	status, err := dbuser.GetPaymentStatus(userID)
	if err != nil {
		return nil, err
	}

	if cache.Client != nil {
		cache.Set(cacheKey, status, statsCacheTTL)
	}

	return status, nil
}
