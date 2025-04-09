package controlloop

type opts struct {
	logger Logger
}

type ClOption func(*opts)

func WithLogger(logger Logger) ClOption {
	return func(o *opts) {
		o.logger = logger
	}
}
