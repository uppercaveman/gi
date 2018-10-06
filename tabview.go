// Copyright (c) 2018, The GoKi Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gi

import (
	"log"
	"reflect"

	"github.com/goki/gi/units"
	"github.com/goki/ki"
	"github.com/goki/ki/kit"
)

// TabView switches among child widgets via tabs.  The selected widget gets
// the full allocated space avail after the tabs are accounted for.  The
// TabView is just a Vertical layout that manages two child widgets: a
// HorizFlow Layout for the tabs (which can flow across multiple rows as
// needed) and a Stacked Frame that actually contains all the children, and
// provides scrollbars as needed to any content within.  Typically should have
// max stretch and a set preferred size, so it expands.
type TabView struct {
	Layout
	MaxChars     int          `desc:"maximum number of characters to include in tab label -- elides labels that are longer than that"`
	TabViewSig   ki.Signal    `json:"-" xml:"-" desc:"signal for tab widget -- see TabViewSignals for the types"`
	NewTabButton bool         `desc:"show a new tab button at right of list of tabs"`
	NewTabType   reflect.Type `desc:"type of widget to create in a new tab via new tab button -- Frame by default"`
}

var KiT_TabView = kit.Types.AddType(&TabView{}, TabViewProps)

var TabViewProps = ki.Props{
	"border-color":     &Prefs.Colors.Border,
	"border-width":     units.NewValue(2, units.Px),
	"background-color": &Prefs.Colors.Background,
	"color":            &Prefs.Colors.Font,
	"max-width":        -1,
	"max-height":       -1,
	"width":            units.NewValue(10, units.Em),
	"height":           units.NewValue(10, units.Em),
}

// NTabs returns number of tabs
func (tv *TabView) NTabs() int {
	fr := tv.Frame()
	if fr == nil {
		return 0
	}
	return len(fr.Kids)
}

// CurTab returns currently-selected tab, and its index -- returns false none
func (tv *TabView) CurTab() (Node2D, int, bool) {
	if tv.NTabs() == 0 {
		return nil, -1, false
	}
	fr := tv.Frame()
	if fr.StackTop < 0 {
		return nil, -1, false
	}
	widg := fr.KnownChild(fr.StackTop).(Node2D)
	return widg, fr.StackTop, true
}

// AddTab adds a widget as a new tab, with given tab label, and returns the
// index of that tab
func (tv *TabView) AddTab(widg Node2D, label string) int {
	fr := tv.Frame()
	idx := len(*fr.Children())
	tv.InsertTab(widg, label, idx)
	return idx
}

// InsertTabOnlyAt inserts just the tab at given index -- after panel has
// already been added to frame -- assumed to be wrapped in update.  Generally
// for internal use.
func (tv *TabView) InsertTabOnlyAt(widg Node2D, label string, idx int) {
	tb := tv.Tabs()
	tab := tb.InsertNewChild(KiT_TabButton, idx, label).(*TabButton)
	tab.Data = idx
	tab.Tooltip = label
	tab.SetText(label)
	tab.ActionSig.ConnectOnly(tv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
		tvv := recv.Embed(KiT_TabView).(*TabView)
		act := send.Embed(KiT_TabButton).(*TabButton)
		tabIdx := act.Data.(int)
		tvv.SelectTabIndex(tabIdx)
	})
	fr := tv.Frame()
	if len(fr.Kids) == 1 {
		fr.StackTop = 0
		tab.SetSelectedState(true)
	} else {
		widg.AsNode2D().SetInvisibleTree() // new tab is invisible until selected
	}
}

// InsertTab inserts a widget into given index position within list of tabs
func (tv *TabView) InsertTab(widg Node2D, label string, idx int) {
	fr := tv.Frame()
	updt := tv.UpdateStart()
	tv.SetFullReRender()
	fr.InsertChild(widg, idx)
	tv.InsertTabOnlyAt(widg, label, idx)
	tv.UpdateEnd(updt)
}

