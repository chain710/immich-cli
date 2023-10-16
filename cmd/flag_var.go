package cmd

type genericVar struct {
	s string
	t string
}

func (t *genericVar) String() string {
	return t.s
}

func (t *genericVar) Set(s string) error {
	t.s = s
	return nil
}

func (t *genericVar) Type() string {
	return t.t
}
