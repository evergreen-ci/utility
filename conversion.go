package utility

// This file includes type conversion functions.

// convertInt32PtrToIntPtr converts *int32 to *int.
func convertInt32PtrToIntPtr(i32 *int32) *int {
	if i32 == nil {
		return nil
	}
	i := int(*i32)
	return &i
}
