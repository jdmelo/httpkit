package filters

import (
	"log"
	"net/http"
	"sync"
)

type FilterFunc func(next http.Handler) http.Handler

// 先进后出
type FilterManager struct {
	mu      sync.RWMutex
	filters []FilterFunc
}

func NewFilterManager() *FilterManager {
	return &FilterManager{
		filters: make([]FilterFunc, 0),
	}
}

func (f *FilterManager) AddFilter(h FilterFunc) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.filters = append(f.filters, h)
}

func (f *FilterManager) Wrap(n http.Handler) http.Handler {
	f.mu.RLock()
	defer f.mu.RUnlock()

	tmp := n

	fLen := len(f.filters)

	for i := fLen - 1; i >= 0; i-- {
		log.Printf("wrap handler %v.\n", i)
		fh := f.filters[i]
		tmp = fh(tmp)
	}

	return tmp
}
