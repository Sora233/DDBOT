/*
 * block cache 是设计给重请求小容量cache的singleflight
 * 同一cacheKey的请求会阻塞直到第一次请求完成，后续请求不再执行，而是使用第一次请求的结果
 */

package blockCache

import (
	lru "github.com/hashicorp/golang-lru"
	"sync"
)

type BlockCache struct {
	blockSize uint32
	blockLock []sync.Mutex
	blockHash HashFunc
	lru       *lru.ARCCache
}

type ActionResult interface {
	Result() interface{}
	Err() error
}

type resultWrapper struct {
	result interface{}
	err    error
}

func (r *resultWrapper) Result() interface{} {
	return r.result
}

func (r *resultWrapper) Err() error {
	return r.err
}

func NewResultWrapper(result interface{}, err error) *resultWrapper {
	return &resultWrapper{
		result: result,
		err:    err,
	}
}

type HashFunc func(b []byte) uint32
type Action func() ActionResult

func (b *BlockCache) WithCacheDo(key string, f Action) ActionResult {
	hashKey := b.blockHash([]byte(key))
	if result := b.tryGetInCache(hashKey); result != nil {
		return result
	}
	if b.blockSize > 0 {
		blockKey := hashKey % b.blockSize
		b.blockLock[blockKey].Lock()
		defer b.blockLock[blockKey].Unlock()
		if result := b.tryGetInCache(hashKey); result != nil {
			return result
		}
	}
	result := f()
	if result != nil {
		b.lru.Add(hashKey, result)
	}
	return result
}

func (b *BlockCache) tryGetInCache(hashKey uint32) ActionResult {
	if result, ok := b.lru.Get(hashKey); ok {
		if result == nil {
			return nil
		}
		return result.(ActionResult)
	}
	return nil
}

// NewBlockCache blockSize 为0表示不使用block，退化为单纯的lru，同一请求有可能执行多次，并覆盖lru缓存
func NewBlockCache(blockSize uint32, lruSize uint32, hashFuncs ...HashFunc) *BlockCache {
	if lruSize == 0 {
		panic("size must greater than 0")
	}
	b := new(BlockCache)
	b.blockHash = fnvHasher
	b.lru, _ = lru.NewARC(int(lruSize))
	if blockSize > 0 {
		b.blockSize = blockSize
		b.blockLock = make([]sync.Mutex, b.blockSize)
	}
	if len(hashFuncs) > 0 && hashFuncs[0] != nil {
		b.blockHash = hashFuncs[0]
	}
	return b
}
