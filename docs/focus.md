# Host-owned keyboard focus metadata

The optional Go Frame.FocusOrder() method returns a copy of the currently
visible interactive Hotspots in stable component/document traversal order.
RenderFrame excludes disabled and conditional-hidden controls already; focus
metadata additionally excludes off-screen, invalid and non-painted hotspots.

Frame.NextFocus(currentID, reverse) resolves the adjacent hotspot with wrap.
When currentID is absent (for example, after changing screens), forward focus
starts at the first visible control and reverse focus at the last. An empty
frame returns (Hotspot{}, false).

Example:

    frame, err := view.RenderFrame(bindings, 80)
    if err != nil { return err }

    // The host chooses whether Tab / Shift+Tab trigger navigation:
    focused, ok := frame.NextFocus(currentFocusID, shiftTabPressed)
    if ok {
        currentFocusID = focused.ID
        // The host draws a focus indicator and handles keyboard actions.
    }

These APIs do not read input, move the terminal cursor, persist state, run
application code or grant execution permissions. The host always checks
authorization when activating a typed action descriptor.

Use explicit stable IDs on nonrepeated controls. Repeated component IDs are
currently path/iteration-scoped, so their identities can change if source
records are reordered: future keyed-repeat contracts will address that.
Focus traps, groups, shortcuts and dialog restoration remain future work.

See docs/go-api.md for the existing Frame, Hotspot and RenderFrame contracts.
