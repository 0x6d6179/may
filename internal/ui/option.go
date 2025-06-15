package ui

type Option[T comparable] struct {
	Label       string
	Description string
	Value       T
}
