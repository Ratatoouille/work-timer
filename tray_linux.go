//go:build linux

package main

import (
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const (
	sniIface    = "org.kde.StatusNotifierItem"
	sniPath     = dbus.ObjectPath("/StatusNotifierItem")
	watcherDest = "org.kde.StatusNotifierWatcher"
	watcherPath = dbus.ObjectPath("/StatusNotifierWatcher")

	// trayLockName — well-known имя на session bus, которое захватывает экземпляр,
	// владеющий треем. Второй экземпляр не сможет его получить и не будет
	// добавлять свою метку в трей.
	trayLockName = "dev.ratatoouille.work-timer"
)

// linuxTrayBackend реализует trayBackend через собственный DBus
// StatusNotifierItem. В отличие от fyne.io/systray он выставляет свойство
// XAyatanaLabel, которое GNOME (ubuntu-appindicators) рендерит как нативную
// текстовую метку рядом с иконкой — чёткий текст вместо пиксельного.
type linuxTrayBackend struct {
	loc     Locale
	ready   chan struct{}
	stateCh chan trayRender
	quit    chan struct{} // закрывается Quit — просьба остановить цикл
	stopped chan struct{} // закрывается run по завершении
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

// startTray запускает трей. Возвращает (nil, false), если трей недоступен
// (приложение продолжает работать без него), и (nil, true), если трей уже
// занят другим экземпляром work-timer.
func startTray(cfg Config, loc Locale) (*TrayManager, bool) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, false
	}

	// Проверяем, не зарегистрирован ли уже SNI-элемент work-timer (ловит в том
	// числе экземпляры, запущенные до появления блокировки по имени).
	if hasExistingTray(conn) {
		_ = conn.Close()
		return nil, true
	}

	// Захватываем well-known имя — атомарная защита от одновременного старта
	// двух экземпляров.
	reply, err := conn.RequestName(trayLockName, dbus.NameFlagDoNotQueue)
	if err != nil || reply != dbus.RequestNameReplyPrimaryOwner {
		_ = conn.Close()
		return nil, true
	}

	backend := &linuxTrayBackend{
		loc:     loc,
		ready:   make(chan struct{}),
		stateCh: make(chan trayRender, 8),
		quit:    make(chan struct{}),
		stopped: make(chan struct{}),
		conn:    conn,
	}

	go backend.run()

	return NewTrayManager(backend, loc), false
}

// hasExistingTray проверяет, зарегистрирован ли в StatusNotifierWatcher элемент
// с Id "work-timer". Элементы приходят в виде "servicename@/StatusNotifierItem";
// путь по умолчанию — /StatusNotifierItem.
func hasExistingTray(conn *dbus.Conn) bool {
	watcher := conn.Object(watcherDest, watcherPath)
	v, err := watcher.GetProperty("org.kde.StatusNotifierWatcher.RegisteredStatusNotifierItems")
	if err != nil {
		return false
	}
	items, ok := v.Value().([]string)
	if !ok {
		return false
	}

	for _, item := range items {
		service, path := item, string(sniPath)
		if at := strings.IndexByte(item, '@'); at >= 0 {
			service = item[:at]
			if p := item[at+1:]; p != "" {
				path = p
			}
		}
		idv, err := conn.Object(service, dbus.ObjectPath(path)).
			GetProperty("org.kde.StatusNotifierItem.Id")
		if err != nil {
			continue
		}
		if id, ok := idv.Value().(string); ok && id == "work-timer" {
			return true
		}
	}
	return false
}

// transparentPixmap — прозрачная иконка 1x1 в формате ARGB (как ждёт SNI).
func transparentPixmap() []PX {
	return []PX{{W: 1, H: 1, Pix: []byte{0, 0, 0, 0}}}
}

func (b *linuxTrayBackend) run() {
	defer close(b.stopped)

	conn := b.conn
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
		case <-b.quit:
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
	case <-b.stopped:
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
	case <-b.stopped:
		return
	default:
	}
	close(b.quit)
	<-b.stopped
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