// AddNewTab adds a new widget as a new tab of given widget type, with given
// tab label, and returns the new widget and its tab index
func (tv *TabView) AddNewTab(typ reflect.Type, label string) (Node2D, int) {
	fr := tv.Frame()
	idx := len(*fr.Children())
	widg := tv.InsertNewTab(typ, label, idx)
	return widg, idx
}

// AddNewTabAction adds a new widget as a new tab of given widget type, with given
// tab label, and returns the new widget and its tab index -- emits TabAdded signal
func (tv *TabView) AddNewTabAction(typ reflect.Type, label string) (Node2D, int) {
	widg, idx := tv.AddNewTab(typ, label)
	tv.TabViewSig.Emit(tv.This, int64(TabAdded), idx)
	return widg, idx
}

// InsertNewTab inserts a new widget of given type into given index position
// within list of tabs, and returns that new widget
func (tv *TabView) InsertNewTab(typ reflect.Type, label string, idx int) Node2D {
	fr := tv.Frame()
	updt := tv.UpdateStart()
	tv.SetFullReRender()
	widg := fr.InsertNewChild(typ, idx, label).(Node2D)
	tv.InsertTabOnlyAt(widg, label, idx)
	tv.UpdateEnd(updt)
	return widg
}

// TabAtIndex returns content widget and tab button at given index, false if
// index out of range (emits log message)
func (tv *TabView) TabAtIndex(idx int) (Node2D, *TabButton, bool) {
	fr := tv.Frame()
	tb := tv.Tabs()
	sz := len(*fr.Children())
	if idx < 0 || idx >= sz {
		log.Printf("giv.TabView: index %v out of range for number of tabs: %v\n", idx, sz)
		return nil, nil, false
	}
	tab := tb.KnownChild(idx).Embed(KiT_TabButton).(*TabButton)
	widg := fr.KnownChild(idx).(Node2D)
	return widg, tab, true
}

// SelectTabIndex selects tab at given index, returning it -- returns false if
// index is invalid
func (tv *TabView) SelectTabIndex(idx int) (Node2D, bool) {
	widg, tab, ok := tv.TabAtIndex(idx)
	if !ok {
		return nil, false
	}
	fr := tv.Frame()
	if fr.StackTop == idx {
		return widg, true
	}
	updt := tv.UpdateStart()
	tv.UnselectOtherTabs(idx)
	tab.SetSelectedState(true)
	fr.StackTop = idx
	// frame  / layout will set invisible etc
	tv.UpdateEnd(updt)
	return widg, true
}

// SelectTabIndexAction selects tab at given index and emits selected signal,
// with the index of the selected tab -- this is what is called when a tab is
// clicked
func (tv *TabView) SelectTabIndexAction(idx int) {
	_, ok := tv.SelectTabIndex(idx)
	if ok {
		tv.TabViewSig.Emit(tv.This, int64(TabSelected), idx)
	}
}

// TabByName returns tab with given name, and its index -- returns false if
// not found
func (tv *TabView) TabByName(label string) (Node2D, int, bool) {
	tb := tv.Tabs()
	idx, ok := tb.Children().IndexByName(label, 0)
	if !ok {
		return nil, -1, false
	}
	fr := tv.Frame()
	widg := fr.KnownChild(idx).(Node2D)
	return widg, idx, true
}

// SelectTabName selects tab by name, returning it -- returns false if not
// found
func (tv *TabView) SelectTabByName(label string) (Node2D, int, bool) {
	widg, idx, ok := tv.TabByName(label)
	if ok {
		tv.SelectTabIndex(idx)
	}
	return widg, idx, ok
}

