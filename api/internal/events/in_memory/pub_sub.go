package inmemory

import (
	"sync"

	"tomerab.com/cam-hub/internal/utils"
)

type InMemoryPubSub struct {
	/// A channel is stored for each subscriber so the events can be
	// distributed evenly
	store map[string][]chan []byte

	mtx sync.RWMutex
}

func NewInMemoryPubSub() *InMemoryPubSub {
	return &InMemoryPubSub{
		store: make(map[string][]chan []byte),
	}
}

func (mem *InMemoryPubSub) Subscribe(uuid string) chan []byte {
	ch := make(chan []byte, 64)
	mem.mtx.Lock()
	defer mem.mtx.Unlock()

	mem.store[uuid] = append(mem.store[uuid], ch)
	return ch
}

func (mem *InMemoryPubSub) Purge(uuid string) {
	mem.mtx.Lock()
	defer mem.mtx.Unlock()
	delete(mem.store, uuid)
}

func (mem *InMemoryPubSub) Unsubscribe(uuid string, ch chan []byte) {
	mem.mtx.Lock()
	defer mem.mtx.Unlock()
	defer close(ch)

	arr := mem.store[uuid]
	for i := range arr {
		if arr[i] == ch {
			utils.Swap(arr, i, len(arr)-1)
			arr = arr[:len(arr)-1]
			break
		}
	}

	if len(arr) == 0 {
		delete(mem.store, uuid)
	} else {
		mem.store[uuid] = arr
	}
}

func (mem *InMemoryPubSub) Broadcast(uuid string, msg []byte) {
	mem.mtx.RLock()
	defer mem.mtx.RUnlock()

	for _, sub := range mem.store[uuid] {
		select {
		case sub <- msg:
		default:
		}
	}
}
