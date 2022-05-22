build:
	go build -o cluster cmd/Cluster/main.go
	go build -o requestcounter cmd/RequestCounter/main.go

build-docker:
	sudo docker build -f cmd/RequestCounter/Dockerfile --tag requestcounter .

run-cluster:
	./cluster

run-requestcounter:
	./requestcounter