// DeleteTabIndex deletes tab at given index, optionally calling destroy on
// tab contents -- returns widget if destroy == false and bool success
func (tv *TabView) DeleteTabIndex(idx int, destroy bool) (Node2D, bool) {
	widg, _, ok := tv.TabAtIndex(idx)
	if !ok {
		return nil, false
	}
	fr := tv.Frame()
	sz := len(*fr.Children())
	tb := tv.Tabs()
	updt := tv.UpdateStart()
	tv.SetFullReRender()
	nxtidx := -1
	if fr.StackTop == idx {
		if idx > 0 {
			nxtidx = idx - 1
		} else if idx < sz-1 {
			nxtidx = idx
		}
	}
	fr.DeleteChildAtIndex(idx, destroy)
	tb.DeleteChildAtIndex(idx, true) // always destroy -- we manage
	tv.RenumberTabs()
	if nxtidx >= 0 {
		tv.SelectTabIndex(nxtidx)
	}
	tv.UpdateEnd(updt)
	if destroy {
		return nil, true
	} else {
		return widg, true
	}
}

// DeleteTabIndexAction deletes tab at given index using destroy flag, and
// emits TabDeleted signal -- this is called by the delete button on the tab
func (tv *TabView) DeleteTabIndexAction(idx int) {
	_, ok := tv.DeleteTabIndex(idx, true)
	if ok {
		tv.TabViewSig.Emit(tv.This, int64(TabDeleted), idx)
	}
}

// ConfigNewTabButton configures the new tab + button at end of list of tabs
func (tv *TabView) ConfigNewTabButton() bool {
	sz := tv.NTabs()
	tb := tv.Tabs()
	ntb := len(tb.Kids)
	if tv.NewTabButton {
		if ntb == sz+1 {
			return false
		}
		if tv.NewTabType == nil {
			tv.NewTabType = KiT_Frame
		}
		tab := tb.InsertNewChild(KiT_Action, ntb, "new-tab").(*Action)
		tab.Data = -1
		tab.SetIcon("plus")
		tab.ActionSig.ConnectOnly(tv.This, func(recv, send ki.Ki, sig int64, data interface{}) {
			tvv := recv.Embed(KiT_TabView).(*TabView)
			tvv.SetFullReRender()
			tvv.AddNewTabAction(tvv.NewTabType, "New Tab")
		})
		return true
	} else {
		if ntb == sz {
			return false
		}
		tb.DeleteChildAtIndex(ntb-1, true) // always destroy -- we manage
		return true
	}
}

// TabViewSignals are signals that the TabView can send
type TabViewSignals int64

const (
	// TabSelected indicates tab was selected -- data is the tab index
	TabSelected TabViewSignals = iota

	// TabAdded indicates tab was added -- data is the tab index
	TabAdded

	// TabDeleted indicates tab was deleted -- data is the tab index
	TabDeleted

	TabViewSignalsN
)

//go:generate stringer -type=TabViewSignals

// InitTabView initializes the tab widget children if it hasn't been done yet
func (tv *TabView) InitTabView() {
	if len(tv.Kids) != 0 {
		return
	}
	if tv.Sty.Font.Size.Val == 0 { // not yet styled
		tv.StyleLayout()
	}
	updt := tv.UpdateStart()
	tv.Lay = LayoutVert
	tv.SetReRenderAnchor()

	tabs := tv.AddNewChild(KiT_Frame, "tabs").(*Frame)
	tabs.Lay = LayoutHoriz
	tabs.SetStretchMaxWidth()
	// tabs.SetStretchMaxHeight()
	// tabs.SetMinPrefWidth(units.NewValue(10, units.Em))
	tabs.SetProp("height", units.NewValue(1.8, units.Em))
	tabs.SetProp("overflow", "hidden") // no scrollbars!
	tabs.SetProp("padding", units.NewValue(0, units.Px))
	tabs.SetProp("margin", units.NewValue(0, units.Px))
	tabs.SetProp("spacing", units.NewValue(4, units.Px))
	tabs.SetProp("background-color", "linear-gradient(pref(Control), highlight-10)")

	frame := tv.AddNewChild(KiT_Frame, "frame").(*Frame)
	frame.Lay = LayoutStacked
	frame.SetMinPrefWidth(units.NewValue(10, units.Em))
	frame.SetMinPrefHeight(units.NewValue(7, units.Em))
	frame.SetStretchMaxWidth()
	frame.SetStretchMaxHeight()

	tv.ConfigNewTabButton()

	tv.UpdateEnd(updt)
}

