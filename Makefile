.PHONY: dev
dev:
	air

.PHONY: migrate
migrate:
	@bash -c 'set -a; source .env; set +a; psql "$$DATABASE_URL" -f migrations/001_init.sql'
