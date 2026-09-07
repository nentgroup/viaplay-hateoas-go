package hal

// generics.go provides typed helpers around the untyped Payload/Mapper
// primitives so that callers working with a known payload type do not need
// to write manual type assertions or Mapper implementations for simple
// cases.

// Payload returns the resource's payload cast to T. The second return value
// reports whether the assertion succeeded; if it did not, the zero value of
// T is returned.
func Payload[T any](r *Resource) (T, bool) {
	v, ok := r.Payload.(T)
	return v, ok
}

// mapperFunc adapts a plain mapping function into a Mapper implementation
// for a concrete payload type, without requiring T itself to implement
// Mapper. This is useful for payload types defined in packages you don't
// own (so you cannot add a GetMap method to them).
type mapperFunc[T any] struct {
	payload T
	fn      func(T) Entry
}

// GetMap implements the Mapper interface by delegating to fn.
func (m mapperFunc[T]) GetMap() Entry {
	return m.fn(m.payload)
}

// WithMapper wraps p so that it satisfies the Mapper interface using fn as
// its GetMap implementation. The result can be passed directly to
// NewResource:
//
//	type externalDTO struct{ Name string }
//
//	mapped := hal.WithMapper(dto, func(d externalDTO) hal.Entry {
//		return hal.Entry{"name": d.Name}
//	})
//	r := hal.NewResource(mapped, "/items/1")
//
//nolint:ireturn // adapter pattern: the whole point is to return the Mapper interface so callers can pass it directly to NewResource
func WithMapper[T any](p T, fn func(T) Entry) Mapper {
	return mapperFunc[T]{payload: p, fn: fn}
}

// NewResourceCollection builds a ResourceCollection from a slice of typed
// items, using selfURI to compute each item's self link href. This removes
// the need for a manual loop when embedding a homogeneous collection of
// resources.
//
//	tasks := hal.NewResourceCollection(items, func(t Task) string {
//		return fmt.Sprintf("/tasks/%d", t.Id)
//	})
//	r.EmbedCollection("tasks", tasks)
func NewResourceCollection[T any](items []T, selfURI func(T) string) ResourceCollection {
	rc := make(ResourceCollection, 0, len(items))
	for _, item := range items {
		rc = append(rc, NewResource(item, selfURI(item)))
	}
	return rc
}
