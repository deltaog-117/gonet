package main

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deltaog-117/gonet/internal/connect"
	"github.com/deltaog-117/gonet/internal/scan"
	"github.com/deltaog-117/gonet/internal/status"
	"github.com/deltaog-117/gonet/internal/shared/models"
)

// --- Styles ---
var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#00FF00")).
			Padding(0, 1)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#333333")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FFFF"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	detailStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#555555")).
			Padding(1, 2)
)

// --- Key Bindings ---
type keyMap struct {
	Up      key.Binding
	Down    key.Binding
	Enter   key.Binding
	Escape  key.Binding
	Quit    key.Binding
	Refresh key.Binding
	Details key.Binding
	Sort    key.Binding
	Help    key.Binding
	Hidden  key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Enter, k.Escape},
		{k.Refresh, k.Details, k.Sort},
		{k.Hidden, k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
	Details: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "details"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "sort"),
	),
	Help: key.NewBinding(
		key.WithKeys("h", "?"),
		key.WithHelp("h", "help"),
	),
	Hidden: key.NewBinding(
		key.WithKeys("n"),
		key.WithHelp("n", "hidden"),
	),
}

// --- UI State ---
type state int

const (
	stateList state = iota
	statePassword
	stateHidden
	stateDetail
	stateConnecting
)

// --- Model ---
type model struct {
	list          list.Model
	passwordInput textinput.Model
	hiddenSSID    textinput.Model
	hiddenPSK     textinput.Model
	spinner       spinner.Model
	help          help.Model

	state          state
	networks       []models.Network
	selected       *models.Network
	connected      *models.Network
	connecting     bool
	successMessage string

	statusMessage string
	errorMessage  string
	sortBySignal  bool

	width  int
	height int
	ready  bool
}

// --- Custom List Delegate ---
type customDelegate struct {
	list.DefaultDelegate
}

func (d customDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(networkItem)
	if !ok {
		return
	}

	// Fixed column widths
	const (
		numWidth   = 3
		ssidWidth  = 28
		signalWidth = 6
		dbmWidth   = 8
	)

	// Number
	num := fmt.Sprintf("%2d.", index+1)

	// SSID (truncated)
	ssid := i.network.SSID
	if len(ssid) > ssidWidth-2 {
		ssid = ssid[:ssidWidth-3] + "…"
	}
	ssid = fmt.Sprintf("%-*s", ssidWidth, ssid)

	// Signal bars
	bars := getSignalBars(i.network.Signal)
	signalColor := getSignalColor(i.network.Signal)
	signalBars := lipgloss.NewStyle().Foreground(signalColor).Render(bars)
	signalBars = fmt.Sprintf("%-*s", signalWidth, signalBars)

	// Signal dBm
	dbm := fmt.Sprintf("%3d dBm", i.network.Signal)
	dbm = fmt.Sprintf("%-*s", dbmWidth, dbm)

	// Security
	secIcon := getSecurityIcon(i.network.SecurityType)
	secType := i.network.SecurityType.String()
	security := fmt.Sprintf("%s %s", secIcon, secType)

	// Connected indicator
	check := "  "
	if i.connected {
		check = successStyle.Render("✔ ")
	}

	// Build line
	line := fmt.Sprintf("%s%s %s %s %s %s",
		check,
		num,
		ssid,
		signalBars,
		dbm,
		security,
	)

	// Selected style
	if index == m.Index() {
		line = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Render("▶ " + line)
	} else {
		line = "  " + line
	}

	fmt.Fprint(w, line)
}

// --- Network Item ---
type networkItem struct {
	network   models.Network
	connected bool
}

func (i networkItem) FilterValue() string {
	return i.network.SSID
}

// --- Helper Functions ---

func getSecurityIcon(st models.SecurityType) string {
	switch st {
	case models.SecurityOpen:
		return "🔓"
	case models.SecurityWEP:
		return "🔐"
	case models.SecurityWPA2:
		return "🔐"
	case models.SecurityWPA3:
		return "🔐"
	case models.SecurityWPA2Enterprise:
		return "🔒"
	case models.SecurityWPA3Enterprise:
		return "🔒"
	default:
		return "?"
	}
}

