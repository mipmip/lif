package tui

import (
	"context"
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/xRiErOS/lif/internal/ui"
	"github.com/xRiErOS/lif/pkg/config"
	"github.com/xRiErOS/lif/pkg/item"
	"github.com/xRiErOS/lif/pkg/itemgraph"
	"github.com/xRiErOS/lif/pkg/itemgraph/model"
)

// itemItem wraps a Item to implement list.Item, with tree context
type itemRow struct {
	item           *item.Item
	cfg            *config.Config
	treePrefix     string // tree prefix for rendering (e.g., "├─" or "  └─")
	matched        bool   // true if item matched filter (vs. ancestor shown for context)
	implicitStatus string // implicit terminal status from an ancestor, if any
}

func (i itemRow) Title() string       { return i.item.Title }
func (i itemRow) Description() string { return i.item.ID + " · " + i.item.Status }
func (i itemRow) FilterValue() string { return i.item.Title + " " + i.item.ID }

// itemDelegate handles rendering of list items
type itemDelegate struct {
	cfg           *config.Config
	hasTags       bool
	width         int
	cols          ui.ResponsiveColumns // cached responsive columns
	idColWidth    int                  // ID column width (accounts for tree prefix)
	selectedItems *map[string]bool     // pointer to marked items for multi-select
}

func newItemDelegate(cfg *config.Config) itemDelegate {
	return itemDelegate{cfg: cfg, hasTags: false, width: 0}
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	item, ok := listItem.(itemRow)
	if !ok {
		return
	}

	// Get colors from config
	colors := d.cfg.GetItemColors(item.item.Status, item.item.Type, item.item.Priority)

	// Calculate max title width using responsive columns
	idWidth := d.cols.ID
	if d.idColWidth > 0 {
		idWidth = d.idColWidth
	}
	baseWidth := idWidth + d.cols.Status + d.cols.Type + 4 // 4 for cursor + padding
	if d.cols.ShowTags {
		baseWidth += d.cols.Tags
	}
	maxTitleWidth := max(0, m.Width()-baseWidth)

	// Check if item is marked for multi-select
	var isMarked bool
	if d.selectedItems != nil {
		isMarked = (*d.selectedItems)[item.item.ID]
	}

	str := ui.RenderItemRow(
		item.item.ID,
		item.item.Status,
		item.item.Type,
		item.item.Title,
		ui.ItemRowConfig{
			StatusColor:    colors.StatusColor,
			TypeColor:      colors.TypeColor,
			PriorityColor:  colors.PriorityColor,
			Priority:       item.item.Priority,
			IsArchive:      colors.IsArchive,
			MaxTitleWidth:  maxTitleWidth,
			ShowCursor:     true,
			IsSelected:     index == m.Index(),
			IsMarked:       isMarked,
			Tags:           item.item.Tags,
			ShowTags:       d.cols.ShowTags,
			TagsColWidth:   d.cols.Tags,
			MaxTags:        d.cols.MaxTags,
			TreePrefix:     item.treePrefix,
			Dimmed:         !item.matched,
			IDColWidth:     d.idColWidth,
			UseFullNames:   d.cols.UseFullTypeStatus,
			ImplicitStatus: item.implicitStatus,
		},
	)

	fmt.Fprint(w, str)
}

// listModel is the model for the item list view
type listModel struct {
	list     list.Model
	resolver *itemgraph.CoreResolver
	config   *config.Config
	width    int
	height   int
	err      error

	// Responsive column state
	hasTags    bool                 // whether any items have tags
	cols       ui.ResponsiveColumns // calculated responsive columns
	idColWidth int                  // ID column width (accounts for tree depth)

	// Active filters
	tagFilter string // if set, only show items with this tag

	// Multi-select state
	selectedItems map[string]bool // IDs of items marked for multi-edit

	// Status message to display in footer
	statusMessage string
}

