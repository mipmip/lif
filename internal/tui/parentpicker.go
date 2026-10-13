package tui

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xRiErOS/lif/internal/ui"
	"github.com/xRiErOS/lif/pkg/candidates"
	"github.com/xRiErOS/lif/pkg/config"
	"github.com/xRiErOS/lif/pkg/item"
	"github.com/xRiErOS/lif/pkg/itemgraph"
)

// parentSelectedMsg is sent when a parent is selected from the picker
type parentSelectedMsg struct {
	itemIDs  []string // the items being modified
	parentID string   // the new parent ID (empty string to clear parent)
}

// closeParentPickerMsg is sent when the parent picker is cancelled
type closeParentPickerMsg struct{}

// parentItem wraps an item to implement list.Item for the parent picker
type parentItem struct {
	item *item.Item
	cfg  *config.Config
}

func (i parentItem) Title() string       { return i.item.Title }
func (i parentItem) Description() string { return i.item.ID }
func (i parentItem) FilterValue() string { return i.item.Title + " " + i.item.ID }

// clearParentItem is a special item to clear the parent
type clearParentItem struct{}

func (i clearParentItem) Title() string       { return "(No Parent)" }
func (i clearParentItem) Description() string { return "Clear the parent assignment" }
func (i clearParentItem) FilterValue() string { return "no parent clear none" }

// parentItemDelegate handles rendering of parent picker items
type parentItemDelegate struct {
	cfg *config.Config
}

func (d parentItemDelegate) Height() int                             { return 1 }
func (d parentItemDelegate) Spacing() int                            { return 0 }
func (d parentItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d parentItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	var cursor string
	if index == m.Index() {
		cursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true).Render("▌") + " "
	} else {
		cursor = "  "
	}

	switch item := listItem.(type) {
	case clearParentItem:
		text := ui.Muted.Render(item.Title())
		fmt.Fprint(w, cursor+text)

	case parentItem:
		// Get colors from config
		colors := d.cfg.GetItemColors(item.item.Status, item.item.Type, item.item.Priority)

		// Format: [type] title (id)
		typeBadge := ui.RenderTypeText(item.item.Type, colors.TypeColor)
		title := item.item.Title
		if colors.IsArchive {
			title = ui.Muted.Render(title)
		}
		id := ui.Muted.Render(" (" + item.item.ID + ")")

		fmt.Fprint(w, cursor+typeBadge+" "+title+id)
	}
}

// parentPickerModel is the model for the parent picker view
type parentPickerModel struct {
	list          list.Model
	itemIDs       []string // the items we're setting the parent for
	itemTitle     string   // display title (single title or "N selected items")
	itemTypes     []string // types of the items (to filter eligible parents)
	currentParent string   // current parent ID (to highlight, only for single item)
	width         int
	height        int
}

func newParentPickerModel(itemIDs []string, itemTitle string, itemTypes []string, currentParent string, resolver *itemgraph.CoreResolver, cfg *config.Config, width, height int) parentPickerModel {
	eligibleItems, _ := candidates.ParentCandidates(context.Background(), resolver, cfg, itemIDs, itemTypes)

	delegate := parentItemDelegate{cfg: cfg}

	// Build items list - start with "clear parent" option
	items := make([]list.Item, 0, len(eligibleItems)+1)
	items = append(items, clearParentItem{})

	selectedIndex := 0 // default to "No Parent"
	for i, b := range eligibleItems {
		items = append(items, parentItem{item: b, cfg: cfg})
		// If this is the current parent, remember its index (+1 for the clear option)
		if b.ID == currentParent {
			selectedIndex = i + 1
		}
	}

	// Calculate modal dimensions (matching View() function)
	modalWidth := max(40, min(80, width*60/100))
	modalHeight := max(10, min(20, height*60/100))
	// List dimensions within modal (account for border, padding, subtitle, help)
	listWidth := modalWidth - 6  // border (2) + padding (4)
	listHeight := modalHeight - 7 // border (2) + subtitle (1) + help (1) + padding (3)

	l := list.New(items, delegate, listWidth, listHeight)
	l.Title = "Select Parent"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	l.Styles.Title = listTitleStyle
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 0, 0)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary)

	// Select the current parent if set
	if selectedIndex > 0 && selectedIndex < len(items) {
		l.Select(selectedIndex)
	}

	return parentPickerModel{
		list:          l,
		itemIDs:       itemIDs,
		itemTitle:     itemTitle,
		itemTypes:     itemTypes,
		currentParent: currentParent,
		width:         width,
		height:        height,
	}
}

func (m parentPickerModel) Init() tea.Cmd {
	return nil
}

func (m parentPickerModel) Update(msg tea.Msg) (parentPickerModel, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Recalculate modal dimensions
		modalWidth := max(40, min(80, msg.Width*60/100))
		modalHeight := max(10, min(20, msg.Height*60/100))
		listWidth := modalWidth - 6
		listHeight := modalHeight - 7
		m.list.SetSize(listWidth, listHeight)

	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case "enter":
				switch item := m.list.SelectedItem().(type) {
				case clearParentItem:
					return m, func() tea.Msg {
						return parentSelectedMsg{itemIDs: m.itemIDs, parentID: ""}
					}
				case parentItem:
					return m, func() tea.Msg {
						return parentSelectedMsg{itemIDs: m.itemIDs, parentID: item.item.ID}
					}
				}
			case "esc", "backspace":
				// Return without selecting
				return m, func() tea.Msg {
					return closeParentPickerMsg{}
				}
			}
		}
	}

	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m parentPickerModel) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	// For multi-select, don't show individual item ID
	var itemID string
	if len(m.itemIDs) == 1 {
		itemID = m.itemIDs[0]
	}

	return renderPickerModal(pickerModalConfig{
		Title:       "Select Parent",
		ItemTitle:   m.itemTitle,
		ItemID:      itemID,
		ListContent: m.list.View(),
		Width:       m.width,
		WidthPct:    60,
		MaxWidth:    80,
	})
}

// ModalView returns the picker rendered as a centered modal overlay on top of the background
func (m parentPickerModel) ModalView(bgView string, fullWidth, fullHeight int) string {
	modal := m.View()
	return overlayModal(bgView, modal, fullWidth, fullHeight)
}
