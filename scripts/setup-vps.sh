#!/usr/bin/env bash
# Bootstrap a fresh Ubuntu VPS: install git & Docker, deploy go-blog, start via Compose.
#
# Prerequisites: shared PostgreSQL on the infra Docker network
# (see docs/example-postgresql-docker-compose/scripts/setup-vps.sh).
#
# Usage (on VPS as root):
#   curl -fsSL https://raw.githubusercontent.com/sigmaray/go-blog/main/scripts/setup-vps.sh | bash
#   # or
#   sudo bash scripts/setup-vps.sh
#   sudo bash scripts/setup-vps.sh --swap
#
# Environment variables:
#   DEPLOY_DIR                  Target directory (default: ~/r/d/go-blog)
#   REPO_URL                    Git clone URL (default: https://github.com/sigmaray/go-blog.git)
#   GIT_REF                     Branch, tag, or commit to deploy (default: main)
#   REPO_SUBPATH                Path inside the repo to sync (default: repo root)
#   REPO_CACHE_DIR              Full-repo clone cache (default: DEPLOY_DIR/.repo-cache)
#   POSTGRES_ENV_FILE           Shared Postgres .env to read password from
#                               (default: ~/r/d/postgresql/.env)
#   GO_BLOG_HTTP_PORT           Host HTTP port (default: 8083)
#   GIN_MODE                    Gin mode (default: release)
#   GO_BLOG_SESSION_SECRET      Session signing key (default: auto-generated, saved in .env)
#   GO_BLOG_SESSION_SECURE      Set to 1 behind HTTPS (default: 0)
#   GO_BLOG_DATABASE_HOST       PostgreSQL host (default: postgresql)
#   GO_BLOG_DATABASE_PORT       PostgreSQL port (default: 5432)
#   GO_BLOG_DATABASE_NAME       Database name (default: goblog)
#   GO_BLOG_DATABASE_USER       Database user (default: postgres)
#   GO_BLOG_DATABASE_PASSWORD   Database password (default: from POSTGRES_ENV_FILE or existing .env)
#   SETUP_SKIP_APT              Set to 1 to skip apt-get (useful in CI where git is preinstalled)
#   SETUP_SKIP_DOCKER_INSTALL   Set to 1 to skip Docker installation (useful in CI)
#   SETUP_SKIP_MIGRATE          Set to 1 to skip ./blog migrate after start
#   SETUP_SOURCE_DIR            Copy project tree from this path instead of cloning (CI / local test)
#   SETUP_ALLOW_NON_ROOT        Set to 1 to skip root check (CI with passwordless sudo)
#   SETUP_FORCE                 Set to 1 to redeploy even when already at GIT_REF and running
#   SETUP_SWAP                  Set to 1 to configure swap (same as --swap)
#   SETUP_SWAP_SIZE_MB          Swap file size in megabytes (default: 2048)
#   SETUP_SWAP_FILE             Swap file path (default: /swapfile)
#   APP_READY_TIMEOUT_SEC       Seconds to wait for /health (default: 180)

set -euo pipefail

DEPLOY_DIR="${DEPLOY_DIR:-${HOME}/r/d/go-blog}"
REPO_URL="${REPO_URL:-https://github.com/sigmaray/go-blog.git}"
GIT_REF="${GIT_REF:-main}"
REPO_SUBPATH="${REPO_SUBPATH:-}"
REPO_CACHE_DIR="${REPO_CACHE_DIR:-${DEPLOY_DIR}/.repo-cache}"
POSTGRES_ENV_FILE="${POSTGRES_ENV_FILE:-${HOME}/r/d/postgresql/.env}"
GO_BLOG_HTTP_PORT="${GO_BLOG_HTTP_PORT:-8083}"
GIN_MODE="${GIN_MODE:-release}"
GO_BLOG_SESSION_SECURE="${GO_BLOG_SESSION_SECURE:-0}"
GO_BLOG_DATABASE_HOST="${GO_BLOG_DATABASE_HOST:-postgresql}"
GO_BLOG_DATABASE_PORT="${GO_BLOG_DATABASE_PORT:-5432}"
GO_BLOG_DATABASE_NAME="${GO_BLOG_DATABASE_NAME:-goblog}"
GO_BLOG_DATABASE_USER="${GO_BLOG_DATABASE_USER:-postgres}"
SETUP_SKIP_APT="${SETUP_SKIP_APT:-0}"
SETUP_SKIP_DOCKER_INSTALL="${SETUP_SKIP_DOCKER_INSTALL:-0}"
SETUP_SKIP_MIGRATE="${SETUP_SKIP_MIGRATE:-0}"
SETUP_SOURCE_DIR="${SETUP_SOURCE_DIR:-}"
SETUP_ALLOW_NON_ROOT="${SETUP_ALLOW_NON_ROOT:-0}"
SETUP_FORCE="${SETUP_FORCE:-0}"
SETUP_SWAP="${SETUP_SWAP:-0}"
SETUP_SWAP_SIZE_MB="${SETUP_SWAP_SIZE_MB:-2048}"
SETUP_SWAP_FILE="${SETUP_SWAP_FILE:-/swapfile}"
APP_READY_TIMEOUT_SEC="${APP_READY_TIMEOUT_SEC:-180}"

