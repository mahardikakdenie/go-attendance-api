include .env
export

.PHONY: help build push pull up down restart logs ps clean

help: ## Show this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the docker image
	docker compose build

push: ## Push the docker image to registry
	docker compose push

pull: ## Pull the latest images from registry
	docker compose pull

up: ## Start the services in background
	docker compose up -d

down: ## Stop and remove containers
	docker compose down

restart: ## Restart the services
	docker compose restart

logs: ## View services logs
	docker compose logs -f

ps: ## List running containers
	docker compose ps

clean: ## Remove unused docker images and volumes
	docker system prune -f
	docker image prune -f
