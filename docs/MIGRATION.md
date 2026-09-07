# Migrating from v1 to v2

Version 2 is a breaking release focused on making the shape of embedded
resources predictable, and on adding capabilities (XML output, generics
helpers) without touching the public JSON shape you already rely on for
non-embedded resources.

## 0. Import path change: `/v2` suffix

Per Go's semantic import versioning rules, the module path gained a `/v2`
suffix. Update your import and `go get`:

```diff
- import hal "github.com/nentgroup/viaplay-hateoas-go"
+ import hal "github.com/nentgroup/viaplay-hateoas-go/v2"
```

```bash
go get github.com/nentgroup/viaplay-hateoas-go/v2
```

The package name imported as `hal` is unchanged; only the module/import path
changes. v1 (`github.com/nentgroup/viaplay-hateoas-go`, no suffix) keeps
working indefinitely for consumers who aren't ready to upgrade yet — Go
module versioning allows both major versions to be depended upon
side-by-side if needed during a gradual migration.

## 1. `Embedded` is now a typed collection map

**Before (v1):**

```go
type Embedded map[Relation]interface{}
```

**Now (v2):**

```go
type Embedded map[Relation]ResourceCollection
```

A new `Get` accessor returns a concrete, typed value instead of `interface{}`:

```go
func (e Embedded) Get(rel Relation) ResourceCollection
```

If you were previously reading `resource.Embedded[rel]` and type-asserting
the result yourself (`.([]*Resource)` or `.(*Resource)`), switch to
`resource.Embedded.Get(rel)`, which always returns a `ResourceCollection`
(nil if the relation is absent).

## 2. JSON shape is driven by cardinality, not by which method you called

**Before (v1):** `Add`/`Embed` always produced a JSON array, even for a
single embedded resource (`"foo": [{...}]`). `Set` produced a JSON object.

**Now (v2):** the number of items in the relation decides the shape when
marshalling to JSON:

- exactly one resource → `"foo": {...}` (object)
- two or more resources → `"foo": [{...}, {...}]` (array)

This matches the pre-v1.0 behavior of the library and is consistent with how
most HAL producers behave: a to-one relation is an object, a to-many relation
is an array.

### What you need to check

If your API consumers currently assume every embedded relation is always an
array (because that was the v1.0+ behavior), you have two options:

1. **Update consumers** to accept either an object or an array for relations
   that can have zero-or-one related resource. This is the recommended,
   HAL-idiomatic fix.
2. **Force array output** for a specific relation by always calling
   `EmbedCollection`/`SetCollection`, even for a single item:

   ```go
   r.EmbedCollection("foo", hal.ResourceCollection{single})
   ```

   A `ResourceCollection` with exactly one item still marshals as an object
   in v2 (cardinality-driven), so if you need to *force* an array
   regardless of length, keep the payload wrapped as JSON yourself instead of
   relying on `Embedded`'s automatic behavior.

## 3. `Embedded.Add`/`Embedded.Set`/`Embedded.SetCollection` signatures are
   unchanged, but nil-safety was added

Calling these methods on a `nil` `Embedded` map is now a no-op instead of a
panic. You still need to initialize `Resource.Embedded` yourself if you are
calling `Embedded` methods directly instead of going through
`Resource.Embed`/`Resource.EmbedCollection` (which lazily initialize it for
you):

```go
r.Embedded = make(hal.Embedded)
r.Embedded.Set("category", child)
```

## 4. New: XML output

`Resource` now implements `xml.Marshaler`, so the exact same resource graph
you build for JSON can also be serialized as HAL+XML:

```go
b, err := xml.MarshalIndent(resource, "", "  ")
```

This did not exist in v1 and requires no changes to existing code — it is
purely additive. See `examples/xml` for a full example.

## 5. New: generics helpers

Three small generic helpers were added to reduce boilerplate around typed
payloads. None of these change existing behavior; they are additive:

- `hal.Payload[T](r *Resource) (T, bool)` — typed payload retrieval.
- `hal.WithMapper[T](p T, fn func(T) Entry) Mapper` — adapt a type to
  `hal.Mapper` without modifying it or its package.
- `hal.NewResourceCollection[T](items []T, selfURI func(T) string) ResourceCollection`
  — build a `ResourceCollection` from a typed slice without a manual loop.

See `examples/generics` for a full example.

## 6. Breaking: `Resource.Flavor` and the new default output shape

`Resource` gained a `Flavor` field controlling the JSON key names and
payload layout used when marshalling:

