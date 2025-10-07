package gui

import (
	"fmt"
	"log"
	"time"

	"fyne.io/systray"
	"github.com/arnavpraneet/lyric-clipboard-app/internal/orchestrator"
)

// SystemTray manages the system tray icon and menu
type SystemTray struct {
	orchestrator   *orchestrator.Orchestrator
	statusItem     *systray.MenuItem
	clipboardItem  *systray.MenuItem
	offsetItems    map[int]*systray.MenuItem
	currentOffset  time.Duration
}

// NewSystemTray creates a new system tray manager
func NewSystemTray(orch *orchestrator.Orchestrator) *SystemTray {
	return &SystemTray{
		orchestrator:  orch,
		offsetItems:   make(map[int]*systray.MenuItem),
		currentOffset: 0,
	}
}

// Run starts the system tray GUI
func (st *SystemTray) Run() {
	systray.Run(st.onReady, st.onExit)
}

// onReady is called when the system tray is ready
func (st *SystemTray) onReady() {
	// Set icon and tooltip
	if len(iconData) > 0 {
		systray.SetIcon(iconData)
	}
	systray.SetTitle("Lyric Clipboard")
	systray.SetTooltip("Lyric Clipboard - Syncing lyrics to clipboard")

	// Status display (disabled menu item for display only)
	st.statusItem = systray.AddMenuItem("Status: Starting...", "Current status")
	st.statusItem.Disable()

	systray.AddSeparator()

	// Clipboard toggle
	st.clipboardItem = systray.AddMenuItem("✓ Clipboard Updates", "Enable/disable clipboard updates")

	systray.AddSeparator()

	// Lyric offset submenu
	mOffset := systray.AddMenuItem("Lyric Offset", "Adjust lyric timing")
	st.offsetItems[-2000] = mOffset.AddSubMenuItem("-2.0s", "Delay lyrics by 2 seconds")
	st.offsetItems[-1000] = mOffset.AddSubMenuItem("-1.0s", "Delay lyrics by 1 second")
	st.offsetItems[-500] = mOffset.AddSubMenuItem("-0.5s", "Delay lyrics by 0.5 seconds")
	st.offsetItems[0] = mOffset.AddSubMenuItem("0s (None)", "No offset")
	st.offsetItems[0].Check() // Default
	st.offsetItems[500] = mOffset.AddSubMenuItem("+0.5s", "Advance lyrics by 0.5 seconds")
	st.offsetItems[1000] = mOffset.AddSubMenuItem("+1.0s", "Advance lyrics by 1 second")
	st.offsetItems[2000] = mOffset.AddSubMenuItem("+2.0s", "Advance lyrics by 2 seconds")

	systray.AddSeparator()

	// Configuration
	mConfig := systray.AddMenuItem("Open Config", "Open configuration file")

	systray.AddSeparator()

	// Quit
	mQuit := systray.AddMenuItem("Quit", "Exit the application")

	// Set status callback
	st.orchestrator.SetStatusCallback(func(status string) {
		st.updateStatus(status)
	})

	// Start orchestrator in background
	go st.orchestrator.Start()

	// Status update ticker
	go st.statusUpdateLoop()

	// Handle menu events
	go st.handleMenuEvents(mConfig, mQuit)
}

// handleMenuEvents handles clicks on menu items
func (st *SystemTray) handleMenuEvents(mConfig, mQuit *systray.MenuItem) {
	for {
		select {
		case <-st.clipboardItem.ClickedCh:
			st.toggleClipboard()

		case <-st.offsetItems[-2000].ClickedCh:
			st.setOffset(-2000 * time.Millisecond)
		case <-st.offsetItems[-1000].ClickedCh:
			st.setOffset(-1000 * time.Millisecond)
		case <-st.offsetItems[-500].ClickedCh:
			st.setOffset(-500 * time.Millisecond)
		case <-st.offsetItems[0].ClickedCh:
			st.setOffset(0)
		case <-st.offsetItems[500].ClickedCh:
			st.setOffset(500 * time.Millisecond)
		case <-st.offsetItems[1000].ClickedCh:
			st.setOffset(1000 * time.Millisecond)
		case <-st.offsetItems[2000].ClickedCh:
			st.setOffset(2000 * time.Millisecond)

		case <-mConfig.ClickedCh:
			st.openConfig()

		case <-mQuit.ClickedCh:
			log.Println("Quit requested from system tray")
			systray.Quit()
			return
		}
	}
}

// toggleClipboard toggles clipboard updates
func (st *SystemTray) toggleClipboard() {
	if st.clipboardItem.Checked() {
		st.clipboardItem.Uncheck()
		st.clipboardItem.SetTitle("☐ Clipboard Updates")
		st.orchestrator.SetUpdateClipboard(false)
	} else {
		st.clipboardItem.Check()
		st.clipboardItem.SetTitle("✓ Clipboard Updates")
		st.orchestrator.SetUpdateClipboard(true)
	}
}

// setOffset sets the lyric offset
func (st *SystemTray) setOffset(offset time.Duration) {
	// Uncheck previous offset
	for _, item := range st.offsetItems {
		item.Uncheck()
	}

	// Check new offset
	offsetMs := int(offset.Milliseconds())
	if item, ok := st.offsetItems[offsetMs]; ok {
		item.Check()
	}

	st.currentOffset = offset
	st.orchestrator.SetLyricOffset(offset)
}

// updateStatus updates the status display
func (st *SystemTray) updateStatus(status string) {
	if len(status) > 60 {
		status = status[:60] + "..."
	}
	st.statusItem.SetTitle(fmt.Sprintf("🎵 %s", status))
}

// statusUpdateLoop periodically updates the status
func (st *SystemTray) statusUpdateLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		status := st.orchestrator.GetCurrentStatus()
		st.updateStatus(status)
	}
}

// openConfig opens the configuration file in the default editor
func (st *SystemTray) openConfig() {
	// This would ideally open the config file in the system's default editor
	// For now, just log the location
	log.Println("Config file location: ~/.config/lyric-clipboard/config.json")
	// TODO: Implement platform-specific file opening
}

// onExit is called when the system tray is exiting
func (st *SystemTray) onExit() {
	log.Println("System tray exiting...")
	st.orchestrator.Stop()
}
