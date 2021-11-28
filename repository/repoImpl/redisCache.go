package repoimpl

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/thteam47/server_management/global"
)

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}
func SetKeyToListKeyCache(key string, listCache string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	keyCache := listCache
	var keyList []string
	keyCacheStatus := global.MyRediscache.Get(ctx, keyCache, &keyList)
	if keyCacheStatus != nil {
		keyList = append(keyList, key)
	} else {
		if !stringInSlice(key, keyList) {
			keyList = append(keyList, key)
		}
	}

	if err := global.MyRediscache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   keyCache,
		Value: keyList,
	}); err != nil {
		panic(err)
	}
}
func RemoveListKeyCache(MyRediscache *cache.Cache, keyListCache string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var keyIndex []string
	keyCacheIndex := MyRediscache.Get(ctx, keyListCache, &keyIndex)
	if keyCacheIndex == nil {
		for _, key := range keyIndex {
			RemoveValueCache(MyRediscache,key)
		}
	}
}
func GetValueCache(MyRediscache *cache.Cache, key string, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := MyRediscache.Get(ctx, key, result)
	if err != nil {
		return err
	}
	return nil
}
func SetValueCache(MyRediscache *cache.Cache, key string, data interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := global.MyRediscache.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: data,
	}); err != nil {
		return err
	}
	return nil
}
func RemoveValueCache(MyRediscache *cache.Cache, key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := MyRediscache.Delete(ctx, key)
	if err != nil {
		return err
	}
	return nil
}
