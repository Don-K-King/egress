# Evido External-Conference Integration: Architektur-Guideline

## Zweck und Geltungsbereich

Diese Guideline definiert die **verbindliche Architektur** für eine Anbindung von Evido über einen externen VC-Provider (z. B. LiveKit) an Egress. Ziel ist die sichere und wartbare Integration ohne Aufweichung bestehender Architekturgrenzen.

**Verbindliche Leitentscheidung:**

- **Kein direkter VC → Core-Durchgriff.**
- **Ausschließlich Ingest-Gateway (DMZ) → Evido Core** über mTLS + Netzwerk-Allowlist.

Diese Entscheidung hat Vorrang vor kurzfristigen Implementierungsabkürzungen.

---

## Architekturprinzipien (verbindlich)

1. **Architektur vor Einfachheit**
   - Keine Einbettung provider-spezifischer Logik in den Core.
   - Integrationspunkte müssen über klar abgegrenzte Control- und Media-Contracts erfolgen.

2. **Security by design**
   - Jede Ingest-Anfrage muss serverseitig authentisiert, autorisiert und replay-sicher validiert werden.
   - Keine Vertrauensannahmen gegenüber Frontend/Client-Seite.

3. **Deterministische Zustandsführung**
   - Session-Bindung und Rollen-/Track-Mapping sind serverseitig eindeutig und überprüfbar.
   - Änderungen erfolgen nur über explizite Rebind-Operationen.

4. **Parität statt Sonderpfade**
   - Ergebnissemantik (Segmentierung, Persistenz, Statusübergänge) muss zum Standardmodus äquivalent bleiben.

---

## Zielarchitektur

## 1) Control Plane vs. Media Plane

### Control Plane

Verantwortlich für Session-Lifecycle, Binding-Verwaltung, Consent-/Policy-Status und Webhooks.

Vorgesehene Endpunkte:

- `POST /plugins/external_conference/session/bind`
- `POST /plugins/external_conference/session/start`
- `POST /plugins/external_conference/session/stop`
- `POST /plugins/external_conference/events`

### Media Plane

Verantwortlich für den kontinuierlichen Audio-Ingest mit Backpressure und Zustandskopplung an serverseitiges Binding.

Primärprotokoll:

- `grpc externalconference.MediaStream/StreamAudio` (bidirektional)

Alternative:

- `GET/WS /plugins/external_conference/stream` (binary)

Fallback (nicht Standard):

- `POST /plugins/external_conference/segment`

**Architekturregel:** Segmentierungslogik bleibt im Core; das Gateway normalisiert Ingest, segmentiert aber nicht primär.

## 2) Netz- und Trust-Boundary

1. VC-Provider spricht nur mit Ingest-Gateway in der DMZ.
2. Ingest-Gateway spricht nur mit explizit erlaubten Core-Zielen/Ports.
3. Core ist nicht direkt vom VC erreichbar.
4. Alle DMZ→Core-Verbindungen sind mTLS-geschützt.

---

## Security-Design (verbindlich)

## 3) Bedrohungen

1. Fremdeinspeisung in Sessions (forged ingest)
2. Replay von Webhooks/Control-Aufrufen
3. Secret-Leak über Logs/Tracing
4. Rollenvertauschung durch fehlerhaftes Track-Mapping
5. Seitwärtsbewegung von DMZ in Core
6. Versehentliches Recording im Nicht-Einvernahme-Modus

## 4) Mindestkontrollen

- Signierte Requests (`timestamp`, `nonce`, `signature`) mit Replay-Schutz
- mTLS zwischen Ingest-Gateway und Core
- Scoped short-lived Service-Tokens (keine globalen statischen Tokens)
- Netzwerk-Allowlist DMZ→Core
- Harte Session-Bindung (`conferenceId ↔ sessionId`) inkl. Statusvalidierung
- Audit-Log-Redaction für `authorization`, `token`, `api_key`
- Policy-Flag: Recording/Transkription nur im Einvernahmeraum aktiv

## 5) Immutable Binding + Signaturpflicht

Für jeden aktiven Teilnehmer-Track ist ein **immutable Binding-Artefakt** erforderlich:

- `binding_id`
- `evido_session_id`
- `conference_id`
- `participant_id`
- `role`
- `track_id`
- `issued_at`
- `expires_at`
- `kid`

Regeln:

1. Nach Ausstellung ist das Binding unveränderlich.
2. Änderungen erfolgen nur über Rebind mit neuer `binding_id`.
3. Signaturprüfung erfolgt ausschließlich serverseitig gegen Key-Registry.
4. Ohne gültige Signatur: Hard-Fail + Audit-Eintrag.

## 6) Rotation und Widerruf

