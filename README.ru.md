# Grafana Git Sync

**Git → Sync → Grafana API.** Автоматическая синхронизация дашбордов из git-репозитория
в Grafana по API, с real-time обновлением на каждом коммите.

> 📚 **Полная документация:** [`docs/shared/components/grafana-git-sync/`](docs/shared/components/grafana-git-sync/)
> (git-submodule из [`vpnmesh/docs`](https://github.com/vpnmesh/docs) — общий источник правды).

## Быстрый старт

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

## Сборка

```bash
make build          # бинарник в bin/
make docker-build   # Docker-образ
make test           # тесты
```

## Документация

| Что нужно | Где смотреть |
|---|---|
| Все переменные окружения | [`docs/shared/components/grafana-git-sync/docs/configuration.md`](docs/shared/components/grafana-git-sync/docs/configuration.md) |
| Как устроено внутри | [`docs/shared/components/grafana-git-sync/docs/architecture.md`](docs/shared/components/grafana-git-sync/docs/architecture.md) |
| Деплой (Docker / K8s / systemd) | [`docs/shared/components/grafana-git-sync/docs/deployment/`](docs/shared/components/grafana-git-sync/docs/deployment/) |
| Траблшутинг | [`docs/shared/components/grafana-git-sync/docs/troubleshooting.md`](docs/shared/components/grafana-git-sync/docs/troubleshooting.md) |

Не забудь подтянуть submodule:

```bash
git submodule update --init --recursive
# или сразу при клонировании:
git clone --recurse-submodules <repo-url>
```

## Контрибьютинг

См. [`CONTRIBUTING.md`](CONTRIBUTING.md). Если меняешь публичный контракт (CLI, env-переменные,
API, поведение sync) — обновляй документацию в [`vpnmesh/docs`](https://github.com/vpnmesh/docs)
**отдельным PR**.

## Лицензия

[MIT](LICENSE)