func newListModel(resolver *itemgraph.CoreResolver, cfg *config.Config) listModel {
	selectedItems := make(map[string]bool)
	delegate := itemDelegate{cfg: cfg, selectedItems: &selectedItems}

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "Items"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	l.Styles.Title = listTitleStyle
	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 1)
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(ui.ColorPrimary)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(ui.ColorPrimary)

	return listModel{
		list:          l,
		resolver:      resolver,
		config:        cfg,
		selectedItems: selectedItems,
	}
}

// itemsLoadedMsg is sent when items are loaded
type itemsLoadedMsg struct {
	items      []ui.FlatItem // flattened tree items
	idColWidth int           // calculated ID column width for tree
}

// errMsg is sent when an error occurs
type errMsg struct {
	err error
}

// selectItemMsg is sent when an item is selected
type selectItemMsg struct {
	item *item.Item
}

func (m listModel) Init() tea.Cmd {
	return m.loadItems
}

func (m listModel) loadItems() tea.Msg {
	// Build filter if tag filter is set
	var filter *model.ItemFilter
	if m.tagFilter != "" {
		filter = &model.ItemFilter{Tags: []string{m.tagFilter}}
	}

	// Query filtered items
	filteredItems, err := m.resolver.Items(context.Background(), filter)
	if err != nil {
		return errMsg{err}
	}

	// Query all items for tree context (ancestors)
	allItems, err := m.resolver.Items(context.Background(), nil)
	if err != nil {
		return errMsg{err}
	}

	// Sort function for tree building
	sortFn := func(items []*item.Item) {
		item.SortByStatusPriorityAndType(items, m.config.StatusNames(), m.config.PriorityNames(), m.config.TypeNames())
	}

	// Pre-compute implicit statuses for all items
	implicitStatuses := make(map[string]string, len(allItems))
	for _, b := range allItems {
		if status, _ := m.resolver.Core.ImplicitStatus(b.ID); status != "" {
			implicitStatuses[b.ID] = status
		}
	}

	// Build tree and flatten it
	tree := ui.BuildTree(filteredItems, allItems, sortFn, implicitStatuses)
	items := ui.FlattenTree(tree)

	// Calculate ID column width based on max ID length and tree depth
	maxIDLen := 0
	for _, b := range allItems {
		if len(b.ID) > maxIDLen {
			maxIDLen = len(b.ID)
		}
	}
	maxDepth := ui.MaxTreeDepth(items)
	// ID column = base ID width + tree indent (3 chars per depth level)
	idColWidth := maxIDLen + 2 // base padding
	if maxDepth > 0 {
		idColWidth += maxDepth * 3 // 3 chars per depth level (├─ + space)
	}

	return itemsLoadedMsg{items: items, idColWidth: idColWidth}
}

// setTagFilter sets the tag filter
func (m *listModel) setTagFilter(tag string) {
	m.tagFilter = tag
}

// clearFilter clears all active filters
func (m *listModel) clearFilter() {
	m.tagFilter = ""
}

// hasActiveFilter returns true if any filter is active
func (m *listModel) hasActiveFilter() bool {
	return m.tagFilter != ""
}

