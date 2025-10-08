package tray

type GenericSlice[T any] struct {
	data []T
}

func NewGenericSlice[T any]() *GenericSlice[T] {
	return &GenericSlice[T]{
		data: make([]T, 0),
	}
}

func NewGenericSliceWithCapacity[T any](capacity int) *GenericSlice[T] {
	return &GenericSlice[T]{
		data: make([]T, 0, capacity),
	}
}

func (gs *GenericSlice[T]) Get(index int) (T, bool) {
	var zero T
	if index < 0 || index >= len(gs.data) {
		return zero, false
	}
	return gs.data[index], true
}

func (gs *GenericSlice[T]) Set(index int, value T) bool {
	if index < 0 || index >= len(gs.data) {
		return false
	}
	gs.data[index] = value
	return true
}

func (gs *GenericSlice[T]) Add(value T) {
	gs.data = append(gs.data, value)
}
func (gs *GenericSlice[T]) Insert(index int, value T) bool {
	if index < 0 || index > len(gs.data) {
		return false
	}

	gs.data = append(gs.data, value)
	copy(gs.data[index+1:], gs.data[index:])
	gs.data[index] = value
	return true
}

func (gs *GenericSlice[T]) Remove(index int) bool {
	if index < 0 || index >= len(gs.data) {
		return false
	}

	gs.data = append(gs.data[:index], gs.data[index+1:]...)
	return true
}

func (gs *GenericSlice[T]) Length() int {
	return len(gs.data)
}
func (gs *GenericSlice[T]) Capacity() int {
	return cap(gs.data)
}

func (gs *GenericSlice[T]) IsEmpty() bool {
	return len(gs.data) == 0
}
func (gs *GenericSlice[T]) Clear() {
	gs.data = make([]T, 0)
}

func (gs *GenericSlice[T]) Values() []T {
	result := make([]T, len(gs.data))
	copy(result, gs.data)
	return result
}
