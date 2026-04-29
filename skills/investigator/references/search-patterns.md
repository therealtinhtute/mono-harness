# Search Patterns

## By Language

### TypeScript/JavaScript
| Pattern | Search |
|---------|--------|
| Component | `*.tsx`, `components/` |
| Hook | `use*.ts`, `hooks/` |
| API route | `api/`, `routes/` |
| Config | `*.config.*`, `config/` |
| Test | `*.test.*`, `*.spec.*` |

### Go
| Pattern | Search |
|---------|--------|
| Handler | `*Handler`, `*_handler.go` |
| Service | `*Service`, `*_service.go` |
| Model | `model/`, `models/` |
| Config | `config/`, `*.yaml` |

### Python
| Pattern | Search |
|---------|--------|
| View | `views.py`, `*_view.py` |
| Model | `models.py`, `*_model.py` |
| API | `api/`, `endpoints/` |

---

## By Task Type

| Task | Where to Look |
|------|---------------|
| Auth | `auth/`, `middleware/`, `jwt`, `session` |
| Database | `db/`, `models/`, `migrations/`, `prisma/` |
| API | `api/`, `routes/`, `handlers/`, `controllers/` |
| Config | Root `*.config.*`, `.env*`, `config/` |
| Styles | `styles/`, `*.css`, `*.scss`, `theme/` |

---

## Common Excludes

Always filter out:
```
node_modules/
vendor/
dist/
build/
.git/
*.min.js
*.map
```
