package commands

type Commands interface {
	Name() string
	Execute(args []string) error
}
