package contracts

type AskRepository interface {
	Ask(question string) (string, error)
}