DEPLOY_ENV_FILE="${DEPLOY_DIR}/.env"

log() {
  printf '[setup-go-blog] %s\n' "$*" >&2
}

die() {
  printf '[setup-go-blog] ERROR: %s\n' "$*" >&2
  exit 1
}

usage() {
  cat <<'EOF'
Usage: setup-vps.sh [--swap]

Bootstrap Ubuntu, install git and Docker, deploy go-blog, and start docker compose.

Prerequisites: shared PostgreSQL on the infra network
(docs/example-postgresql-docker-compose/scripts/setup-vps.sh).

Options:
  --swap                      Create and enable a swap file if swap is not configured

Environment variables:
  DEPLOY_DIR                  Deployment directory (default: ~/r/d/go-blog)
  REPO_URL                    Git repository URL
  GIT_REF                     Branch, tag, or commit (default: main)
  REPO_SUBPATH                Subdirectory inside the repo to deploy (default: repo root)
  REPO_CACHE_DIR              Local cache for the full git clone
  POSTGRES_ENV_FILE           Shared Postgres .env for database password
  GO_BLOG_HTTP_PORT           Host HTTP port (default: 8083)
  GIN_MODE                    Gin mode (default: release)
  GO_BLOG_SESSION_SECRET      Session secret (saved to DEPLOY_DIR/.env when unset)
  GO_BLOG_SESSION_SECURE      Set to 1 behind HTTPS (default: 0)
  GO_BLOG_DATABASE_HOST       PostgreSQL host (default: postgresql)
  GO_BLOG_DATABASE_PORT       PostgreSQL port (default: 5432)
  GO_BLOG_DATABASE_NAME       Database name (default: goblog)
  GO_BLOG_DATABASE_USER       Database user (default: postgres)
  GO_BLOG_DATABASE_PASSWORD   Database password (from POSTGRES_ENV_FILE when unset)
  SETUP_SKIP_APT              Skip apt-get when set to 1
  SETUP_SKIP_DOCKER_INSTALL   Skip Docker install when set to 1
  SETUP_SKIP_MIGRATE          Skip migrations when set to 1
  SETUP_SOURCE_DIR            Use existing directory instead of git clone
  SETUP_ALLOW_NON_ROOT        Allow running without root (for CI)
  SETUP_FORCE                 Redeploy even when already at GIT_REF and running
  SETUP_SWAP                  Configure swap when set to 1 (same as --swap)
  SETUP_SWAP_SIZE_MB          Swap file size in megabytes (default: 2048)
  SETUP_SWAP_FILE             Swap file path (default: /swapfile)
  APP_READY_TIMEOUT_SEC       /health readiness timeout in seconds
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      -h|--help)
        usage
        exit 0
        ;;
      --swap)
        SETUP_SWAP=1
        shift
        ;;
      *)
        die "Unknown option: $1 (try --help)"
        ;;
    esac
  done
}

require_root() {
  if [[ "${SETUP_ALLOW_NON_ROOT}" == "1" ]]; then
    return 0
  fi
  if [[ "${EUID:-$(id -u)}" -ne 0 ]]; then
    die "Run as root: sudo bash $0"
  fi
}

read_env_value() {
  local key="$1"
  local file="$2"
  [[ -f "${file}" ]] || return 1
  grep -E "^${key}=" "${file}" | tail -1 | cut -d= -f2-
}

generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -base64 48 | tr -d '\n'
    return 0
  fi
  die "openssl is not available to generate secrets"
}

prefer_existing_env() {
  local key="$1"
  local current="$2"
  local default="$3"
  local existing=""

  existing="$(read_env_value "${key}" "${DEPLOY_ENV_FILE}" 2>/dev/null || true)"
  if [[ -n "${existing}" ]] && [[ "${current}" == "${default}" ]]; then
    printf '%s' "${existing}"
    return 0
  fi
  printf '%s' "${current}"
}

