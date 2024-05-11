package cmd

const (
	RestApp = "REST"
)

type App interface {
	Start()
	Stop()
}
