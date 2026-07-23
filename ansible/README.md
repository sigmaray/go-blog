# Ansible VPS deploy

Declarative alternative to the bash bootstrap scripts:

- [`docs/example-postgresql-docker-compose/scripts/setup-vps.sh`](../docs/example-postgresql-docker-compose/scripts/setup-vps.sh)
- [`scripts/setup-vps.sh`](../scripts/setup-vps.sh)

Use this when you prefer Ansible over `curl | bash`. The bash scripts remain the simplest one-liner path for a fresh Ubuntu VPS.

## Requirements

- Ansible 2.14+ on the control machine (`pip install ansible` or distro packages)
- Target: Ubuntu VPS with SSH (or `localhost` for CI / local smoke tests)
- Root or passwordless sudo on the target

## Quick start

```bash
cd ansible

# Edit inventory/example.ini (host + user), then:
cp inventory/example.ini inventory/production.ini
# edit production.ini

ansible-playbook -i inventory/production.ini playbooks/setup-all.yml

# Optional swap (same idea as --swap on the bash scripts):
ansible-playbook -i inventory/production.ini playbooks/setup-all.yml -e setup_swap=true
```

Deploy only one stack:

```bash
ansible-playbook -i inventory/production.ini playbooks/setup-postgresql.yml
ansible-playbook -i inventory/production.ini playbooks/setup-go-blog.yml
```

Local smoke test (same host, like CI):

```bash
ansible-playbook playbooks/setup-all.yml \
  -e postgresql_deploy_dir=/tmp/postgresql-deploy \
  -e go_blog_deploy_dir=/tmp/go-blog-deploy \
  -e go_blog_postgres_env_file=/tmp/postgresql-deploy/.env \
  -e setup_skip_apt=true \
  -e setup_skip_docker_install=true \
  -e postgresql_setup_source_dir="$PWD/../docs/example-postgresql-docker-compose" \
  -e go_blog_setup_source_dir="$PWD/.." \
  -e postgresql_password=ci-test-postgres-password
```

## Molecule + Incus (isolated VPS-like test)

Runs the real playbooks inside a nested Ubuntu Incus container (installs Docker, deploys Postgres + go-blog). Requires Incus on the host and membership in `incus-admin`.

The Molecule instance is launched **privileged** with AppArmor unconfined so Docker-in-Incus can start containers (avoids `ip_unprivileged_port_start` / runc errors on Ubuntu 24.04).

```bash
# One-time host setup (Ubuntu)
sudo apt-get install -y incus incus-client
sudo usermod -aG incus-admin "$USER"   # then re-login or: newgrp incus-admin
incus admin init --minimal

cd ansible
python3 -m venv .venv
.venv/bin/pip install -r requirements-molecule.txt
.venv/bin/ansible-galaxy collection install -r molecule/default/collections.yml

# Full create → converge → verify → destroy
sg incus-admin -c '.venv/bin/molecule test'

# Or step by step:
sg incus-admin -c '.venv/bin/molecule create'
sg incus-admin -c '.venv/bin/molecule converge'
sg incus-admin -c '.venv/bin/molecule verify'
sg incus-admin -c '.venv/bin/molecule destroy'
```

Scenario files live under `molecule/default/`. The instance is named `molecule-go-blog`; the repo is mounted at `/opt/src` inside it.

## Layout

```
ansible/
├── ansible.cfg
├── requirements-molecule.txt
├── molecule/default/     # Incus Molecule scenario
├── inventory/
│   ├── localhost.ini     # connection: local (CI)
│   ├── example.ini       # copy to production.ini
│   └── group_vars/all.yml
├── playbooks/
│   ├── setup-postgresql.yml
│   ├── setup-go-blog.yml
│   └── setup-all.yml
└── roles/
    ├── common/           # apt, etckeeper, Docker, swap
    ├── postgresql/
    └── go_blog/
```

## Variables

Shared (`group_vars/all.yml` or `-e`):

| Variable | Default | Bash equivalent |
|----------|---------|-----------------|
| `setup_skip_apt` | `false` | `SETUP_SKIP_APT` |
| `setup_skip_docker_install` | `false` | `SETUP_SKIP_DOCKER_INSTALL` |
| `setup_force` | `false` | `SETUP_FORCE` |
| `setup_swap` | `false` | `SETUP_SWAP` / `--swap` |
| `setup_swap_size_mb` | `2048` | `SETUP_SWAP_SIZE_MB` |
| `repo_url` | go-blog GitHub URL | `REPO_URL` |
| `git_ref` | `main` | `GIT_REF` |

PostgreSQL role:

| Variable | Default | Bash equivalent |
|----------|---------|-----------------|
| `postgresql_deploy_dir` | `~/r/d/postgresql` | `DEPLOY_DIR` |
| `postgresql_setup_source_dir` | empty (git clone) | `SETUP_SOURCE_DIR` |
| `postgresql_password` | auto / existing `.env` | `POSTGRES_PASSWORD` |
| `postgresql_user` | `postgres` | `POSTGRES_USER` |
| `postgresql_port` | `5432` | `POSTGRES_PORT` |

go-blog role:

| Variable | Default | Bash equivalent |
|----------|---------|-----------------|
| `go_blog_deploy_dir` | `~/r/d/go-blog` | `DEPLOY_DIR` |
| `go_blog_setup_source_dir` | empty (git clone) | `SETUP_SOURCE_DIR` |
| `go_blog_postgres_env_file` | `~/r/d/postgresql/.env` | `POSTGRES_ENV_FILE` |
| `go_blog_http_port` | `8083` | `GO_BLOG_HTTP_PORT` |
| `go_blog_session_secret` | auto / existing `.env` | `GO_BLOG_SESSION_SECRET` |
| `go_blog_database_password` | from Postgres `.env` | `GO_BLOG_DATABASE_PASSWORD` |
| `go_blog_skip_migrate` | `false` | `SETUP_SKIP_MIGRATE` |

`*_setup_source_dir` must be a path **on the target host** (used in CI). Production deploys normally clone from `repo_url`.
