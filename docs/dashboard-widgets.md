# Terminal dashboard widgets (SML v3)

Six native presentation-only components complement the existing `gauge`, `sparkline`, `progress`, `metric` and `table`. These widgets are safe to use in regular views and plugin views. They do **not** execute code, access sockets or subscribe independently to events. The Go host supplies the bindings and controls repaint frequency.

```xml
<screen name="visuals" version="3">
  <kpi label="Connections" value="{{sessions.online}}" delta="+3%" unit="users"/>
  <heatmap label="Activity intensity" values="{{metrics.activity}}" columns="12"/>
  <bar-chart label="Per service" labels="API,SSH,Web" values="34,22,18" width="20"/>
  <stacked-bar label="Memory mix" labels="Used,Cache,Free" values="45,20,35" width="24"/>
  <waveform label="Signal" values="{{metrics.samples}}" width="30" height="4"/>
  <activity-feed label="Recent events" events="{{metrics.events}}" max-items="5"/>
</screen>
```

| Component | Input | Output |
| --- | --- | --- |
| `kpi` | `label`, `value`, optional numeric `delta` and `unit` | Compact value and directional delta |
| `heatmap` | Comma-separated `values` between 0 and 100 | Dense shaded cell activity grid |
| `bar-chart` | Numeric `values` and optional comma-separated `labels` | Bounded relative horizontal bars |
| `stacked-bar` | Positive numeric `values`, `labels` | Proportional bar with legend |
| `waveform` | Numeric time series `values`, optional `width` and `height` | Small multirow glyph plot |
| `activity-feed` | Semicolon-separated `HH:MM\|state\|message` entries | Icon-coded event rows |

Sizing: heatmap `columns=2..48`, bar/stacked `width=4..60`, waveform `width=8..60` and `height=2..8`, activity feed `max-items=1..12`. The parser rejects invalid attributes; numerical inputs are capped (96 heatmap, 12 bars, 8 stacked groups, 128 waveform observations, 64 feed events). Rendered rows respect SSH cell widths. Invalid or missing values show non-crashing empty states. The widgets are static per frame; hosts can supply new values and render again with existing SML frame-diff handling.

### Backward compatibility
Existing SML v1/v2/v3 views render unchanged. Hosts using an older SML module must upgrade before loading views containing these tags. No CSS, arbitrary ANSI, script expressions, filesystem paths or new UI actions are accepted.

### Tests

```sh
go test ./...
go vet ./...
```

Check behavior in 20, 40, 80 and 120 terminal columns. Test on an actual PuTTY session for optional Unicode glyph fallback; the SSH host controls terminal profile and replaces incompatible glyphs where necessary.

## Live preview in Sentinel

A first-party, read-only SML Showcase is being added in the Sentinel core. It documents the currently pinned engine and labels these six new elements as requiring an engine module update. Once the new SML release is adopted, the showcase can render these widgets directly using the same bounded host fixtures.
