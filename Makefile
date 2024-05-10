.PHONY: run run-background stop

run:
	@echo "Running Excaliroom..."
	@docker-compose up --build --wait

stop:
	@echo "Stopping Excaliroom..."
	@docker-compose down --remove-orphans --volumes