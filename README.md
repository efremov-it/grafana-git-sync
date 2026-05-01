# Grafana Git Sync

**Git → Sync → Grafana API.** Автоматическая синхронизация дашбордов из git-репо в Grafana через API,
с real-time обновлением на каждом коммите.

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

> 📚 **Полная документация:** [`docs/shared/components/grafana-git-sync/`](docs/shared/components/grafana-git-sync/)
> (подключена как git-submodule из [`vpnmesh/docs`](https://github.com/vpnmesh/docs)).

## Quick start

```bash
docker run -d \
  --name grafana-git-sync -p 8080:8080 \
  -e GIT_REPO_URL=ssh://git@github.com/your-org/dashboards.git \
  -e GIT_BRANCH=main \
  -e GIT_SSH_KEY="$(cat ~/.ssh/id_rsa)" \
  -e GRAFANA_URL=http://localhost:3000 \
  -e GF_SECURITY_ADMIN_USER=admin \
  -e GF_SECURITY_ADMIN_PASSWORD=admin \
  grafana-git-sync:latest
```

Примеры docker-compose / k8s — в [`examples/`](examples/).

## Build

```bash
make build          # бинарник в bin/
make docker-build   # docker-образ
make test           # тесты
```

## Documentation

| Что мне нужно | Где | 
|---|---|
| Все env-переменные | [`docs/shared/components/grafana-git-sync/docs/configuration.md`](docs/shared/components/grafana-git-sync/docs/configuration.md) |
| Как это работает внутри | [`docs/shared/components/grafana-git-sync/docs/architecture.md`](docs/shared/components/grafana-git-sync/docs/architecture.md) |
| Деплой (Docker / K8s / systemd) | [`docs/shared/components/grafana-git-sync/docs/deployment/`](docs/shared/components/grafana-git-sync/docs/deployment/) |
| Решение проблем | [`docs/shared/components/grafana-git-sync/docs/troubleshooting.md`](docs/shared/components/grafana-git-sync/docs/troubleshooting.md) |

После клонирования репо:

```bash
git submodule update --init --recursive
# или сразу
git clone --recurse-submodules <repo-url>
```

## Contributing

См. [`CONTRIBUTING.md`](CONTRIBUTING.md). Если меняешь публичный контракт (CLI, env, API) —
обновляй документацию в [`vpnmesh/docs`](https://github.com/vpnmesh/docs) **отдельным PR**.

## License

[MIT](LICENSE)
