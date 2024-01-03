package dream

func Map[T any, O any](items []T, f func(T)O) []O {
	output := []O{}

	for i := 0; i < len(items); i++ {
		output = append(output, f(items[i]))
	}
	return output
}

func MapIndexed[T any, O any](items []T, f func(int, T)O) []O {
	output := []O{}

	for i := 0; i < len(items); i++ {
		output = append(output, f(i, items[i]))
	}
	return output
}

func Filter[T any](items []T, f func(T)bool) []T {
	output := []T{}

	for i := 0; i < len(items); i++ {
		if f(items[i]){
			output = append(output, items[i])
		}
	}
	return output
}

func FilterIndexed[T any](items []T, f func(int, T)bool) []T {
	output := []T{}

	for i := 0; i < len(items); i++ {
		if f(i, items[i]){
			output = append(output, items[i])
		}
	}
	return output
}

func Reduce[T any, O any](items []T, f func(T, *O)bool, init O) O {
	for i := 0; i < len(items); i++ {
		f(items[i], &init)
	}
	return init
}

func Each[T any](items []T, f func(T)) {
	for i := 0; i < len(items); i++ {
		f(items[i])
	}
}

func ReduceIndexed[T any, O any](items []T, f func(int, T, *O)bool, init O) O {
	for i := 0; i < len(items); i++ {
		f(i, items[i], &init)
	}
	return init
}

func EachIndexed[T any](items []T, f func(int, T)) {
	for i := 0; i < len(items); i++ {
		f(i, items[i])
	}
}