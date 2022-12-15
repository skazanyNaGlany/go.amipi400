package interfaces

type DetachedBlockDevice func(name string) (bool, error)
