package main

import (
	"embed"
	"fmt"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

//go:embed resources
var resources embed.FS

type Stop struct {
	ID        string
	Name      string
	Timestamp time.Time
	Duration  time.Duration
}

type StopManager struct {
	stops   []Stop
	current time.Time
	app     fyne.App
	prefs   fyne.Preferences
}

func NewStopManager() *StopManager {
	return &StopManager{
		stops:   make([]Stop, 0),
		current: time.Now(),
	}
}

func (sm *StopManager) Init() {
	sm.app = app.New()
	sm.prefs = sm.app.Preferences()

	// Load saved stops
	if savedStops := sm.prefs.StringList("stops"); len(savedStops) > 0 {
		for _, s := range savedStops {
			// Parse stop data from preferences
			parts := splitStopData(s)
			if len(parts) == 3 {
				sm.stops = append(sm.stops, Stop{
					ID:        parts[0],
					Name:      parts[1],
					Timestamp: parseTime(parts[2]),
				})
			}
		}
	}
}

func (sm *StopManager) CreateUI() fyne.CanvasObject {
	sm.Init()

	title := widget.NewLabelWithStyle("Stop Manager", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.Wrapping = fyne.TextWrapWord

	startBtn := widget.NewButton("Start Stop", func() {
		sm.startStop()
	})

	stopBtn := widget.NewButton("Stop Stop", func() {
		sm.stopStop()
	})

	saveBtn := widget.NewButton("Save Stops", func() {
		sm.saveStops()
	})

	loadBtn := widget.NewButton("Load Stops", func() {
		sm.loadStops()
	})

	stopList := sm.createStopList()

	actions := container.NewHBox(startBtn, stopBtn, saveBtn, loadBtn)

	content := container.NewVBox(
		title,
		actions,
		widget.NewSeparator(),
		stopList,
	)

	return container.NewScroll(content)
}

func (sm *StopManager) startStop() {
	now := time.Now()
	stop := Stop{
		ID:        fmt.Sprintf("stop_%d", now.UnixNano()),
		Name:      fmt.Sprintf("Stop %d", len(sm.stops)+1),
		Timestamp: now,
	}

	sm.stops = append(sm.stops, stop)
	sm.current = now

	// Update UI in main thread
	fyne.Do(func() {
		// Refresh the list
		sm.updateStopList()
	})
}

func (sm *StopManager) stopStop() {
	if len(sm.stops) == 0 {
		return
	}

	// Update the last stop's duration
	lastStop := &sm.stops[len(sm.stops)-1]
	lastStop.Duration = time.Since(lastStop.Timestamp)

	// Update UI in main thread
	fyne.Do(func() {
		// Refresh the list
		sm.updateStopList()
	})
}

func (sm *StopManager) saveStops() {
	// Save stops to preferences
	var stopData []string
	for _, stop := range sm.stops {
		data := fmt.Sprintf("%s|%s|%d", stop.ID, stop.Name, stop.Timestamp.UnixNano())
		stopData = append(stopData, data)
	}

	sm.prefs.SetStringList("stops", stopData)

	// Show confirmation dialog
	fyne.Do(func() {
		dialog.ShowInformation("Saved", "Stops saved successfully!", sm.app.NewWindow())
	})
}

func (sm *StopManager) loadStops() {
	// Load stops from preferences
	if savedStops := sm.prefs.StringList("stops"); len(savedStops) > 0 {
		sm.stops = make([]Stop, 0)
		for _, s := range savedStops {
			parts := splitStopData(s)
			if len(parts) == 3 {
				sm.stops = append(sm.stops, Stop{
					ID:        parts[0],
					Name:      parts[1],
					Timestamp: parseTime(parts[2]),
				})
			}
		}

		// Update UI in main thread
		fyne.Do(func() {
			sm.updateStopList()
		})
	}
}

func (sm *StopManager) createStopList() *widget.List {
	return widget.NewList(
		func() int {
			return len(sm.stops)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("Template")
		},
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id < len(sm.stops) {
				stop := sm.stops[id]
				obj.(*widget.Label).SetText(fmt.Sprintf("%s: %s (%v)", stop.Name, stop.Timestamp.Format("2006-01-02 15:04:05"), stop.Duration))
			}
		},
	)
}

func (sm *StopManager) updateStopList() {
	// This method would be called to refresh the stop list display
	// In a real implementation, you'd update the list widget
}

func splitStopData(data string) []string {
	parts := make([]string, 0)
	current := ""
	inQuotes := false

	for _, char := range data {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case '|':
			if !inQuotes {
				parts = append(parts, current)
				current = ""
				continue
			}
			fallthrough
		default:
			current += string(char)
		}
	}

	if current != "" {
		parts = append(parts, current)
	}

	return parts
}

func parseTime(unixNano string) time.Time {
	if unixNano == "" {
		return time.Now()
	}

	var unix int64
	fmt.Sscanf(unixNano, "%d", &unix)
	return time.Unix(0, unix)
}

func main() {
	sm := NewStopManager()

	w := sm.app.NewWindow("Stop Manager")
	w.SetContent(sm.CreateUI())
	w.ShowAndRun()
}