ensure_deploy_env() {
  [[ -d "${DEPLOY_DIR}" ]] || die "Deploy directory missing: ${DEPLOY_DIR}"

  local existing_secret=""
  existing_secret="$(read_env_value GO_BLOG_SESSION_SECRET "${DEPLOY_ENV_FILE}" 2>/dev/null || true)"
  if [[ -z "${GO_BLOG_SESSION_SECRET:-}" ]]; then
    if [[ -n "${existing_secret}" ]]; then
      GO_BLOG_SESSION_SECRET="${existing_secret}"
      log "Using GO_BLOG_SESSION_SECRET from ${DEPLOY_ENV_FILE}"
    else
      GO_BLOG_SESSION_SECRET="$(generate_secret)"
      log "Generated new GO_BLOG_SESSION_SECRET"
    fi
  fi
  if [[ "${#GO_BLOG_SESSION_SECRET}" -lt 32 ]]; then
    die "GO_BLOG_SESSION_SECRET must be at least 32 characters"
  fi
  export GO_BLOG_SESSION_SECRET

  local existing_db_password=""
  existing_db_password="$(read_env_value GO_BLOG_DATABASE_PASSWORD "${DEPLOY_ENV_FILE}" 2>/dev/null || true)"
  if [[ -z "${GO_BLOG_DATABASE_PASSWORD:-}" ]]; then
    if [[ -n "${existing_db_password}" ]]; then
      GO_BLOG_DATABASE_PASSWORD="${existing_db_password}"
      log "Using GO_BLOG_DATABASE_PASSWORD from ${DEPLOY_ENV_FILE}"
    else
      local postgres_password=""
      postgres_password="$(read_env_value POSTGRES_PASSWORD "${POSTGRES_ENV_FILE}" 2>/dev/null || true)"
      if [[ -n "${postgres_password}" ]]; then
        GO_BLOG_DATABASE_PASSWORD="${postgres_password}"
        log "Using GO_BLOG_DATABASE_PASSWORD from ${POSTGRES_ENV_FILE}"
      else
        die "GO_BLOG_DATABASE_PASSWORD is unset; set it or deploy Postgres so ${POSTGRES_ENV_FILE} exists"
      fi
    fi
  fi
  export GO_BLOG_DATABASE_PASSWORD

  GO_BLOG_HTTP_PORT="$(prefer_existing_env GO_BLOG_HTTP_PORT "${GO_BLOG_HTTP_PORT}" "8083")"
  GIN_MODE="$(prefer_existing_env GIN_MODE "${GIN_MODE}" "release")"
  GO_BLOG_SESSION_SECURE="$(prefer_existing_env GO_BLOG_SESSION_SECURE "${GO_BLOG_SESSION_SECURE}" "0")"
  GO_BLOG_DATABASE_HOST="$(prefer_existing_env GO_BLOG_DATABASE_HOST "${GO_BLOG_DATABASE_HOST}" "postgresql")"
  GO_BLOG_DATABASE_PORT="$(prefer_existing_env GO_BLOG_DATABASE_PORT "${GO_BLOG_DATABASE_PORT}" "5432")"
  GO_BLOG_DATABASE_NAME="$(prefer_existing_env GO_BLOG_DATABASE_NAME "${GO_BLOG_DATABASE_NAME}" "goblog")"
  GO_BLOG_DATABASE_USER="$(prefer_existing_env GO_BLOG_DATABASE_USER "${GO_BLOG_DATABASE_USER}" "postgres")"

  local tmp
  tmp="$(mktemp)"
  chmod 600 "${tmp}"
  {
    printf 'GO_BLOG_HTTP_PORT=%s\n' "${GO_BLOG_HTTP_PORT}"
    printf 'GIN_MODE=%s\n' "${GIN_MODE}"
    printf 'GO_BLOG_SESSION_SECRET=%s\n' "${GO_BLOG_SESSION_SECRET}"
    printf 'GO_BLOG_SESSION_SECURE=%s\n' "${GO_BLOG_SESSION_SECURE}"
    printf 'GO_BLOG_DATABASE_HOST=%s\n' "${GO_BLOG_DATABASE_HOST}"
    printf 'GO_BLOG_DATABASE_PORT=%s\n' "${GO_BLOG_DATABASE_PORT}"
    printf 'GO_BLOG_DATABASE_NAME=%s\n' "${GO_BLOG_DATABASE_NAME}"
    printf 'GO_BLOG_DATABASE_USER=%s\n' "${GO_BLOG_DATABASE_USER}"
    printf 'GO_BLOG_DATABASE_PASSWORD=%s\n' "${GO_BLOG_DATABASE_PASSWORD}"
  } > "${tmp}"
  mv "${tmp}" "${DEPLOY_ENV_FILE}"
  chmod 600 "${DEPLOY_ENV_FILE}"
  if [[ -n "${SUDO_USER:-}" ]]; then
    chown "${SUDO_USER}:${SUDO_USER}" "${DEPLOY_ENV_FILE}"
  fi
  log "Wrote ${DEPLOY_ENV_FILE}"
}