func (m listModel) Update(msg tea.Msg) (listModel, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Track cursor position before update
	prevIndex := m.list.Index()

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		// Reserve space for border and footer
		m.list.SetSize(msg.Width-2, msg.Height-4)
		// Recalculate responsive columns
		m.cols = ui.CalculateResponsiveColumns(msg.Width, m.hasTags)
		m.updateDelegate()

	case itemsLoadedMsg:
		items := make([]list.Item, len(msg.items))
		// Check if any items have tags
		m.hasTags = false
		for i, flatItem := range msg.items {
			items[i] = itemRow{
				item:           flatItem.Item,
				cfg:            m.config,
				treePrefix:     flatItem.TreePrefix,
				matched:        flatItem.Matched,
				implicitStatus: flatItem.ImplicitStatus,
			}
			if len(flatItem.Item.Tags) > 0 {
				m.hasTags = true
			}
		}
		m.list.SetItems(items)
		m.idColWidth = msg.idColWidth
		// Calculate responsive columns based on hasTags and width
		m.cols = ui.CalculateResponsiveColumns(m.width, m.hasTags)
		m.updateDelegate()
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() != list.Filtering {
			switch msg.String() {
			case " ":
				// Toggle selection for multi-select, then move to next item
				if item, ok := m.list.SelectedItem().(itemRow); ok {
					if m.selectedItems[item.item.ID] {
						delete(m.selectedItems, item.item.ID)
					} else {
						m.selectedItems[item.item.ID] = true
					}
					m.list.CursorDown()
				}
				return m, nil
			case "enter":
				if item, ok := m.list.SelectedItem().(itemRow); ok {
					return m, func() tea.Msg {
						return selectItemMsg{item: item.item}
					}
				}
			case "p":
				// Open parent picker for selected item(s)
				if len(m.selectedItems) > 0 {
					// Multi-select mode
					ids := make([]string, 0, len(m.selectedItems))
					types := make([]string, 0, len(m.selectedItems))
					for id := range m.selectedItems {
						ids = append(ids, id)
						// Find the item to get its type
						for _, item := range m.list.Items() {
							if bi, ok := item.(itemRow); ok && bi.item.ID == id {
								types = append(types, bi.item.Type)
								break
							}
						}
					}
					return m, func() tea.Msg {
						return openParentPickerMsg{
							itemIDs:   ids,
							itemTitle: fmt.Sprintf("%d selected items", len(ids)),
							itemTypes: types,
						}
					}
				} else if item, ok := m.list.SelectedItem().(itemRow); ok {
					return m, func() tea.Msg {
						return openParentPickerMsg{
							itemIDs:       []string{item.item.ID},
							itemTitle:     item.item.Title,
							itemTypes:     []string{item.item.Type},
							currentParent: item.item.Parent,
						}
					}
				}
			case "s":
				// Open status picker for selected item(s)
				if len(m.selectedItems) > 0 {
					// Multi-select mode
					ids := make([]string, 0, len(m.selectedItems))
					for id := range m.selectedItems {
						ids = append(ids, id)
					}
					return m, func() tea.Msg {
						return openStatusPickerMsg{
							itemIDs:   ids,
							itemTitle: fmt.Sprintf("%d selected items", len(ids)),
						}
					}
				} else if item, ok := m.list.SelectedItem().(itemRow); ok {
					return m, func() tea.Msg {
						return openStatusPickerMsg{
							itemIDs:       []string{item.item.ID},
							itemTitle:     item.item.Title,
							currentStatus: item.item.Status,
						}
					}
				}
			case "t":
				// Open type picker for selected item(s)
				if len(m.selectedItems) > 0 {
					// Multi-select mode
					ids := make([]string, 0, len(m.selectedItems))
					for id := range m.selectedItems {
						ids = append(ids, id)
					}
					return m, func() tea.Msg {
						return openTypePickerMsg{
							itemIDs:   ids,
							itemTitle: fmt.Sprintf("%d selected items", len(ids)),
						}
					}
				} else if item, ok := m.list.SelectedItem().(itemRow); ok {
					return m, func() tea.Msg {
						return openTypePickerMsg{
							itemIDs:     []string{item.item.ID},
							itemTitle:   item.item.Title,
							currentType: item.item.Type,
						}
					}
				}
			case "P":
				// Open priority picker for selected item(s)
				if len(m.selectedItems) > 0 {
					// Multi-select mode
					ids := make([]string, 0, len(m.selectedItems))
					for id := range m.selectedItems {
						ids = append(ids, id)
					}
					return m, func() tea.Msg {
						return openPriorityPickerMsg{
							itemIDs:   ids,
							itemTitle: fmt.Sprintf("%d selected items", len(ids)),
						}
					}
				} else if item, ok := m.list.SelectedItem().(itemRow); ok {
					return m, func() tea.Msg {
						return openPriorityPickerMsg{
							itemIDs:         []string{item.item.ID},
							itemTitle:       item.item.Title,
							currentPriority: item.item.Priority,
						}
					}
				}
			case "b":
				// Open blocking picker for selected item
				if item, ok := m.list.SelectedItem().(itemRow); ok {
					return m, func() tea.Msg {
						return openBlockingPickerMsg{
							itemID:          item.item.ID,
							itemTitle:       item.item.Title,
							currentBlocking: item.item.Blocking,
						}
					}
				}
			case "c":
				// Open create modal
				return m, func() tea.Msg {
					return openCreateModalMsg{}
				}
			case "e":
				// Open editor for selected item
				if item, ok := m.list.SelectedItem().(itemRow); ok {
					return m, func() tea.Msg {
						return openEditorMsg{
							itemID:   item.item.ID,
							itemPath: item.item.Path,
						}
					}
				}
			case "y":
				// Copy item ID(s) to clipboard
				if len(m.selectedItems) > 0 {
					// Multi-select mode: copy all selected IDs
					ids := make([]string, 0, len(m.selectedItems))
					for id := range m.selectedItems {
						ids = append(ids, id)
					}
					return m, func() tea.Msg {
						return copyItemIDMsg{ids: ids}
					}
				} else if item, ok := m.list.SelectedItem().(itemRow); ok {
					// Single item mode
					return m, func() tea.Msg {
						return copyItemIDMsg{ids: []string{item.item.ID}}
					}
				}
			case "esc", "backspace":
				// First clear selection if any items are selected
				if len(m.selectedItems) > 0 {
					clear(m.selectedItems)
					return m, nil
				}
				// Then clear active filter if any
				if m.hasActiveFilter() {
					return m, func() tea.Msg {
						return clearFilterMsg{}
					}
				}
			}
		}
	}

	// Always forward to the list component
	m.list, cmd = m.list.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// Check if cursor moved and emit message
	if m.list.Index() != prevIndex {
		if item, ok := m.list.SelectedItem().(itemRow); ok {
			cmds = append(cmds, func() tea.Msg {
				return cursorChangedMsg{itemID: item.item.ID}
			})
		}
	}

	return m, tea.Batch(cmds...)
}

