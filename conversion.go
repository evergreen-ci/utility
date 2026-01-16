package utility

// This file includes type conversion functions.

// ConvertInt32PtrToIntPtr converts *int32 to *int.
func ConvertInt32PtrToIntPtr(i32 *int32) *int {
	if i32 == nil {
		return nil
	}
	i := int(*i32)
	return &i
}
