package docker

type Option func(*Network)

func WithKeepStoppedContainers(keepStoppedContainers bool) Option {
	return func(network *Network) {
		network.keepStoppedContainers = keepStoppedContainers
	}
}
