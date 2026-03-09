.PHONY: css e2e e2e-up e2e-down e2e-coverage

css:
	npx @tailwindcss/cli -i static/css/input.css -o static/css/styles.css --minify

e2e: e2e-up
	cd e2e && npx playwright test; \
	status=$$?; \
	cd .. && $(MAKE) e2e-coverage; \
	$(MAKE) e2e-down; \
	exit $$status

e2e-up:
	mkdir -p coverage
	docker compose -f docker-compose.test.yml up -d --build
	@echo "Waiting for services to become healthy..."
	@while [ "$$(docker compose -f docker-compose.test.yml ps 2>/dev/null | grep -c '(healthy)')" -lt 3 ]; do sleep 2; done
	@while ! docker compose -f docker-compose.test.yml ps 2>/dev/null | grep oauth2-proxy | grep -q "Up"; do sleep 2; done
	docker compose -f docker-compose.test.yml exec -T postgres \
		psql -U deckel -d deckel_test -f /seed.sql

e2e-down:
	docker compose -f docker-compose.test.yml down -v
	rm -rf coverage

e2e-coverage:
	docker compose -f docker-compose.test.yml stop app
	go tool covdata textfmt -i=./coverage -o=coverage.out
	go tool cover -func=coverage.out