func getSignalBars(dbm int) string {
	if dbm >= -30 {
		return "██████"
	} else if dbm >= -50 {
		return "█████ "
	} else if dbm >= -60 {
		return "████  "
	} else if dbm >= -70 {
		return "███   "
	} else if dbm >= -80 {
		return "██    "
	} else {
		return "█     "
	}
}

func getSignalColor(dbm int) lipgloss.Color {
	if dbm >= -50 {
		return lipgloss.Color("#00FF00")
	} else if dbm >= -60 {
		return lipgloss.Color("#88FF00")
	} else if dbm >= -70 {
		return lipgloss.Color("#FFFF00")
	} else if dbm >= -80 {
		return lipgloss.Color("#FF8800")
	} else {
		return lipgloss.Color("#FF0000")
	}
}

// --- Messages ---

type scanResultMsg struct {
	networks []models.Network
	err      error
}

type statusResultMsg struct {
	status *status.Status
	err    error
}

type connectResultMsg struct {
	ssid string
	err  error
}

type tickMsg struct{}

// --- TUI Commands ---

func tuiScanCmd() tea.Cmd {
	return func() tea.Msg {
		networks, err := scan.ScanInterface(globalIface)
		return scanResultMsg{networks: networks, err: err}
	}
}

func tuiStatusCmd() tea.Cmd {
	return func() tea.Msg {
		s, err := status.GetStatusInterface(globalIface)
		return statusResultMsg{status: s, err: err}
	}
}

func tuiConnectCmd(ssid, psk string) tea.Cmd {
	return func() tea.Msg {
		err := connect.ConnectInterface(globalIface, ssid, psk)
		return connectResultMsg{ssid: ssid, err: err}
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg{}
	})
}

// --- Initialization ---

func NewModel(iface string) model {
	delegate := customDelegate{
		DefaultDelegate: list.NewDefaultDelegate(),
	}
	delegate.SetSpacing(0)
	delegate.ShowDescription = false

	items := []list.Item{}
	listModel := list.New(items, delegate, 0, 0)
	listModel.Title = "🌐 Available Networks"
	listModel.SetFilteringEnabled(false)
	listModel.Styles.Title = titleStyle
	listModel.Styles.PaginationStyle = helpStyle

	s := spinner.New()
	s.Spinner = spinner.Dot

	pi := textinput.New()
	pi.EchoMode = textinput.EchoPassword
	pi.EchoCharacter = '•'
	pi.Placeholder = "Enter password"
	pi.Focus()

	hs := textinput.New()
	hs.Placeholder = "Enter SSID"
	hp := textinput.New()
	hp.EchoMode = textinput.EchoPassword
	hp.EchoCharacter = '•'
	hp.Placeholder = "Enter password"

	return model{
		list:          listModel,
		passwordInput: pi,
		hiddenSSID:    hs,
		hiddenPSK:     hp,
		spinner:       s,
		help:          help.New(),
		state:         stateList,
		statusMessage: "Press r to scan",
		sortBySignal:  true,
	}
}

// --- Init ---

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		tuiScanCmd(),
		tuiStatusCmd(),
		tickCmd(),
		m.spinner.Tick,
	)
}

