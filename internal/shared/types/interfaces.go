package types

import "context"

// Validator defines the validation contract
type Validator interface {
	ValidateStruct(ctx context.Context, data any) []ValidationError
}

// OperationExecutor defines the operation execution contract
type OperationExecutor[T any] interface {
	Execute(ctx context.Context, fn func(context.Context) (T, error)) (T, error)
}