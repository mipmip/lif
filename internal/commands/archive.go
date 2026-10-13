package commands

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/xRiErOS/lif/internal/output"
	"github.com/xRiErOS/lif/pkg/item"
)

var archiveJSON bool

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Args:  cobra.NoArgs,
	Short: "Move completed/scrapped beans to the archive",
	Long: `Moves all beans with status "completed" or "scrapped" to the archive directory (.lif/archive/).
Archived beans are preserved for project memory and remain visible in all queries.
The archive keeps the main .lif directory tidy while preserving project history.
A project that still uses the legacy name .beans has its archive in .beans/archive/.

Relationships (parent, blocking) are preserved in archived beans.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		allItems := core.All()

		// Find items with any archive status
		var archiveItems []*item.Item
		archiveSet := make(map[string]bool)
		for _, b := range allItems {
			if cfg.IsArchiveStatus(b.Status) {
				archiveItems = append(archiveItems, b)
				archiveSet[b.ID] = true
			}
		}

		if len(archiveItems) == 0 {
			if archiveJSON {
				return output.SuccessMessage("No beans to archive")
			}
			fmt.Println("No beans with archive status to archive.")
			return nil
		}

		// Sort items for consistent display
		item.SortByStatusPriorityAndType(archiveItems, cfg.StatusNames(), cfg.PriorityNames(), cfg.TypeNames())

		// Archive all items with archive status
		var archived []string
		for _, b := range archiveItems {
			if err := core.Archive(b.ID); err != nil {
				if archiveJSON {
					return output.Error(output.ErrFileError, fmt.Sprintf("failed to archive bean %s: %s", b.ID, err.Error()))
				}
				return fmt.Errorf("failed to archive bean %s: %w", b.ID, err)
			}
			archived = append(archived, b.ID)
		}

		archiveDir := filepath.Base(core.Root()) + "/archive/"
		if archiveJSON {
			return output.SuccessMessage(fmt.Sprintf("Archived %d bean(s) to %s", len(archived), archiveDir))
		}

		fmt.Printf("Archived %d bean(s) to %s\n", len(archived), archiveDir)
		return nil
	},
}

func RegisterArchiveCmd(root *cobra.Command) {
	archiveCmd.Flags().BoolVar(&archiveJSON, "json", false, "Output as JSON")
	root.AddCommand(archiveCmd)
}
