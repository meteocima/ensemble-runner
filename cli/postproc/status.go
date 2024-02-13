package main

type FileKind int

const (
	WrfOutFile = FileKind(0)
	AuxFile    = FileKind(1)
)

type PostProcessCompleted struct {
	Domain    int
	ProgrHour int
	Kind      FileKind
}

type PostProcessStatus struct {
	CompletedCh <-chan PostProcessCompleted
}
