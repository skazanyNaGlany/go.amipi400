package interfaces

type KeyEventCallback func(
	sender any,
	key string,
	pressed bool,
	prevPressedKeys map[string]int,
	newPressedKeys map[string]int)
