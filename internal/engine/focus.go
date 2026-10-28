package engine

// FocusOrder returns the current visible controls in document/layout traversal
// order. Hidden and disabled controls never receive hotspots from RenderFrame;
// any hotspot with no visible attributed cell is omitted. A new slice is
// returned so the host cannot mutate the underlying frame's focus metadata.
//
// SML supplies presentation metadata only. Key events and action authorization
// remain responsibilities of the embedding application.
func (f Frame) FocusOrder() []Hotspot {
	result := make([]Hotspot, 0, len(f.Hotspots))
	seen := make(map[string]bool, len(f.Hotspots))
	for _, spot := range f.Hotspots {
		if spot.ID == "" || seen[spot.ID] || spot.Line < 0 || spot.Line >= f.Height ||
			spot.Start < 0 || spot.End > f.Width || spot.Start >= spot.End {
			continue
		}
		if len(f.Cells) > 0 {
			if spot.Line >= len(f.Cells) {
				continue
			}
			visible := false
			for x := spot.Start; x < spot.End && x < len(f.Cells[spot.Line]); x++ {
				if f.Cells[spot.Line][x].ElementID == spot.ID {
					visible = true
					break
				}
			}
			if !visible {
				continue
			}
		}
		seen[spot.ID] = true
		spot.Options = append([]string(nil), spot.Options...)
		result = append(result, spot)
	}
	return result
}

// NextFocus resolves the next or previous focusable control, wrapping at the
// ends. When the current ID is missing (for instance after navigation), it
// returns the first or last visible control. It does not consume key events.
func (f Frame) NextFocus(currentID string, reverse bool) (Hotspot, bool) {
	ordered := f.FocusOrder()
	if len(ordered) == 0 {
		return Hotspot{}, false
	}
	for i, spot := range ordered {
		if spot.ID != currentID {
			continue
		}
		if reverse {
			return ordered[(i+len(ordered)-1)%len(ordered)], true
		}
		return ordered[(i+1)%len(ordered)], true
	}
	if reverse {
		return ordered[len(ordered)-1], true
	}
	return ordered[0], true
}