// Tabs returns the layout containing the tabs -- the first element within us
func (tv *TabView) Tabs() *Frame {
	tv.InitTabView()
	return tv.KnownChild(0).(*Frame)
}

// Frame returns the stacked frame layout -- the second element
func (tv *TabView) Frame() *Frame {
	tv.InitTabView()
	return tv.KnownChild(1).(*Frame)
}

// UnselectOtherTabs turns off all the tabs except given one
func (tv *TabView) UnselectOtherTabs(idx int) {
	sz := tv.NTabs()
	tbs := tv.Tabs()
	for i := 0; i < sz; i++ {
		if i == idx {
			continue
		}
		tb := tbs.KnownChild(i).Embed(KiT_TabButton).(*TabButton)
		if tb.IsSelected() {
			tb.SetSelectedState(false)
		}
	}
}

// RenumberTabs assigns proper index numbers to each tab
func (tv *TabView) RenumberTabs() {
	sz := tv.NTabs()
	tbs := tv.Tabs()
	for i := 0; i < sz; i++ {
		tb := tbs.KnownChild(i).Embed(KiT_TabButton).(*TabButton)
		tb.Data = i
	}
}

func (tv *TabView) Style2D() {
	tv.InitTabView()
	tv.Layout.Style2D()
}

// RenderTabSeps renders the separators between tabs
func (tv *TabView) RenderTabSeps() {
	rs := &tv.Viewport.Render
	pc := &rs.Paint
	st := &tv.Sty
	pc.StrokeStyle.Width = st.Border.Width
	pc.StrokeStyle.SetColor(&st.Border.Color)
	bw := st.Border.Width.Dots

	tbs := tv.Tabs()
	sz := len(tbs.Kids)
	for i := 1; i < sz; i++ {
		tb := tbs.KnownChild(i).(Node2D)
		ni := tb.AsWidget()

		pos := ni.LayData.AllocPos
		sz := ni.LayData.AllocSize.AddVal(-2.0 * st.Layout.Margin.Dots)
		pc.DrawLine(rs, pos.X-bw, pos.Y, pos.X-bw, pos.Y+sz.Y)
	}
	pc.FillStrokeClear(rs)
}

func (tv *TabView) Render2D() {
	if tv.FullReRenderIfNeeded() {
		return
	}
	if tv.PushBounds() {
		tv.This.(Node2D).ConnectEvents2D()
		tv.RenderScrolls()
		tv.Render2DChildren()
		tv.RenderTabSeps()
		tv.PopBounds()
	} else {
		tv.DisconnectAllEvents(AllPris) // uses both Low and Hi
	}
}

////////////////////////////////////////////////////////////////////////////////////////
// TabButton

// TabButton is a larger select action and a small close action. Indicator
// icon is used for close icon.
type TabButton struct {
	Action
}

var KiT_TabButton = kit.Types.AddType(&TabButton{}, TabButtonProps)

// TabButtonMinWidth is the minimum width of the tab button, in Ch units
var TabButtonMinWidth = float32(8)

