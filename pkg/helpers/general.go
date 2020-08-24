package helpers

type Tuple struct {
	Key   string `json:"k"`
	Value string `json:"v"`
}

type TupleInt struct {
	Key   int `json:"k"`
	Value int `json:"v"`
}

// IgnoreErrors returns nil if an error is of one of the provided types, returns the provided error otherwise.
func IgnoreErrors(err error, errs ...error) error {

	for _, v := range errs {
		if err == v {
			return nil
		}
	}

	return err
}