install_packages() {
  if [[ "${SETUP_SKIP_APT}" == "1" ]]; then
    command -v git >/dev/null 2>&1 || die "git not found (install it or unset SETUP_SKIP_APT)"
    command -v curl >/dev/null 2>&1 || die "curl not found (install it or unset SETUP_SKIP_APT)"
    command -v rsync >/dev/null 2>&1 || die "rsync not found (install it or unset SETUP_SKIP_APT)"
    log "Skipping apt-get (SETUP_SKIP_APT=1)"
    return 0
  fi

  log "Installing git and prerequisites..."
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  apt-get install -y -qq git curl ca-certificates openssl rsync
}

install_etckeeper() {
  if [[ "${SETUP_SKIP_APT}" == "1" ]]; then
    return 0
  fi

  if command -v etckeeper >/dev/null 2>&1; then
    log "etckeeper is already installed"
    return 0
  fi

  log "Installing etckeeper..."
  export DEBIAN_FRONTEND=noninteractive
  apt-get update -qq
  apt-get install -y -qq etckeeper
}

install_docker() {
  if [[ "${SETUP_SKIP_DOCKER_INSTALL}" == "1" ]]; then
    log "Skipping Docker installation (SETUP_SKIP_DOCKER_INSTALL=1)"
    command -v docker >/dev/null 2>&1 || die "docker not found and installation was skipped"
    docker compose version >/dev/null 2>&1 || die "docker compose plugin not found"
    return 0
  fi

  if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then
    log "Docker is already installed"
    return 0
  fi

  log "Installing Docker..."
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
}

swap_is_configured() {
  local swap_kb
  swap_kb="$(awk '/^SwapTotal:/ {print $2}' /proc/meminfo)"
  [[ "${swap_kb:-0}" -gt 0 ]]
}

ensure_swap_in_fstab() {
  local swap_file="$1"
  if grep -qF "${swap_file}" /etc/fstab 2>/dev/null; then
    return 0
  fi
  echo "${swap_file} none swap sw 0 0" >> /etc/fstab
}

setup_swap() {
  if [[ "${SETUP_SWAP}" != "1" ]]; then
    return 0
  fi

  if swap_is_configured; then
    log "Swap is already configured — skipping"
    swapon --show >&2 || true
    return 0
  fi

  local swap_file="${SETUP_SWAP_FILE}"
  local swap_mb="${SETUP_SWAP_SIZE_MB}"

  log "Configuring ${swap_mb}MB swap at ${swap_file}..."

  if [[ -f "${swap_file}" ]]; then
    log "Swap file ${swap_file} exists — enabling it"
  elif ! fallocate -l "${swap_mb}M" "${swap_file}" 2>/dev/null; then
    log "fallocate failed; creating swap file with dd (this may take a while)..."
    dd if=/dev/zero of="${swap_file}" bs=1M count="${swap_mb}" status=none
  fi

  chmod 600 "${swap_file}"
  mkswap "${swap_file}" >/dev/null
  swapon "${swap_file}"
  ensure_swap_in_fstab "${swap_file}"

  log "Swap enabled:"
  swapon --show >&2 || true
}

ensure_infra_network() {
  if docker network inspect infra >/dev/null 2>&1; then
    log "Docker network infra already exists"
  else
    log "Creating Docker network infra..."
    docker network create infra >/dev/null
  fi

  if docker container inspect postgresql >/dev/null 2>&1; then
    log "Found postgresql container"
  else
    log "WARNING: postgresql container not found — deploy shared Postgres first if /health fails"
  fi
}