- `hal.FlavorViaplay` — **the new default** (`hal.DefaultFlavor`). Matches
  `specs/viaplay-hateoas.md`: payload nested under a `data` key, unprefixed
  `links`/`embedded` keys.
- `hal.FlavorHAL` — canonical HAL: payload flattened onto the root object,
  `_links`/`_embedded` keys.
- `hal.FlavorDefault` (zero value) — the old v1 behavior: payload flattened
  onto the root object, unprefixed `links`/`embedded` keys. No longer the
  default `NewResource` assigns, but still available as an explicit opt-in.

**This is a breaking change.** Previously, resources marshalled with their
payload flattened onto the root object. Now, `hal.DefaultFlavor` is
`hal.FlavorViaplay`, so every resource created via `NewResource` (without
touching `Flavor`) will have its payload nested under a `"data"` key
instead. If your application depends on the old flattened shape, restore it
by either:

- Setting `hal.DefaultFlavor = hal.FlavorDefault` once at start-up
  (affects every resource created afterwards), or
- Setting `resource.Flavor = hal.FlavorDefault` (or `hal.FlavorHAL`)
  per-resource.

Note that embedded child resources are marshalled using their own `Flavor`
field, not inherited from their parent — make sure child resources share
the same flavor as their parent (e.g. by setting `hal.DefaultFlavor` once
before building the resource graph, or explicitly setting `Flavor` on each
resource).

### Watch out for manual `"data"` wrapping in existing `Mapper` implementations

If your v1 `Mapper.GetMap()` implementations already wrapped their payload
in a `"data"` key by hand (a common pattern for matching the Viaplay spec
before this library supported it natively), you will now get **double
nesting** once `FlavorViaplay` becomes the default, since the library adds
its own `"data"` wrapper on top:

```go
// v1 workaround — manually nesting payload under "data"
func (a AdInfo) GetMap() hal.Entry {
	return hal.Entry{
		"data": map[string]interface{}{
			"articleId": a.ArticleID,
			"wmap":      a.WMap,
		},
	}
}
```

```json
// v2 output with FlavorViaplay (BUG: double-nested)
{
  "data": {
    "data": {
      "articleId": "123",
      "wmap": "..."
    }
  },
  "links": { "self": { "href": "/x" } }
}
```

Fix this by removing the manual `"data"` wrapper and returning the payload
fields directly — the library now does the nesting for you:

```go
// v2 — let the library nest under "data"
func (a AdInfo) GetMap() hal.Entry {
	return hal.Entry{
		"articleId": a.ArticleID,
		"wmap":      a.WMap,
	}
}
```

```json
{
  "data": {
    "articleId": "123",
    "wmap": "..."
  },
  "links": { "self": { "href": "/x" } }
}
```

If you'd rather keep full manual control over the payload shape (e.g. during
an incremental migration), set `resource.Flavor = hal.FlavorDefault` (or
`hal.FlavorHAL`) so the library flattens instead of nesting, and your
existing manual `"data"` key passes through untouched as a regular payload
field.

## 7. Minimum Go version

The module now requires Go 1.27 (generics-related standard library and
toolchain improvements). Update your `go.mod` accordingly.

## 8. Performance: `Mapper.GetMap` must return a fresh map

For performance, `Resource.GetMap()` now adds `links`/`embedded` (or nests
payload under `"data"` for `FlavorViaplay`) **directly into the `Entry`
returned by your `Mapper.GetMap()`** instead of copying it into a new map.
This is a behavioral requirement, not just an implementation detail: your
`GetMap()` must return a freshly allocated `hal.Entry` on every call (e.g. a
map literal built inline), not a shared or cached map instance reused across
calls. This is already how virtually all `Mapper` implementations are
written (see the Basic Example above), so most consumers are unaffected —
but if you cache/memoize the map returned by `GetMap()`, switch to
allocating a new one per call.

## Summary checklist

- [ ] Update your import path to `github.com/nentgroup/viaplay-hateoas-go/v2`.
- [ ] Replace direct type assertions on `Embedded[rel]` with `Embedded.Get(rel)`.
- [ ] Audit API consumers that assume embedded relations are always arrays.
- [ ] Bump your Go toolchain to 1.27+.
- [ ] Ensure `Mapper.GetMap()` implementations return a fresh map per call,
      not a cached/shared instance.
- [ ] (Optional) Adopt `hal.WithMapper`/`hal.NewResourceCollection` to simplify
      typed payload code.
- [ ] (Optional) Expose an XML representation of your API using the new
      `xml.Marshaler` support, if useful.
