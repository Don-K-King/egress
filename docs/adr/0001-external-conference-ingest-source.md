# ADR 0001: External Conference Ingest als eigener Request-/Source-Pfad

## Status
Accepted

## Kontext
Bisherige Egress-Pfade unterscheiden primär zwischen `web` (browser-basiert) und `sdk` (room subscription). Für ingest-zentrierte externe Konferenzen entstehen dabei Architekturkonflikte:
- semantische Vermischung mit Web-URLs,
- unklare Verantwortlichkeiten zwischen Control Plane (Orchestrierung) und Media Plane (Track-Verarbeitung),
- fehlende harte Eingangsvalidierung für ingest-spezifische Identitäten.

## Entscheidung
Wir führen explizite Typen ein:
- `RequestTypeExternalConference` (`external_conference`)
- `SourceTypeExternalIngest` (`external_ingest`)

Routing erfolgt in `PipelineConfig.Update` über einen dedizierten Pfad mit strikter Validierung der Parameter:
- `sessionId`
- `conferenceId`
- `participantId`
- `trackId`
- `role` (`audio`|`video`)

Die eigentliche Source-Integration erfolgt über eine eigene Implementierung `ExternalIngestSource`, injiziert ausschließlich über `source.New(...)`.
Der Controller-Flow bleibt unverändert.

## Konsequenzen
### Positiv
- Klare Architekturgrenze für ingest-orientierte Flows.
- Geringeres Risiko von Seiteneffekten auf bestehende Web-/SDK-Pfade.
- Verbesserte Eingangsvalidierung reduziert Missbrauchs- und Fehlkonfigurationsrisiken.

### Trade-off
- Bis zur Erweiterung der öffentlichen RPC-Protos wird der Trigger noch über `WebEgressRequest.Url` Query-Parameter erkannt. Das ist als Übergangsmechanismus dokumentiert und sollte in einem Folge-Change durch ein eigenes RPC-Oneof ersetzt werden.

## Security-Betrachtung
- Whitelist-Validierung (`^[A-Za-z0-9_-]+$`) für ingest-IDs verhindert problematische Zeichenfolgen in Dateinamen/Logs/Templates.
- `role` ist auf bekannte Werte beschränkt; unbekannte Rollen werden hard-fail abgewiesen.

## Alternativen
1. **SDK-Pfad überladen**: verworfen, da Domänensemantik verloren geht.
2. **Web-Pfad beibehalten**: verworfen, da Browser- und Ingest-Pipelines unterschiedliche Betriebsmodelle besitzen.
3. **Sofort neues RPC-Oneof**: architektonisch sauber, aber außerhalb des aktuellen Repos/Scopes.