RSYNC_OPTS=(
  -a
  --delete
  --exclude '.env'
  --exclude '.repo-cache'
  --exclude '.git'
  --exclude 'node_modules'
  --exclude 'test-results'
  --exclude 'playwright-report'
)

count_rsync_changes() {
  local source_dir="$1"
  rsync "${RSYNC_OPTS[@]}" --dry-run --itemize-changes "${source_dir}/" "${DEPLOY_DIR}/" \
    | grep -vE '^(\.d\.\.t|>f\.\.t)' \
    | grep -c . || true
}

sync_compose_tree() {
  local source_dir="$1"
  mkdir -p "${DEPLOY_DIR}"

  local changes
  changes="$(count_rsync_changes "${source_dir}")"
  if [[ "${changes}" -eq 0 ]]; then
    log "Project tree already synced to ${DEPLOY_DIR}"
    printf 'current'
    return 0
  fi

  rsync "${RSYNC_OPTS[@]}" "${source_dir}/" "${DEPLOY_DIR}/"
  printf 'sync'
}

repo_source_dir() {
  if [[ -n "${REPO_SUBPATH}" ]]; then
    printf '%s/%s' "${REPO_CACHE_DIR}" "${REPO_SUBPATH}"
  else
    printf '%s' "${REPO_CACHE_DIR}"
  fi
}

fetch_existing_clone() {
  if git -C "${REPO_CACHE_DIR}" fetch --depth 1 origin "${GIT_REF}" 2>/dev/null; then
    return 0
  fi
  git -C "${REPO_CACHE_DIR}" fetch --depth 1 origin "refs/tags/${GIT_REF}:refs/tags/${GIT_REF}"
}

remote_ref_sha() {
  git -C "${REPO_CACHE_DIR}" rev-parse FETCH_HEAD 2>/dev/null \
    || git -C "${REPO_CACHE_DIR}" rev-parse "refs/tags/${GIT_REF}" 2>/dev/null \
    || git -C "${REPO_CACHE_DIR}" rev-parse "origin/${GIT_REF}" 2>/dev/null \
    || git -C "${REPO_CACHE_DIR}" rev-parse "${GIT_REF}"
}

project_worktree_clean() {
  git -C "${REPO_CACHE_DIR}" diff --quiet HEAD \
    && git -C "${REPO_CACHE_DIR}" diff --cached --quiet HEAD
}

reset_existing_clone() {
  local target_sha
  target_sha="$(remote_ref_sha)"
  git -C "${REPO_CACHE_DIR}" checkout --detach "${target_sha}"
  git -C "${REPO_CACHE_DIR}" reset --hard "${target_sha}"
}

clone_repo_cache() {
  mkdir -p "$(dirname "${REPO_CACHE_DIR}")"
  if git clone --branch "${GIT_REF}" --depth 1 "${REPO_URL}" "${REPO_CACHE_DIR}" 2>/dev/null; then
    return 0
  fi

  log "Shallow branch clone failed; fetching ${GIT_REF} by ref..."
  git clone --depth 1 "${REPO_URL}" "${REPO_CACHE_DIR}"
  fetch_existing_clone
  local target_sha
  target_sha="$(remote_ref_sha)"
  git -C "${REPO_CACHE_DIR}" checkout --detach "${target_sha}"
}

assess_existing_clone() {
  fetch_existing_clone

  local local_sha remote_sha
  local_sha="$(git -C "${REPO_CACHE_DIR}" rev-parse HEAD)"
  remote_sha="$(remote_ref_sha)"

  local source_dir
  source_dir="$(repo_source_dir)"

  if [[ "${local_sha}" == "${remote_sha}" ]] && project_worktree_clean; then
    log "Repo cache already at ${GIT_REF} (${local_sha:0:7}) in ${REPO_CACHE_DIR}"
    sync_compose_tree "${source_dir}"
    return 0
  fi

  if [[ "${local_sha}" != "${remote_sha}" ]]; then
    log "Updating repo cache ${local_sha:0:7} -> ${remote_sha:0:7}"
  else
    log "Resetting local changes in ${REPO_CACHE_DIR}"
  fi
  reset_existing_clone
  sync_compose_tree "${source_dir}"
}

deploy_from_source() {
  [[ -d "${SETUP_SOURCE_DIR}" ]] || die "SETUP_SOURCE_DIR does not exist: ${SETUP_SOURCE_DIR}"
  [[ -f "${SETUP_SOURCE_DIR}/docker-compose.yml" ]] \
    || die "SETUP_SOURCE_DIR must contain docker-compose.yml: ${SETUP_SOURCE_DIR}"
  [[ -f "${SETUP_SOURCE_DIR}/Dockerfile" ]] \
    || die "SETUP_SOURCE_DIR must contain Dockerfile: ${SETUP_SOURCE_DIR}"
  sync_compose_tree "${SETUP_SOURCE_DIR}"
}

