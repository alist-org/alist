/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

// storageCmd represents the storage command
var storageCmd = &cobra.Command{
	Use:   "storage",
	Short: "Manage storage",
}

var disableStorageCmd = &cobra.Command{
	Use:   "disable",
	Short: "Disable a storage",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			utils.Log.Errorf("mount path is required")
			return
		}
		mountPath := args[0]
		Init()
		defer Release()
		storage, err := db.GetStorageByMountPath(mountPath)
		if err != nil {
			utils.Log.Errorf("failed to query storage: %+v", err)
		} else {
			storage.Disabled = true
			err = db.UpdateStorage(storage)
			if err != nil {
				utils.Log.Errorf("failed to update storage: %+v", err)
			} else {
				utils.Log.Infof("Storage with mount path [%s] have been disabled", mountPath)
			}
		}
	},
}

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type model struct {
	table table.Model
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
			//case "enter":
			//	return m, tea.Batch(
			//		tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			//	)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return baseStyle.Render(m.table.View()) + "\n"
}

var storageTableHeight int
var listStorageCmd = &cobra.Command{
	Use:   "list",
	Short: "List all storages",
	Run: func(cmd *cobra.Command, args []string) {
		Init()
		defer Release()
		storages, _, err := db.GetStorages(1, -1)
		if err != nil {
			utils.Log.Errorf("failed to query storages: %+v", err)
		} else {
			utils.Log.Infof("Found %d storages", len(storages))
			columns := []table.Column{
				{Title: "ID", Width: 4},
				{Title: "Driver", Width: 16},
				{Title: "Mount Path", Width: 30},
				{Title: "Enabled", Width: 7},
			}

			var rows []table.Row
			for i := range storages {
				storage := storages[i]
				enabled := "true"
				if storage.Disabled {
					enabled = "false"
				}
				rows = append(rows, table.Row{
					strconv.Itoa(int(storage.ID)),
					storage.Driver,
					storage.MountPath,
					enabled,
				})
			}
			t := table.New(
				table.WithColumns(columns),
				table.WithRows(rows),
				table.WithFocused(true),
				table.WithHeight(storageTableHeight),
			)

			s := table.DefaultStyles()
			s.Header = s.Header.
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("240")).
				BorderBottom(true).
				Bold(false)
			s.Selected = s.Selected.
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("57")).
				Bold(false)
			t.SetStyles(s)

			m := model{t}
			if _, err := tea.NewProgram(m).Run(); err != nil {
				utils.Log.Errorf("failed to run program: %+v", err)
				os.Exit(1)
			}
		}
	},
}

func init() {

	RootCmd.AddCommand(storageCmd)
	storageCmd.AddCommand(disableStorageCmd)
	storageCmd.AddCommand(listStorageCmd)
	storageCmd.PersistentFlags().IntVarP(&storageTableHeight, "height", "H", 10, "Table height")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// storageCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// storageCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
