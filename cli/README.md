# multi-git CLI (Fase 2 — em desenvolvimento)

CLI multiplataforma (Go) para automatizar a configuração documentada na Fase 1.

## Comandos

```bash
multi-git detect   # mostra chaves SSH, CLIs (gh/az) e config existente
multi-git apply    # fluxo interativo com backup e confirmação
multi-git verify   # valida arquivos e identidade por pasta
```

Flags:
- `--dry-run` — não escreve arquivos
- `--yes` — responde sim a todas as confirmações

## Executar sem instalar

Linux/macOS/WSL:

```bash
curl -fsSL https://raw.githubusercontent.com/guionardo/multi-git/main/cli/scripts/install.sh | bash
```

Com argumentos:

```bash
curl -fsSL https://raw.githubusercontent.com/guionardo/multi-git/main/cli/scripts/install.sh | bash -s -- apply --dry-run
```

Windows (PowerShell):

```powershell
irm https://raw.githubusercontent.com/guionardo/multi-git/main/cli/scripts/install.ps1 | iex
```

## Segurança

- Sempre faz backup de arquivos existentes (`.bak.<timestamp>`).
- Pede confirmação antes de sobrescrever.
- Nunca altera chaves privadas.
- Use `--dry-run` para revisar.