package httpshell

type StartupObserver interface {
	BeforeStartup() error
}

type ShutdownObserver interface {
	AfterShutdown()
}
