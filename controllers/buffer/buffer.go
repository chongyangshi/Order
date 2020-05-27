package buffer

import (
	"fmt"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This buffer holds a list of recently updated resources whose depending pod controllers
// are managed by Order, but some of which cannot yet be restarted as a cooldown since
// the last restarted time is not yet met. So that we can apply precise deduplication for
// bursts of state changes on the same object, we are not using a pre-rolled one like
// https://github.com/kubernetes/client-go/blob/master/util/workqueue/queue.go
var (
	buffer      = map[string]*BufferItem{}
	bufferMutex = sync.Mutex{}
)

// BufferItem uniquely identifies a managed resource based on its cluster
// attributes.
// We record last processed time when pushing to the buffer, so that when
// an item is popped and processed, the calling client can determine whether
// the last processed time fits under any cooldowns it needs to process.
type BufferItem struct {
	TypeMeta               metav1.TypeMeta
	Name                   string
	Namespace              string
	PendingResourceVersion string
	LastProcessed          time.Time
	Attempts               int
}

// HashKey returns a string which identifies a managed resource, we won't fiddle
// with GVK here for the sake of simplificty as not all Custom Resource Definitions
// are guaranteed to implement the get GVK API.
func (b BufferItem) HashKey() string {
	return fmt.Sprintf("%s:%s:%s:%s", b.TypeMeta.APIVersion, b.TypeMeta.Kind, b.Name, b.Namespace)
}

// Push stores a managed resource in the buffer in a uniquely identified form.
func Push(item *BufferItem) {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	if item == nil {
		return
	}

	// Kubernetes does not guarantee that the resource version numeric values are ordered,
	// or whether they will be numeric at all. So we overwrite any existing key for the
	// same pod controller in all cases.
	buffer[item.HashKey()] = item
}

// Pop removes a random pod controller from buffer, if one exists. If not all of its depending
// pod controllers can be processed, it should be popped back onto the buffer by the caller.
func Pop() *BufferItem {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	if len(buffer) == 0 {
		return nil
	}

	// Because this buffer only stores managed resources whose states have recently changed,
	// and is deduplicated, the chronical precedence really doesn't matter. We can implement
	// a doubly linked list here to achiebe O(1) retrieval, but this is completely unnecessary.
	for k, v := range buffer {
		item := *v
		delete(buffer, k)

		// Increment the attempts counter before returning.
		item.Attempts = item.Attempts + 1
		return &item
	}

	// Should never happen
	return nil
}
