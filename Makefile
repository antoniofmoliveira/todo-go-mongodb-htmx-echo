.PHONY: start stop

start:
	@echo "Starting process"
	@air & echo $$! > p2.pid &

stop:
	@echo "Stopping process"
	@kill `cat p2.pid` || true
	@rm -f p2.pid