deploy_project() {
  log "Deploying go-blog to ${DEPLOY_DIR}..."

  if [[ -n "${SETUP_SOURCE_DIR}" ]]; then
    deploy_from_source
    return 0
  fi

  if [[ -d "${REPO_CACHE_DIR}/.git" ]]; then
    assess_existing_clone
    return 0
  fi

  if [[ -d "${REPO_CACHE_DIR}" ]] && [[ -n "$(ls -A "${REPO_CACHE_DIR}" 2>/dev/null)" ]]; then
    die "${REPO_CACHE_DIR} exists but is not a git repository. Remove or rename it, then re-run."
  fi

  clone_repo_cache
  local source_dir
  source_dir="$(repo_source_dir)"
  [[ -d "${source_dir}" ]] || die "Missing deploy source path: ${source_dir}"
  [[ -f "${source_dir}/docker-compose.yml" ]] \
    || die "Missing docker-compose.yml in ${source_dir}"
  sync_compose_tree "${source_dir}"
}

compose_stack_running() {
  [[ -d "${DEPLOY_DIR}" ]] || return 1

  local services
  services="$(cd "${DEPLOY_DIR}" && docker compose ps --status running --format '{{.Service}}' 2>/dev/null)" \
    || return 1
  grep -qx 'go-blog' <<<"${services}" || return 1
}

app_is_ready() {
  compose_stack_running || return 1
  cd "${DEPLOY_DIR}"
  docker compose exec -T go-blog wget -q -O /dev/null http://127.0.0.1:8083/health
}

start_compose() {
  local rebuild="${1:-1}"

  if [[ "${rebuild}" == "1" ]]; then
    log "Building and starting docker compose stack..."
    cd "${DEPLOY_DIR}"
    docker compose up -d --build
    return 0
  fi

  log "Starting docker compose stack (no rebuild)..."
  cd "${DEPLOY_DIR}"
  docker compose up -d
}

wait_for_app() {
  log "Waiting for go-blog /health (timeout: ${APP_READY_TIMEOUT_SEC}s)..."
  local deadline=$((SECONDS + APP_READY_TIMEOUT_SEC))
  while (( SECONDS < deadline )); do
    if app_is_ready; then
      log "go-blog is ready"
      return 0
    fi
    sleep 2
  done

  log "go-blog failed to become ready; recent logs:"
  cd "${DEPLOY_DIR}"
  docker compose logs --tail=50 go-blog || true
  die "go-blog readiness check failed"
}

run_migrations() {
  if [[ "${SETUP_SKIP_MIGRATE}" == "1" ]]; then
    log "Skipping migrations (SETUP_SKIP_MIGRATE=1)"
    return 0
  fi

  log "Running database migrations..."
  cd "${DEPLOY_DIR}"
  docker compose exec -T go-blog ./blog migrate
}

main() {
  parse_args "$@"

  require_root
  setup_swap
  install_packages
  install_etckeeper
  install_docker
  ensure_infra_network

  local deploy_action
  deploy_action="$(deploy_project)"

  ensure_deploy_env

  if [[ "${SETUP_FORCE}" == "1" ]]; then
    log "SETUP_FORCE=1 — rebuilding and restarting the stack"
    start_compose 1
    wait_for_app
    run_migrations
  elif [[ "${deploy_action}" == "current" ]] \
        && compose_stack_running \
        && app_is_ready; then
    log "go-blog is already deployed at ${GIT_REF} and the stack is healthy — skipping redeploy"
    log "Use SETUP_FORCE=1 to rebuild and restart anyway"
  elif [[ "${deploy_action}" == "current" ]]; then
    log "Project tree is already current; ensuring stack is up"
    start_compose 0
    wait_for_app
    run_migrations
  else
    start_compose 1
    wait_for_app
    run_migrations
  fi

  log "Deployment complete."
  log "  Directory: ${DEPLOY_DIR}"
  log "  URL:       http://127.0.0.1:${GO_BLOG_HTTP_PORT}/"
  log "  Health:    http://127.0.0.1:${GO_BLOG_HTTP_PORT}/health"
  log "  Next step: docker compose exec go-blog ./blog users-create"
}

main "$@"
