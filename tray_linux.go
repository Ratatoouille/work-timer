//go:build linux

package main

import (
	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const (
	sniIface    = "org.kde.StatusNotifierItem"
	sniPath     = dbus.ObjectPath("/StatusNotifierItem")
	watcherDest = "org.kde.StatusNotifierWatcher"
	watcherPath = dbus.ObjectPath("/StatusNotifierWatcher")
)

// linuxTrayBackend реализует trayBackend через собственный DBus
// StatusNotifierItem. В отличие от fyne.io/systray он выставляет свойство
// XAyatanaLabel, которое GNOME (ubuntu-appindicators) рендерит как нативную
// текстовую метку рядом с иконкой — чёткий текст вместо пиксельного.
type linuxTrayBackend struct {
	loc     Locale
	ready   chan struct{}
	stateCh chan trayRender
	done    chan struct{}
	conn    *dbus.Conn
	props   *prop.Properties
}

type trayRender struct {
	label    string
	tooltip  string
	dayEnded bool
}

// PX — пиктограмма для DBus (соответствует a(iiay)).
type PX struct {
	W, H int
	Pix  []byte
}

// trayTooltip — формат свойства ToolTip (sa(iiay)ss).
type trayTooltip struct {
	V0 string
	V1 []PX
	V2 string
	V3 string
}

// startTray запускает трей в отдельной горутине. Возвращает nil, если трей
// недоступен (в этом случае приложение продолжает работать без него).
func startTray(cfg Config, loc Locale) *TrayManager {
	backend := &linuxTrayBackend{
		loc:     loc,
		ready:   make(chan struct{}),
		stateCh: make(chan trayRender, 8),
		done:    make(chan struct{}),
	}

	go backend.run()

	return NewTrayManager(backend, loc)
}

// transparentPixmap — прозрачная иконка 1x1 в формате ARGB (как ждёт SNI).
func transparentPixmap() []PX {
	return []PX{{W: 1, H: 1, Pix: []byte{0, 0, 0, 0}}}
}

func (b *linuxTrayBackend) run() {
	defer close(b.done)

	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return
	}
	b.conn = conn
	defer func() { _ = conn.Close() }()

	propsSpec := map[string]map[string]*prop.Prop{
		sniIface: {
			"Category":      {Value: "ApplicationStatus", Emit: prop.EmitTrue},
			"Id":            {Value: "work-timer", Emit: prop.EmitTrue},
			"Title":         {Value: "work-timer", Emit: prop.EmitTrue},
			"Status":        {Value: "Active", Emit: prop.EmitTrue},
			"WindowId":      {Value: int32(0), Emit: prop.EmitTrue},
			"IconName":      {Value: "", Emit: prop.EmitTrue},
			"IconThemePath": {Value: "", Emit: prop.EmitTrue},
			"IconPixmap":    {Value: transparentPixmap(), Writable: true, Emit: prop.EmitTrue},
			"Menu":          {Value: dbus.ObjectPath("/NO_DBUSMENU"), Emit: prop.EmitTrue},
			"ItemIsMenu":    {Value: false, Emit: prop.EmitTrue},
			"XAyatanaLabel": {Value: b.loc.TrayNoData, Writable: true, Emit: prop.EmitTrue},
			"ToolTip":       {Value: trayTooltip{V3: b.loc.TrayNoData}, Writable: true, Emit: prop.EmitTrue},
		},
	}

	props, err := prop.Export(conn, sniPath, propsSpec)
	if err != nil {
		return
	}
	b.props = props

	if err := conn.Export(&sniObject{}, sniPath, sniIface); err != nil {
		return
	}
	if err := exportIntrospectable(conn); err != nil {
		return
	}

	_ = conn.Object(watcherDest, watcherPath).
		Call("org.kde.StatusNotifierWatcher.RegisterStatusNotifierItem", 0, conn.Names()[0]).Err

	close(b.ready)

	for {
		select {
		case <-b.done:
			return
		case r := <-b.stateCh:
			b.apply(r)
		}
	}
}

func (b *linuxTrayBackend) apply(r trayRender) {
	if b.props == nil {
		return
	}
	label := r.label
	if label == "" {
		label = "--:--"
	}
	b.props.SetMust(sniIface, "XAyatanaLabel", label)
	b.props.SetMust(sniIface, "Title", label)
	b.props.SetMust(sniIface, "ToolTip", trayTooltip{V3: r.tooltip})
}

// Apply реализует trayBackend.
func (b *linuxTrayBackend) Apply(label, tooltip string, dayEnded bool) {
	select {
	case <-b.ready:
	case <-b.done:
		return
	}
	select {
	case b.stateCh <- trayRender{label: label, tooltip: tooltip, dayEnded: dayEnded}:
	default:
	}
}

// Quit реализует trayBackend.
func (b *linuxTrayBackend) Quit() {
	select {
	case <-b.done:
		return
	default:
	}
	_ = b.conn.Close()
	<-b.done
}

// sniObject реализует методы org.kde.StatusNotifierItem.
type sniObject struct{}

func (o *sniObject) Activate(x, y int32) *dbus.Error            { return nil }
func (o *sniObject) SecondaryActivate(x, y int32) *dbus.Error   { return nil }
func (o *sniObject) XAyatanaSecondaryActivate(ts uint32) *dbus.Error { return nil }
func (o *sniObject) ContextMenu(x, y int32) *dbus.Error         { return nil }
func (o *sniObject) Scroll(delta int32, orientation string) *dbus.Error {
	return nil
}

func exportIntrospectable(conn *dbus.Conn) error {
	node := &introspect.Node{
		Name: "/StatusNotifierItem",
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name: sniIface,
				Methods: []introspect.Method{
					{Name: "Activate", Args: []introspect.Arg{{Name: "x", Type: "i", Direction: "in"}, {Name: "y", Type: "i", Direction: "in"}}},
					{Name: "SecondaryActivate", Args: []introspect.Arg{{Name: "x", Type: "i", Direction: "in"}, {Name: "y", Type: "i", Direction: "in"}}},
					{Name: "XAyatanaSecondaryActivate", Args: []introspect.Arg{{Name: "timestamp", Type: "u", Direction: "in"}}},
					{Name: "ContextMenu", Args: []introspect.Arg{{Name: "x", Type: "i", Direction: "in"}, {Name: "y", Type: "i", Direction: "in"}}},
					{Name: "Scroll", Args: []introspect.Arg{{Name: "delta", Type: "i", Direction: "in"}, {Name: "orientation", Type: "s", Direction: "in"}}},
				},
			},
		},
	}
	return conn.Export(introspect.NewIntrospectable(node), sniPath, "org.freedesktop.DBus.Introspectable")
}