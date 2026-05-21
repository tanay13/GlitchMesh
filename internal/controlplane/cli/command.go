package cli

type Command interface {
	Name() string
	Execute(args []string) error
}
