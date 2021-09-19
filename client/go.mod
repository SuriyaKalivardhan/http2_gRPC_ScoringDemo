module azuremachinelearning.com/client

go 1.16

replace azuremachinelearning.com/scorer => ../contract

require (
	azuremachinelearning.com/scorer v0.0.0-00010101000000-000000000000 // indirect
	google.golang.org/grpc v1.40.0 // indirect
)
