package ui
import "errors"

// Tracks both components of any drag-and-drop operation:
// the thing being dragged (Incoming), and the destination of the drag (Target)
type DragManager struct {
    Dragging bool
    Target   interface{}
    Incoming interface{}
}

func (dm *DragManager) StartDrag(incoming interface{}) {
    dm.Dragging = true
    dm.Incoming = incoming

    // TODO: listen harder?
}

func (dm *DragManager) SetTarget(target interface{}) {
    dm.Target = target
}

// this needs significant reconsideration
func (dm *DragManager) EndDrag() error {
    // end the drag no matter what
    t, i := dm.Target, dm.Incoming
    dm.Target, dm.Incoming = nil, nil

    if !dm.Dragging {
        return errors.New("Cannot end drag: not currently dragging")
    }
    dm.Dragging = false

    if i == nil {
        return errors.New("Cannot end drag: no incoming window")
    }

    if t == nil {
        return errors.New("Cannot end drag: no target window")
    }
    return nil
}
