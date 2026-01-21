package progress

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// Manager manages multiple progress bars for concurrent operations
type Manager struct {
	bars    map[string]*progressbar.ProgressBar
	mu      sync.Mutex
	enabled bool
}

// NewManager creates a new progress manager
func NewManager(enabled bool) *Manager {
	return &Manager{
		bars:    make(map[string]*progressbar.ProgressBar),
		enabled: enabled,
	}
}

// AddBar adds a new progress bar for an operation
func (m *Manager) AddBar(name string, total int64) *progressbar.ProgressBar {
	if !m.enabled {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(name),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	m.bars[name] = bar
	return bar
}

// AddBarWithSpeed adds a progress bar that shows transfer speed
func (m *Manager) AddBarWithSpeed(name string, total int64) *progressbar.ProgressBar {
	if !m.enabled {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	bar := progressbar.NewOptions64(
		total,
		progressbar.OptionSetDescription(name),
		progressbar.OptionSetWidth(50),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
	)

	m.bars[name] = bar
	return bar
}

// UpdateBar updates a progress bar
func (m *Manager) UpdateBar(name string, current int64) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, exists := m.bars[name]; exists {
		bar.Set64(current)
	}
}

// IncrementBar increments a progress bar by the specified amount
func (m *Manager) IncrementBar(name string, amount int64) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, exists := m.bars[name]; exists {
		bar.Add64(amount)
	}
}

// FinishBar marks a progress bar as complete
func (m *Manager) FinishBar(name string) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, exists := m.bars[name]; exists {
		bar.Finish()
		delete(m.bars, name)
	}
}

// RemoveBar removes a progress bar
func (m *Manager) RemoveBar(name string) {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if bar, exists := m.bars[name]; exists {
		bar.Finish()
		delete(m.bars, name)
	}
}

// Close closes all progress bars
func (m *Manager) Close() {
	if !m.enabled {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for name, bar := range m.bars {
		bar.Finish()
		delete(m.bars, name)
	}
}

// IsEnabled returns whether progress bars are enabled
func (m *Manager) IsEnabled() bool {
	return m.enabled
}

// IsTTY checks if the output is a terminal
func IsTTY() bool {
	return isTerminal(os.Stderr)
}

// isTerminal checks if a file descriptor is a terminal
func isTerminal(f *os.File) bool {
	return term.IsTerminal(int(f.Fd()))
}

// Writer wraps an io.Writer to update progress while writing
type Writer struct {
	writer io.Writer
	bar    *progressbar.ProgressBar
}

// NewWriter creates a new progress writer
func NewWriter(writer io.Writer, bar *progressbar.ProgressBar) *Writer {
	return &Writer{
		writer: writer,
		bar:    bar,
	}
}

// Write implements io.Writer
func (w *Writer) Write(p []byte) (n int, err error) {
	n, err = w.writer.Write(p)
	if w.bar != nil {
		w.bar.Add64(int64(n))
	}
	return n, err
}