var TabButtonProps = ki.Props{
	"min-width":        units.NewValue(TabButtonMinWidth, units.Ch),
	"min-height":       units.NewValue(1.6, units.Em),
	"border-width":     units.NewValue(0, units.Px),
	"border-radius":    units.NewValue(0, units.Px),
	"border-color":     &Prefs.Colors.Border,
	"border-style":     BorderSolid,
	"box-shadow.color": &Prefs.Colors.Shadow,
	"text-align":       AlignCenter,
	"background-color": &Prefs.Colors.Control,
	"color":            &Prefs.Colors.Font,
	"padding":          units.NewValue(4, units.Px), // we go to edge of bar
	"margin":           units.NewValue(0, units.Px),
	"indicator":        "close",
	"#icon": ki.Props{
		"width":   units.NewValue(1, units.Em),
		"height":  units.NewValue(1, units.Em),
		"margin":  units.NewValue(0, units.Px),
		"padding": units.NewValue(0, units.Px),
		"fill":    &Prefs.Colors.Icon,
		"stroke":  &Prefs.Colors.Font,
	},
	"#label": ki.Props{
		"margin":  units.NewValue(0, units.Px),
		"padding": units.NewValue(0, units.Px),
	},
	"#close-stretch": ki.Props{
		"width": units.NewValue(1, units.Ch),
	},
	"#close": ki.Props{
		"width":          units.NewValue(.5, units.Ex),
		"height":         units.NewValue(.5, units.Ex),
		"margin":         units.NewValue(0, units.Px),
		"padding":        units.NewValue(0, units.Px),
		"vertical-align": AlignBottom,
	},
	"#shortcut": ki.Props{
		"margin":  units.NewValue(0, units.Px),
		"padding": units.NewValue(0, units.Px),
	},
	"#sc-stretch": ki.Props{
		"min-width": units.NewValue(2, units.Em),
	},
	ButtonSelectors[ButtonActive]: ki.Props{
		"background-color": "linear-gradient(lighter-0, highlight-10)",
	},
	ButtonSelectors[ButtonInactive]: ki.Props{
		"border-color": "lighter-50",
		"color":        "lighter-50",
	},
	ButtonSelectors[ButtonHover]: ki.Props{
		"background-color": "linear-gradient(highlight-10, highlight-10)",
	},
	ButtonSelectors[ButtonFocus]: ki.Props{
		"border-width":     units.NewValue(2, units.Px),
		"background-color": "linear-gradient(samelight-50, highlight-10)",
	},
	ButtonSelectors[ButtonDown]: ki.Props{
		"color":            "lighter-90",
		"background-color": "linear-gradient(highlight-30, highlight-10)",
	},
	ButtonSelectors[ButtonSelected]: ki.Props{
		"background-color": "linear-gradient(pref(Select), highlight-10)",
	},
}

func (tb *TabButton) ButtonAsBase() *ButtonBase {
	return &(tb.ButtonBase)
}

func (tb *TabButton) TabView() *TabView {
	tv, ok := tb.ParentByType(KiT_TabView, true)
	if !ok {
		return nil
	}
	return tv.Embed(KiT_TabView).(*TabView)
}

func (tb *TabButton) ConfigParts() {
	config := kit.TypeAndNameList{}
	clsIdx := 0
	config.Add(KiT_Action, "close")
	config.Add(KiT_Stretch, "close-stretch")
	icIdx, lbIdx := tb.ConfigPartsIconLabel(&config, string(tb.Icon), tb.Text)
	mods, updt := tb.Parts.ConfigChildren(config, false) // not unique names
	tb.ConfigPartsSetIconLabel(string(tb.Icon), tb.Text, icIdx, lbIdx)
	if mods {
		cls := tb.Parts.KnownChild(clsIdx).(*Action)
		if tb.Indicator.IsNil() {
			tb.Indicator = "close"
		}
		tb.StylePart(Node2D(cls))

		icnm := string(tb.Indicator)
		cls.SetIcon(icnm)
		cls.SetProp("no-focus", true)
		cls.ActionSig.ConnectOnly(tb.This, func(recv, send ki.Ki, sig int64, data interface{}) {
			tbb := recv.Embed(KiT_TabButton).(*TabButton)
			tabIdx := tbb.Data.(int)
			tvv := tb.TabView()
			if tvv != nil {
				tvv.DeleteTabIndexAction(tabIdx)
			}
		})
		tb.UpdateEnd(updt)
	}
}

func (tb *TabButton) Size2D(iter int) {
	ppref := tb.Parts.LayData.Size.Pref // get from parts
	spc := tb.Sty.BoxSpace()
	tb.SetProp("width", units.NewValue(ppref.X+2*spc, units.Dot))
	tb.SetProp("height", units.NewValue(ppref.Y+2*spc, units.Dot))
	tb.InitLayout2D() // sets from props
}