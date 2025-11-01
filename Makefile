.PHONY: dev
dev:
	air

.PHONY: migrate
migrate:
	@bash -c 'set -a; source .env; set +a; \
	for f in migrations/*.sql; do \
	  echo "Running $$f"; \
	  psql "$$DATABASE_URL" -f $$f || exit 1; \
	done'
