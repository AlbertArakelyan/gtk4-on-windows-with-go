package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const appID = "com.example.gotk4todo"

type filter int

const (
	filterAll filter = iota
	filterActive
	filterDone
)

type todo struct {
	text  string
	done  bool
	row   *gtk.ListBoxRow
	label *gtk.Label
}

type todoApp struct {
	window  *gtk.ApplicationWindow
	entry   *gtk.Entry
	list    *gtk.ListBox
	counter *gtk.Label
	todos   []*todo
	filter  filter
}

func main() {
	app := gtk.NewApplication(appID, gio.ApplicationFlagsNone)
	app.ConnectActivate(func() { newTodoApp(app).window.Present() })

	if code := app.Run(os.Args); code > 0 {
		os.Exit(code)
	}
}

func newTodoApp(app *gtk.Application) *todoApp {
	t := &todoApp{}

	t.entry = gtk.NewEntry()
	t.entry.SetPlaceholderText("What needs to be done?")
	t.entry.SetHExpand(true)
	t.entry.ConnectActivate(t.addFromEntry)

	addButton := gtk.NewButtonWithLabel("Add")
	addButton.AddCSSClass("suggested-action")
	addButton.ConnectClicked(t.addFromEntry)

	inputBox := gtk.NewBox(gtk.OrientationHorizontal, 6)
	inputBox.Append(t.entry)
	inputBox.Append(addButton)

	t.list = gtk.NewListBox()
	t.list.SetSelectionMode(gtk.SelectionNone)
	t.list.AddCSSClass("boxed-list")
	t.list.SetPlaceholder(gtk.NewLabel("Nothing to show"))

	scroller := gtk.NewScrolledWindow()
	scroller.SetPolicy(gtk.PolicyNever, gtk.PolicyAutomatic)
	scroller.SetVExpand(true)
	scroller.SetChild(t.list)

	t.counter = gtk.NewLabel("")
	t.counter.SetXAlign(0)
	t.counter.SetHExpand(true)
	t.counter.AddCSSClass("dim-label")

	clearButton := gtk.NewButtonWithLabel("Clear completed")
	clearButton.ConnectClicked(t.clearCompleted)

	footer := gtk.NewBox(gtk.OrientationHorizontal, 6)
	footer.Append(t.counter)
	footer.Append(t.filterButtons())
	footer.Append(clearButton)

	content := gtk.NewBox(gtk.OrientationVertical, 12)
	content.SetMarginTop(12)
	content.SetMarginBottom(12)
	content.SetMarginStart(12)
	content.SetMarginEnd(12)
	content.Append(inputBox)
	content.Append(scroller)
	content.Append(footer)

	t.window = gtk.NewApplicationWindow(app)
	t.window.SetTitle("Todo")
	t.window.SetDefaultSize(520, 600)
	t.window.SetChild(content)

	t.refresh()
	t.entry.GrabFocus()
	return t
}

// filterButtons builds a linked group of radio-like toggle buttons.
func (t *todoApp) filterButtons() *gtk.Box {
	box := gtk.NewBox(gtk.OrientationHorizontal, 0)
	box.AddCSSClass("linked")

	var first *gtk.ToggleButton
	for _, f := range []struct {
		label string
		value filter
	}{{"All", filterAll}, {"Active", filterActive}, {"Done", filterDone}} {
		f := f
		btn := gtk.NewToggleButtonWithLabel(f.label)
		if first == nil {
			first = btn
			btn.SetActive(true)
		} else {
			btn.SetGroup(first)
		}
		btn.ConnectToggled(func() {
			if btn.Active() {
				t.filter = f.value
				t.refresh()
			}
		})
		box.Append(btn)
	}
	return box
}

func (t *todoApp) addFromEntry() {
	text := strings.TrimSpace(t.entry.Text())
	if text == "" {
		return
	}
	t.entry.SetText("")
	t.add(text)
}

func (t *todoApp) add(text string) {
	item := &todo{text: text}

	check := gtk.NewCheckButton()
	check.ConnectToggled(func() {
		item.done = check.Active()
		t.refresh()
	})

	item.label = gtk.NewLabel("")
	item.label.SetXAlign(0)
	item.label.SetHExpand(true)
	item.label.SetWrap(true)

	deleteButton := gtk.NewButtonFromIconName("user-trash-symbolic")
	deleteButton.SetTooltipText("Delete")
	deleteButton.AddCSSClass("flat")
	deleteButton.ConnectClicked(func() { t.remove(item) })

	box := gtk.NewBox(gtk.OrientationHorizontal, 8)
	box.SetMarginTop(6)
	box.SetMarginBottom(6)
	box.SetMarginStart(6)
	box.SetMarginEnd(6)
	box.Append(check)
	box.Append(item.label)
	box.Append(deleteButton)

	item.row = gtk.NewListBoxRow()
	item.row.SetChild(box)

	t.list.Append(item.row)
	t.todos = append(t.todos, item)
	t.refresh()
}

func (t *todoApp) remove(item *todo) {
	for i, it := range t.todos {
		if it == item {
			t.todos = append(t.todos[:i], t.todos[i+1:]...)
			break
		}
	}
	t.list.Remove(item.row)
	t.refresh()
}

func (t *todoApp) clearCompleted() {
	kept := t.todos[:0]
	for _, it := range t.todos {
		if it.done {
			t.list.Remove(it.row)
		} else {
			kept = append(kept, it)
		}
	}
	t.todos = kept
	t.refresh()
}

// refresh syncs labels, row visibility and the counter with the model.
func (t *todoApp) refresh() {
	left := 0
	for _, it := range t.todos {
		escaped := glib.MarkupEscapeText(it.text)
		if it.done {
			it.label.SetMarkup("<s>" + escaped + "</s>")
			it.label.AddCSSClass("dim-label")
		} else {
			it.label.SetMarkup(escaped)
			it.label.RemoveCSSClass("dim-label")
			left++
		}

		switch t.filter {
		case filterActive:
			it.row.SetVisible(!it.done)
		case filterDone:
			it.row.SetVisible(it.done)
		default:
			it.row.SetVisible(true)
		}
	}

	noun := "items"
	if left == 1 {
		noun = "item"
	}
	t.counter.SetText(fmt.Sprintf("%d %s left", left, noun))
}
