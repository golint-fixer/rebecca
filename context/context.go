package context

// Context is for representing querying context.
// It is required for implementation of orderby, groupby, limit and skip.
type Context interface {
	GetOrder() string
	GetGroup() string
	GetLimit() int
	GetSkip() int

	SetOrder(string) Context
	SetGroup(string) Context
	SetLimit(int) Context
	SetSkip(int) Context
}