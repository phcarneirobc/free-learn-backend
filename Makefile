.PHONY: docker

all: 
	go run .

docker:
	docker compose up -d

clean:
	docker compose down