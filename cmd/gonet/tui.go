package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deltaog-117/gonet/internal/connect"
	"github.com/deltaog-117/gonet/internal/scan"
	"github.com/deltaog-117/gonet/internal/shared/models"
)

// TUI model
type model struct {
	list           list.Model
	networks       []models.Network
	selected       *models.Network
	connecting     bool
	passwordInput  textinput.Model
	showPassword   bool
	statusMessage  string
	errorMessage   string
	width          int
	height         int
	ready          bool
}

// Item implements list.Item for BubbleTea list.
type networkItem struct {
	network models.Network
}

func (i networkItem) Title() string {
	return i.network.SSID
}

func (i networkItem) Description() string {
	return fmt.Sprintf("%s | %d dBm | %s", i.network.BSSID, i.network.Signal, i.network.SecurityType)
}

func (i networkItem) FilterValue() string {
	return i.network.SSID
}

// Init is the BubbleTea init function.
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tuiScanCmd(), // initial scan
	)
}

// Update is the BubbleTea update loop.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 8)
		m.ready = true
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q", "esc":
			return m, tea.Quit

		case "enter":
			if m.connecting {
				return m, nil
			}
			// If password input is visible, attempt connection
			if m.showPassword {
				psk := m.passwordInput.Value()
				if psk == "" {
					m.errorMessage = "Password cannot be empty"
					return m, nil
				}
				m.connecting = true
				m.statusMessage = fmt.Sprintf("Connecting to %s...", m.selected.SSID)
				m.showPassword = false
				return m, tuiConnectCmd(m.selected.SSID, psk)
			}
			// Otherwise, select the highlighted network
			if i, ok := m.list.SelectedItem().(networkItem); ok {
				m.selected = &i.network
				// Show password prompt
				m.showPassword = true
				m.passwordInput.Focus()
				m.passwordInput.SetValue("")
				m.errorMessage = ""
				m.statusMessage = fmt.Sprintf("Enter password for %s", m.selected.SSID)
				return m, textinput.Blink
			}
		}

		// If password input is active, delegate key events to it
		if m.showPassword {
			var cmd tea.Cmd
			m.passwordInput, cmd = m.passwordInput.Update(msg)
			return m, cmd
		}

		// Otherwise, delegate to the list
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd

	case scanResultMsg:
		m.networks = msg.networks
		m.statusMessage = fmt.Sprintf("Scan complete: %d networks found", len(m.networks))
		m.errorMessage = ""
		items := make([]list.Item, len(m.networks))
		for i, net := range m.networks {
			items[i] = networkItem{network: net}
		}
		m.list.SetItems(items)
		return m, nil

	case scanErrorMsg:
		m.errorMessage = fmt.Sprintf("Scan error: %v", msg.err)
		m.statusMessage = "Scan failed"
		return m, nil

	case connectResultMsg:
		m.connecting = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Connection failed: %v", msg.err)
			m.statusMessage = "Connection failed"
		} else {
			m.statusMessage = fmt.Sprintf("Connected to %s successfully!", msg.ssid)
			m.errorMessage = ""
		}
		return m, nil

	case tickMsg:
		// Refresh scan
		return m, tuiScanCmd()

	default:
		// Handle list updates when not in password mode
		if !m.showPassword {
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}
		return m, nil
	}
}

// View renders the TUI.
func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var b strings.Builder

	// Title
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00FF00")).
		Padding(0, 1).
		Render("🛜 gonet - Wi-Fi Manager")
	b.WriteString(title + "\n\n")

	// Status line
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1)
	status := m.statusMessage
	if m.errorMessage != "" {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Render(m.errorMessage)
	}
	b.WriteString(statusStyle.Render(status))
	b.WriteString("\n\n")

	// Password prompt if active
	if m.showPassword {
		promptStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00")).
			Bold(true)
		b.WriteString(promptStyle.Render("Enter password: "))
		b.WriteString(m.passwordInput.View())
		b.WriteString("\n\n")
	}

	// Network list
	listView := m.list.View()
	// If we are connecting, overlay a message
	if m.connecting {
		connectingStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF")).
			Bold(true).
			Padding(1, 2)
		b.WriteString(connectingStyle.Render("⏳ Connecting..."))
	} else {
		b.WriteString(listView)
	}

	// Help footer
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 1)
	footerText := "↑/↓: navigate • Enter: select/connect • Esc/q: quit"
	if m.showPassword {
		footerText = "Type password • Enter: connect • Esc: cancel"
	}
	b.WriteString("\n" + footer.Render(footerText))

	return b.String()
}

// --- TUI-specific Commands and Messages ---

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

type scanResultMsg struct {
	networks []models.Network
}

type scanErrorMsg struct {
	err error
}

// tuiScanCmd is the TUI version of the scan command.
func tuiScanCmd() tea.Cmd {
	return func() tea.Msg {
		networks, err := scan.ScanInterface(globalIface)
		if err != nil {
			return scanErrorMsg{err: err}
		}
		return scanResultMsg{networks: networks}
	}
}

type connectResultMsg struct {
	ssid string
	err  error
}

// tuiConnectCmd is the TUI version of the connect command.
func tuiConnectCmd(ssid, psk string) tea.Cmd {
	return func() tea.Msg {
		err := connect.ConnectInterface(globalIface, ssid, psk)
		return connectResultMsg{ssid: ssid, err: err}
	}
}

// NewModel creates a new TUI model with default list and input.
func NewModel(iface string) model {
	// Create a list
	items := []list.Item{}
	listDelegate := list.NewDefaultDelegate()
	listDelegate.SetSpacing(1)
	listModel := list.New(items, listDelegate, 0, 0)
	listModel.Title = "Available Networks"
	listModel.SetFilteringEnabled(false) // disable filtering for simplicity

	// Create password input
	pi := textinput.New()
	pi.EchoMode = textinput.EchoPassword
	pi.EchoCharacter = '•'
	pi.Placeholder = "Enter PSK"

	return model{
		list:          listModel,
		passwordInput: pi,
		connecting:    false,
		showPassword:  false,
		statusMessage: "Scanning...",
	}
}