- Versionierte Schlüssel mit `kid` und definierter Verify-Grace-Period
- Rebind/Rejoin erzeugt immer neue `binding_id`; alte Bindings sofort ungültig
- Widerruf wirkt near-real-time (Revocation-Store + Cache-Invalidation)
- Rotation/Widerruf erzeugen Audit-Events mit `binding_id`, `kid`, Auslöser, Zeitstempel

---

## Ingest-Akzeptanzregeln (verbindlich)

Audio-Ingest wird nur akzeptiert, wenn **alle** Bedingungen erfüllt sind:

1. `conferenceId ↔ sessionId` ist erfolgreich gebunden
2. Claim ist signaturgültig, nicht abgelaufen, nicht widerrufen
3. `role` und `track_id` sind konsistent zur serverseitigen Binding-Relation
4. `binding_id` ist aktiv und nicht superseded

Bei Verstoß: Ablehnung + normierter Audit-Grundcode.

Mindest-Fehlercodes:

- `400` invalid payload/mapping
- `401/403` auth/authz failed
- `404` session not found
- `409` session-conference conflict / invalid state transition
- `422` signature/replay validation failed
- `429` rate-limited ingest

---

## Architekturkonflikte und verbotene Fehlstrategien

## 7) Explizit nicht zulässig

1. **Direkte VC-Integration im Core-Prozess**
   - Konflikt: verletzt DMZ/Core-Trennung und erhöht Blast Radius.
2. **Provider-spezifische Sonderpfade in UI/Core-Domäne**
   - Konflikt: hoher Wartungsaufwand, sinkende Austauschbarkeit.
3. **Clientseitige oder optionale Signaturprüfung**
   - Konflikt: Umgehbarkeit, keine belastbare Security.
4. **Soft-Expiry ohne unmittelbare Invalidierung alter Bindings**
   - Konflikt: Window für Missbrauch nach Rollen-/Track-Wechsel.

## 8) Alternative bei Infrastruktur-Blockern

Wenn bidirektionales gRPC organisatorisch nicht verfügbar ist:

- temporär WS-binary als Alternative verwenden,
- semantisches Framing identisch halten (`sequenceNo`, `timestamp`, `trackId`),
- spätere Migration auf gRPC vertraglich einplanen,
- Segment-POST nur als Notfall-/Legacy-Fallback.

---

## Umsetzungsleitplanken

## 9) Chronologische Umsetzung

### Phase 0 – Spezifikation/ADR

1. ADR für DMZ/Core, Control-/Media-Trennung, Gateway als Normalizer
2. API-/Event-Verträge + Fehlercodematrix
3. Security-Spec für Signatur, Nonce-TTL, Rotation, mTLS

### Phase 1 – Fail-first Tests

1. Negative Security-Tests:
   - manipulierte Rolle
   - Replay
   - fremde Session
   - abgelaufene Claims
   - Track-Swap
2. Streaming-Resilienztests:
   - Backpressure
   - Reconnect
   - Duplicate/Out-of-Order
3. Paritätstest Segmentgrenzen Standard vs. External

### Phase 2 – Backend Foundations

1. Plugin/Sessionmode `external_conference`
2. Binding-Service + Role/Track-Mapping
3. Reuse bestehender Persistenzpfade (Segment-/Transcript-Outputs)

### Phase 3 – Ingest Gateway

1. Separater DMZ-Dienst mit Adapter-Schnittstelle
2. Provider-Adapter v1
3. Signierte Ingest-Calls + mTLS Uplink
4. Observability + redacted Logs

### Phase 4 – Frontend-Parität

1. Plugin-gated Mode-Option
2. SessionCore-Reuse
3. Sichtbarer Security-/Consent-Status

---

## Governance und Betrieb

- Sicherheitsfreigabe pro Provider-Adapter
- Automatisierte, getestete Secret-/Key-Rotation
- Auditpflicht für Bindings, Consent, Start/Stop, jeden Ingest accept/reject
- Incident Playbooks für:
  - ingest compromised
  - VC outage
  - mapping corruption
- Pflichtmetriken:
  - Stream-Liveness
  - Frame-Lag p50/p95/p99
  - Frame-Drop-Rate je Ursache
  - Retry-/Reconnect-Burst und Circuit-Breaker-Events

---

## Abnahmekriterien (Definition of Done)

Eine Implementierung gilt nur als architekturkonform, wenn:

1. Kein VC→Core-Direktzugriff existiert.
2. Ingest nur via mTLS + Signatur + Replay-Schutz akzeptiert wird.
3. Immutable Binding inkl. Rebind/Widerruf technisch durchgesetzt ist.
4. Audit- und Redaction-Pflichten nachweisbar umgesetzt sind.
5. Fail-first Sicherheits- und Resilienztests grün sind.
6. Ergebnisparität zum Standardmodus nachgewiesen ist.
