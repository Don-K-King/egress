# Agent.md

## Pflichtregeln bei Änderungen

1. **Regressionstests sind Pflicht** bei:
   - Pipeline-Änderungen
   - Build-/CI-Prozessänderungen
   - strukturellen/architektonischen Änderungen

2. **Dokumentationspflicht** bei nutzer- oder betreiberrelevantem Verhalten:
   - API-/Request-Verhalten
   - neue Routingpfade
   - neue Architekturgrenzen (ADR + Contract-Dokument)

## Mindest-Regressionstests für Pipeline-Änderungen
- `go test ./pkg/config ./pkg/pipeline/source`
- Falls Routing betroffen: explizite Tests für gültige und ungültige Eingaben.

## Sicherheits-Checkliste
- Eingaben auf erlaubte Formate validieren (Whitelist bevorzugt).
- Keine impliziten Fallbacks bei sicherheitskritischen Parametern.
- Fehler früh und eindeutig zurückgeben (`ErrInvalidInput`).
