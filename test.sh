#!/usr/bin/env bash
set -euo pipefail

RESTORE='\033[0m'
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BOLD='\033[1m'

pass=0
fail=0

print_result() {
	if [ "$1" = "PASS" ]; then
		echo -e "  ${GREEN}✓ PASS${RESTORE} $2"
		pass=$((pass + 1))
	else
		echo -e "  ${RED}✗ FAIL${RESTORE} $2"
		fail=$((fail + 1))
	fi
}

check_deps() {
	local missing=0
	for cmd in go docker; do
		if ! command -v "$cmd" &>/dev/null; then
			echo -e "${RED}missing: $cmd${RESTORE}"
			missing=1
		fi
	done
	return $missing
}

echo -e "${BOLD}═══════════════════════════════════════${RESTORE}"
echo -e "${BOLD}  Carevo — Test Suite${RESTORE}"
echo -e "${BOLD}═══════════════════════════════════════${RESTORE}"
echo ""

# ─── Prerequisites ──────────────────────────────────
echo -e "${YELLOW}▶ prerequisites${RESTORE}"
if ! check_deps; then
	echo -e "${RED}install missing deps and try again${RESTORE}"
	exit 1
fi
print_result PASS "go + docker available"
echo ""

# ─── Go Vet ─────────────────────────────────────────
echo -e "${YELLOW}▶ go vet${RESTORE}"
if go vet ./... 2>&1; then
	print_result PASS "go vet"
else
	print_result FAIL "go vet"
fi
echo ""

# ─── Go Build ───────────────────────────────────────
echo -e "${YELLOW}▶ go build${RESTORE}"
if go build ./... 2>&1; then
	print_result PASS "go build"
else
	print_result FAIL "go build"
fi
echo ""

# ─── Unit + Integration Tests (Testcontainers) ─────
echo -e "${YELLOW}▶ go test (Testcontainers — spins up real PostgreSQL)${RESTORE}"
echo -e "  ${YELLOW}  needs Docker, ~30s first run (pulls postgres:17-alpine)${RESTORE}"
if go test ./... -v -count=1 -timeout=300s 2>&1 | tail -20; then
	print_result PASS "go test"
else
	print_result FAIL "go test"
fi
echo ""

# ─── Docker Build ───────────────────────────────────
echo -e "${YELLOW}▶ docker build (API + seed targets)${RESTORE}"
if docker build --target api -t carevo-api:test . &>/dev/null && \
   docker build --target seed -t carevo-seed:test . &>/dev/null; then
	print_result PASS "docker build (api + seed)"
else
	print_result FAIL "docker build"
fi
echo ""

# ─── Summary ────────────────────────────────────────
echo -e "${BOLD}═══════════════════════════════════════${RESTORE}"
echo -e "${BOLD}  Results: ${pass} passed, ${fail} failed${RESTORE}"
echo -e "${BOLD}═══════════════════════════════════════${RESTORE}"

# ─── Optional: live server curl tests ───────────────
if nc -z localhost 8080 2>/dev/null; then
	echo ""
	echo -e "${YELLOW}▶ live server detected — running curl smoke tests${RESTORE}"

	check_endpoint() {
		local url="$1"
		local expect="$2"
		local label="$3"
		local status
		status=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
		if [ "$status" = "$expect" ]; then
			print_result PASS "$label ($status)"
		else
			print_result FAIL "$label (expected $expect, got $status)"
		fi
	}

	check_endpoint "http://localhost:8080/ping" "200" "GET /ping"
	check_endpoint "http://localhost:8080/health" "200" "GET /health"
	check_endpoint "http://localhost:8080/careers" "200" "GET /careers"
	check_endpoint "http://localhost:8080/careers/search?q=software" "200" "GET /careers/search?q=software"
	check_endpoint "http://localhost:8080/categories" "200" "GET /categories"
	check_endpoint "http://localhost:8080/careers/1/resources" "200" "GET /careers/1/resources"
	check_endpoint "http://localhost:8080/careers/1/roadmap" "200" "GET /careers/1/roadmap"
	check_endpoint "http://localhost:8080/careers/1/projects" "200" "GET /careers/1/projects"
	check_endpoint "http://localhost:8080/careers/nonexistent" "404" "GET /careers/nonexistent"
	check_endpoint "http://localhost:8080/swagger/doc.json" "200" "GET /swagger/doc.json"

	echo ""
	echo -e "${BOLD}═══════════════════════════════════════${RESTORE}"
	echo -e "${BOLD}  Live: ${pass} passed, ${fail} failed${RESTORE}"
	echo -e "${BOLD}═══════════════════════════════════════${RESTORE}"
else
	echo ""
	echo -e "  ${YELLOW}(no server on :8080 — skipping curl smoke tests)${RESTORE}"
	echo -e "  ${YELLOW}  start with: docker compose up -d && ./test.sh${RESTORE}"
fi

exit $fail
