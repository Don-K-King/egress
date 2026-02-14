# Contract: External Conference (Control Plane vs. Media Plane)

## Ziel
Dokumentiert die stabile Schnittstelle zwischen Steuerung (Control Plane) und Medienverarbeitung (Media Plane) für `external_conference`.

## Control Plane
### Eingang
Request wird als `StartEgressRequest.Web` angenommen, wenn `requestType=external_conference` in der URL-Query gesetzt ist.

### Pflichtparameter
- `sessionId`
- `conferenceId`
- `participantId`
- `trackId`
- `role` (`audio`|`video`)

### Validierung
- Alle Felder sind mandatory.
- `sessionId`, `conferenceId`, `participantId`, `trackId` müssen `^[A-Za-z0-9_-]+$` erfüllen.
- `role` muss `audio` oder `video` sein.
- Bei Verstoß: `ErrInvalidInput(<field>)`.

### Ausgabe der Control Plane
- `RequestType = external_conference`
- `SourceType = external_ingest`
- `Info.RoomName = conferenceId`
- audio/video Flags gemäß `role`

## Media Plane
### Source-Auswahl
`source.New(...)` wählt `ExternalIngestSource` wenn `SourceType=external_ingest`.

### Verhalten
`ExternalIngestSource` kapselt den ingest-zentrierten Fluss und verwendet intern SDK-basierte Track-Subscription, ohne den Controller-Flow zu ändern.

### Invarianten
- Keine direkte Änderung in `controller.go` erforderlich.
- Start/Stop/Timestamps folgen weiterhin dem `Source`-Interface.