// updateDelegate updates the list delegate with current responsive columns
func (m *listModel) updateDelegate() {
	delegate := itemDelegate{
		cfg:           m.config,
		hasTags:       m.hasTags,
		width:         m.width,
		cols:          m.cols,
		idColWidth:    m.idColWidth,
		selectedItems: &m.selectedItems,
	}
	m.list.SetDelegate(delegate)
}

func (m listModel) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\nPress q to quit.", m.err)
	}

	if m.width == 0 {
		return "Loading..."
	}

	// Update title based on active filter
	if m.tagFilter != "" {
		m.list.Title = fmt.Sprintf("Items [tag: %s]", m.tagFilter)
	} else {
		m.list.Title = "Items"
	}

	// Inner height: total height minus border (2) minus footer (1) minus padding (1)
	return m.viewContent(m.height-4) + "\n" + m.Footer()
}

// viewContent renders just the bordered list without footer.
// innerHeight is the content height inside the border (not including border lines).
func (m listModel) viewContent(innerHeight int) string {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.ColorMuted).
		Width(m.width - 2).
		Height(innerHeight)

	return border.Render(m.list.View())
}

// Footer renders the help/status footer for the list view.
func (m listModel) Footer() string {
	var help string

	// Show selection count if any items are selected
	var selectionPrefix string
	if len(m.selectedItems) > 0 {
		selectionStyle := lipgloss.NewStyle().Foreground(ui.ColorWarning).Bold(true)
		selectionPrefix = selectionStyle.Render(fmt.Sprintf("(%d selected) ", len(m.selectedItems)))
	}

	if len(m.selectedItems) > 0 {
		// When items are selected, show esc to clear selection
		help = helpKeyStyle.Render("space") + " " + helpStyle.Render("toggle") + "  " +
			helpKeyStyle.Render("P") + " " + helpStyle.Render("priority") + "  " +
			helpKeyStyle.Render("s") + " " + helpStyle.Render("status") + "  " +
			helpKeyStyle.Render("t") + " " + helpStyle.Render("type") + "  " +
			helpKeyStyle.Render("y") + " " + helpStyle.Render("copy id") + "  " +
			helpKeyStyle.Render("esc") + " " + helpStyle.Render("clear selection") + "  " +
			helpKeyStyle.Render("?") + " " + helpStyle.Render("help") + "  " +
			helpKeyStyle.Render("q") + " " + helpStyle.Render("quit")
	} else if m.hasActiveFilter() {
		help = helpKeyStyle.Render("space") + " " + helpStyle.Render("select") + "  " +
			helpKeyStyle.Render("enter") + " " + helpStyle.Render("view") + "  " +
			helpKeyStyle.Render("b") + " " + helpStyle.Render("blocking") + "  " +
			helpKeyStyle.Render("c") + " " + helpStyle.Render("create") + "  " +
			helpKeyStyle.Render("e") + " " + helpStyle.Render("edit") + "  " +
			helpKeyStyle.Render("p") + " " + helpStyle.Render("parent") + "  " +
			helpKeyStyle.Render("P") + " " + helpStyle.Render("priority") + "  " +
			helpKeyStyle.Render("s") + " " + helpStyle.Render("status") + "  " +
			helpKeyStyle.Render("t") + " " + helpStyle.Render("type") + "  " +
			helpKeyStyle.Render("y") + " " + helpStyle.Render("copy id") + "  " +
			helpKeyStyle.Render("esc") + " " + helpStyle.Render("clear filter") + "  " +
			helpKeyStyle.Render("?") + " " + helpStyle.Render("help") + "  " +
			helpKeyStyle.Render("q") + " " + helpStyle.Render("quit")
	} else {
		help = helpKeyStyle.Render("space") + " " + helpStyle.Render("select") + "  " +
			helpKeyStyle.Render("enter") + " " + helpStyle.Render("view") + "  " +
			helpKeyStyle.Render("b") + " " + helpStyle.Render("blocking") + "  " +
			helpKeyStyle.Render("c") + " " + helpStyle.Render("create") + "  " +
			helpKeyStyle.Render("e") + " " + helpStyle.Render("edit") + "  " +
			helpKeyStyle.Render("p") + " " + helpStyle.Render("parent") + "  " +
			helpKeyStyle.Render("P") + " " + helpStyle.Render("priority") + "  " +
			helpKeyStyle.Render("s") + " " + helpStyle.Render("status") + "  " +
			helpKeyStyle.Render("t") + " " + helpStyle.Render("type") + "  " +
			helpKeyStyle.Render("y") + " " + helpStyle.Render("copy id") + "  " +
			helpKeyStyle.Render("/") + " " + helpStyle.Render("filter") + "  " +
			helpKeyStyle.Render("?") + " " + helpStyle.Render("help") + "  " +
			helpKeyStyle.Render("q") + " " + helpStyle.Render("quit")
	}

	// Show status message if present, otherwise show help
	footer := selectionPrefix
	if m.statusMessage != "" {
		statusStyle := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)
		footer += statusStyle.Render(m.statusMessage)
	} else {
		footer += help
	}

	return footer
}

// ViewConstrained renders the list constrained to the given width and height.
// Used for the left pane in two-column mode. Returns only the content without footer.
// The output will be exactly `height` lines tall.
func (m listModel) ViewConstrained(width, height int) string {
	// Temporarily set constrained dimensions
	m.width = width
	m.height = height

	// Inner height for border content (height minus 2 for top/bottom border)
	innerHeight := height - 2
	m.list.SetSize(width-2, innerHeight)

	// Recalculate columns for constrained width
	m.cols = ui.CalculateResponsiveColumns(width, m.hasTags)
	m.updateDelegate()

	// Update title based on active filter
	if m.tagFilter != "" {
		m.list.Title = fmt.Sprintf("Items [tag: %s]", m.tagFilter)
	} else {
		m.list.Title = "Items"
	}

	return m.viewContent(innerHeight)
}