// --- Update ---

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.list.SetWidth(msg.Width - 4)
		m.list.SetHeight(msg.Height - 8)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "esc":
			switch m.state {
			case statePassword, stateHidden, stateDetail:
				m.state = stateList
				m.errorMessage = ""
				m.statusMessage = "Press r to scan"
				return m, nil
			default:
				return m, tea.Quit
			}

		case "r":
			m.statusMessage = "Scanning..."
			m.errorMessage = ""
			return m, tuiScanCmd()

		case "d":
			if m.state == stateList && m.selected != nil {
				m.state = stateDetail
				return m, nil
			}

		case "s":
			m.sortBySignal = !m.sortBySignal
			m.statusMessage = fmt.Sprintf("Sorting by %s", map[bool]string{true: "signal", false: "name"}[m.sortBySignal])
			items := buildItems(m.networks, m.connected, m.sortBySignal)
			m.list.SetItems(items)
			return m, nil

		case "n":
			if m.state == stateList {
				m.state = stateHidden
				m.hiddenSSID.Focus()
				return m, nil
			}

		case "enter":
			switch m.state {
			case stateList:
				if i, ok := m.list.SelectedItem().(networkItem); ok {
					m.selected = &i.network
					m.state = statePassword
					m.passwordInput.Focus()
					m.passwordInput.SetValue("")
					m.statusMessage = fmt.Sprintf("Password for %s:", i.network.SSID)
					return m, textinput.Blink
				}
			case statePassword:
				psk := m.passwordInput.Value()
				if psk == "" {
					m.errorMessage = "Password cannot be empty"
					return m, nil
				}
				m.state = stateConnecting
				m.statusMessage = fmt.Sprintf("Connecting to %s...", m.selected.SSID)
				return m, tuiConnectCmd(m.selected.SSID, psk)
			case stateHidden:
				ssid := m.hiddenSSID.Value()
				psk := m.hiddenPSK.Value()
				if ssid == "" {
					m.errorMessage = "SSID cannot be empty"
					return m, nil
				}
				if psk == "" {
					m.errorMessage = "Password cannot be empty"
					return m, nil
				}
				m.state = stateConnecting
				m.statusMessage = fmt.Sprintf("Connecting to hidden network %s...", ssid)
				return m, tuiConnectCmd(ssid, psk)
			case stateDetail:
				if m.selected != nil {
					m.state = statePassword
					m.passwordInput.Focus()
					m.passwordInput.SetValue("")
					m.statusMessage = fmt.Sprintf("Password for %s:", m.selected.SSID)
					return m, textinput.Blink
				}
			}
		}

		switch m.state {
		case statePassword:
			var cmd tea.Cmd
			m.passwordInput, cmd = m.passwordInput.Update(msg)
			return m, cmd
		case stateHidden:
			var cmd1, cmd2 tea.Cmd
			m.hiddenSSID, cmd1 = m.hiddenSSID.Update(msg)
			m.hiddenPSK, cmd2 = m.hiddenPSK.Update(msg)
			return m, tea.Batch(cmd1, cmd2)
		default:
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		}

	case scanResultMsg:
		if msg.err != nil {
			if strings.Contains(msg.err.Error(), "busy") {
				m.statusMessage = "⏳ Interface busy, retrying..."
			} else {
				m.errorMessage = fmt.Sprintf("Scan failed: %v", msg.err)
				m.statusMessage = "Scan failed"
			}
		} else {
			m.networks = msg.networks
			m.statusMessage = fmt.Sprintf("%d networks", len(m.networks))
			m.errorMessage = ""
			items := buildItems(m.networks, m.connected, m.sortBySignal)
			m.list.SetItems(items)
		}
		return m, nil
		m.networks = msg.networks
		m.statusMessage = fmt.Sprintf("%d networks", len(m.networks))
		m.errorMessage = ""
		items := buildItems(m.networks, m.connected, m.sortBySignal)
		m.list.SetItems(items)
		return m, nil

	case statusResultMsg:
		if msg.err == nil && msg.status != nil && msg.status.SSID != "" {
			found := false
			for i := range m.networks {
				if m.networks[i].SSID == msg.status.SSID {
					m.connected = &m.networks[i]
					found = true
					break
				}
			}
			if !found {
				m.connected = &models.Network{
					SSID:   msg.status.SSID,
					Signal: msg.status.Signal,
				}
			}
			items := buildItems(m.networks, m.connected, m.sortBySignal)
			m.list.SetItems(items)
		}
		return m, nil

	case connectResultMsg:
		m.connecting = false
		if msg.err != nil {
			m.errorMessage = fmt.Sprintf("Connection failed: %v", msg.err)
			m.statusMessage = "Connection failed"
			m.state = stateList
			return m, nil
		}
		m.successMessage = msg.ssid
		m.statusMessage = fmt.Sprintf("✅ Connected to %s", msg.ssid)
		m.state = stateList
		return m, tea.Batch(tuiScanCmd(), tuiStatusCmd())

	case tickMsg:
		if m.state != stateConnecting && m.state != statePassword && m.state != stateHidden {
			return m, tuiScanCmd()
		}
		return m, tickCmd()

	case spinner.TickMsg:
		if m.state == stateConnecting || m.statusMessage == "Scanning..." {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	default:
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}
}

