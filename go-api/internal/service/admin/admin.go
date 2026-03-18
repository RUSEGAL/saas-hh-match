package admin

import (
	"errors"
	"fmt"
	"time"

	"go-api/internal/cache"
	dbadmin "go-api/internal/repository/admin"
	types_internal "go-api/internal/types/int"
)

const (
	CacheAllStats     = 2 * time.Minute
	CacheUserStats    = 3 * time.Minute
	CacheUsers        = 5 * time.Minute
	CacheUserResumes  = 5 * time.Minute
	CacheUserPayments = 5 * time.Minute
)

var ErrCacheMiss = errors.New("cache miss")

func GetUsers() ([]types_internal.User, error) {
	var users []types_internal.User
	if err := cache.Get("admin:users", &users); err == nil {
		return users, nil
	}

	data, err := dbadmin.GetAllUsers()
	if err != nil {
		return nil, err
	}
	cache.Set("admin:users", data, CacheUsers)
	return data, nil
}

func GetStats(userID *int64) ([]types_internal.UserStats, *types_internal.UserStats, error) {
	if userID != nil {
		cacheKey := fmt.Sprintf("admin:stats:user:%d", *userID)
		var stats types_internal.UserStats
		if err := cache.Get(cacheKey, &stats); err == nil {
			return nil, &stats, nil
		}

		data, err := dbadmin.GetUserStats(*userID)
		if err != nil {
			return nil, nil, err
		}
		if data == nil {
			return nil, nil, nil
		}
		cache.Set(cacheKey, data, CacheUserStats)
		return nil, data, nil
	}

	var stats []types_internal.UserStats
	if err := cache.Get("admin:stats:all", &stats); err == nil {
		return stats, nil, nil
	}

	data, err := dbadmin.GetAllStats()
	if err != nil {
		return nil, nil, err
	}
	cache.Set("admin:stats:all", data, CacheAllStats)
	return data, nil, nil
}

func GetResumes(userID int64) ([]types_internal.ResumeWithUser, error) {
	cacheKey := fmt.Sprintf("admin:resumes:%d", userID)
	var resumes []types_internal.ResumeWithUser
	if err := cache.Get(cacheKey, &resumes); err == nil {
		return resumes, nil
	}

	data, err := dbadmin.GetUserResumes(userID)
	if err != nil {
		return nil, err
	}
	cache.Set(cacheKey, data, CacheUserResumes)
	return data, nil
}

func GetPayments(userID int64) ([]types_internal.PaymentWithUser, error) {
	cacheKey := fmt.Sprintf("admin:payments:%d", userID)
	var payments []types_internal.PaymentWithUser
	if err := cache.Get(cacheKey, &payments); err == nil {
		return payments, nil
	}

	data, err := dbadmin.GetUserPayments(userID)
	if err != nil {
		return nil, err
	}
	cache.Set(cacheKey, data, CacheUserPayments)
	return data, nil
}

func InvalidateUserCache(userID int64) {
	cache.Delete(fmt.Sprintf("admin:stats:user:%d", userID))
	cache.Delete(fmt.Sprintf("admin:resumes:%d", userID))
	cache.Delete(fmt.Sprintf("admin:payments:%d", userID))
	cache.Delete("admin:stats:all")
	cache.Delete("admin:users")
}
