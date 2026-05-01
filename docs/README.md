# `docs/`

Эта папка содержит только submodule `shared/`, указывающий на [`vpnmesh/docs`](https://github.com/vpnmesh/docs)
— единый источник правды документации продукта VPNMesh.

- Документация по `grafana-git-sync` — в [`shared/components/grafana-git-sync/`](shared/components/grafana-git-sync/).
- Глобальный индекс / навигация — в [`shared/INDEX.md`](shared/INDEX.md) и [`shared/AGENTS.md`](shared/AGENTS.md).

> **Не правь файлы внутри `shared/` через этот репо.** Изменения делаются PR'ом в `vpnmesh/docs`.
> Здесь submodule только обновляется до нового коммита (CI делает это автоматически раз в день).