// --- Helper: Build Items ---

func buildItems(networks []models.Network, connected *models.Network, sortBySignal bool) []list.Item {
	sorted := make([]models.Network, len(networks))
	copy(sorted, networks)

	if sortBySignal {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Signal > sorted[j].Signal
		})
	} else {
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].SSID < sorted[j].SSID
		})
	}

	items := make([]list.Item, len(sorted))
	for i, net := range sorted {
		conn := false
		if connected != nil && connected.SSID == net.SSID {
			conn = true
		}
		items[i] = networkItem{network: net, connected: conn}
	}
	return items
}

// --- View ---

func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var content strings.Builder

	content.WriteString(titleStyle.Render("🌐 gonet"))
	content.WriteString("\n\n")

	status := m.statusMessage
	if m.errorMessage != "" {
		status = errorStyle.Render("⚠ " + m.errorMessage)
	} else if m.state == stateConnecting {
		status = m.spinner.View() + " " + status
	}
	content.WriteString(statusStyle.Render(status))
	content.WriteString("\n\n")

	switch m.state {
	case stateList:
		content.WriteString(m.list.View())
		content.WriteString("\n")

		if m.connected != nil && m.connected.SSID != "" {
			content.WriteString(successStyle.Render("  Connected to: " + m.connected.SSID))
			content.WriteString("\n")
		}

		content.WriteString(helpStyle.Render("  [↑/↓] navigate  [Enter] select  [r] refresh  [n] hidden  [d] details  [s] sort  [q] quit"))

	case statePassword:
		content.WriteString(infoStyle.Render("🔑 " + m.statusMessage))
		content.WriteString("\n")
		content.WriteString(m.passwordInput.View())
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("  [Enter] confirm  [Esc] cancel"))

	case stateHidden:
		content.WriteString(infoStyle.Render("🔍 Connect to hidden network:"))
		content.WriteString("\n")
		content.WriteString("  SSID: " + m.hiddenSSID.View())
		content.WriteString("\n")
		content.WriteString("  PSK:  " + m.hiddenPSK.View())
		content.WriteString("\n\n")
		content.WriteString(helpStyle.Render("  [Enter] confirm  [Esc] cancel"))

	case stateDetail:
		if m.selected != nil {
			content.WriteString(detailStyle.Render(m.getDetailView()))
			content.WriteString("\n\n")
			content.WriteString(helpStyle.Render("  [Enter] connect  [Esc] back"))
		}

	case stateConnecting:
		content.WriteString(spinnerStyle.Render("⏳ " + m.statusMessage))
	}

	return appStyle.Render(content.String())
}

// --- Helper Methods ---

func (m model) getDetailView() string {
	n := m.selected
	var b strings.Builder

	b.WriteString(titleStyle.Render("📡 Network Details"))
	b.WriteString("\n\n")

	b.WriteString(fmt.Sprintf("  SSID:      %s\n", n.SSID))
	b.WriteString(fmt.Sprintf("  BSSID:     %s\n", n.BSSID))
	b.WriteString(fmt.Sprintf("  Signal:    %d dBm  %s\n", n.Signal, getSignalBars(n.Signal)))
	b.WriteString(fmt.Sprintf("  Channel:   %d\n", n.Channel))
	b.WriteString(fmt.Sprintf("  Security:  %s\n", n.SecurityType))

	if m.connected != nil && m.connected.SSID == n.SSID {
		b.WriteString("\n")
		b.WriteString(successStyle.Render("  ✅ Currently connected"))
	}

	return b.String()
}

// --- Styles ---

var spinnerStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#00FFFF"))

