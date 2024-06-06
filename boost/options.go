package boost

// withDefaults prepends boost withDefaults to the sequence of options
func withDefaults(options ...Option) []Option {
	const (
		noDefaults = 1
	)
	o := make([]Option, 0, len(options)+noDefaults)
	o = append(o, WithGenerator(&Sequential{
		Format: "ID:%08d",
	}))
	o = append(o, options...)

	return o
}
