package buffer

import (
	"fmt"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This buffer holds a list of recently updated resources whose depending pod controllers
// are managed by orderrrr, but some of which cannot yet be restarted as a cooldown since
// the last restarted time is not yet met.
var (
	buffer      = map[string]*BufferItem{}
	bufferMutex = sync.Mutex{}
)

// BufferItem uniquely identifies a managed resource based on its cluster attributes.
type BufferItem struct {
	TypeMeta               metav1.TypeMeta
	Name                   string
	Namespace              string
	PendingResourceVersion string
}

// HashKey returns a string which identifies a managed resource, we won't fiddle
// with GVK here for the sake of simplificty as not all Custom Resource Definitions
// are guaranteed to implement the get GVK API.
func (b BufferItem) HashKey() string {
	return fmt.Sprintf("%s:%s:%s:%s", b.TypeMeta.APIVersion, b.TypeMeta.Kind, b.Name, b.Namespace)
}

// PushToBuffer stores a managed resource in the buffer in a uniquely identified form.
func PushToBuffer(tm metav1.TypeMeta, name, namespace, resourceVersion string) {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	item := BufferItem{
		TypeMeta:               tm,
		Name:                   name,
		Namespace:              namespace,
		PendingResourceVersion: resourceVersion,
	}

	// Kubernetes does not guarantee that the resource version numeric values are ordered,
	// or whether they will be numeric at all. So we overwrite any existing key for the
	// same pod controller in all cases.
	buffer[item.HashKey()] = &item
}

// PopFromBuffer removes a pod controller from buffer, if it exists. If not all of its depending
// pod controllers can be processed, it should be popped back onto the buffer by the caller.
func PopFromBuffer(tm metav1.TypeMeta, name, namespace, resourceVersion string) *BufferItem {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	item := BufferItem{
		TypeMeta:  tm,
		Name:      name,
		Namespace: namespace,
	}

	// The resource version returned in any found item in the buffer will have been the
	// latest inserted.
	if item, found := buffer[item.HashKey()]; found {
		delete(buffer, item.HashKey())
		return item
	}

	return nil
}
