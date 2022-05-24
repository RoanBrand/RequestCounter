build:
	sudo docker compose build

run:
	sudo docker compose up

rebuild-run:
	sudo docker compose up --build

bench:
	go test -bench=. -count=2 -race ./...
