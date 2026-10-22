package container

type Container struct {
	Infra    *Infra
	Repos    *Repos
	Services *Services
	Handlers *Handlers
}

func New() (*Container, error) {
	in, err := NewInfra()
	if err != nil {
		return nil, err
	}

	repos := buildRepos(in)

	services, err := buildServices(in, repos)
	if err != nil {
		in.Close()
		return nil, err
	}

	in.addCloser(services.SensitiveWord.Close)
	in.addCloser(services.Realtime.Close)
	in.addCloser(services.IMRealtime.Close)

	handlers := buildHandlers(in, services)

	return &Container{
		Infra:    in,
		Repos:    repos,
		Services: services,
		Handlers: handlers,
	}, nil
}

func (c *Container) Close() {
	c.Infra.Close()
}
