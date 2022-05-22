build:
	go build -o cluster cmd/Cluster/main.go
	go build -o requestcounter cmd/RequestCounter/main.go

run-cluster:
	./cluster

run-requestcounter:
	./requestcounter
