# Lyric Clipboard

A cross-platform application that monitors currently playing music and automatically copies synchronized lyrics to the system clipboard in real-time.

## Features

- 🎵 **Automatic music detection** - Detects currently playing songs from Spotify, VLC, Chrome, and other media players
- 📝 **Real-time lyrics sync** - Fetches synced lyrics from [lrclib.net](https://lrclib.net) and updates your clipboard as the song progresses
- 🖥️ **Cross-platform** - Supports Linux and Windows
- ⚡ **Fast and lightweight** - Written in Go with minimal resource usage
- 🎭 **Demo mode** - Test the app without a media player

## Platform Support

| Platform | Status | Detection Method |
|----------|--------|------------------|
| Linux | ✅ Full support | D-Bus MPRIS |
| Windows | ✅ Full support | Windows Media Transport Controls (PowerShell) |
| macOS | ❌ Not yet supported | Planned |

## Installation

### Prerequisites

**Linux:**
- Go 1.19 or later
- D-Bus (usually pre-installed)
- Optional: `xclip` for better clipboard support in WSL

**Windows:**
- Go 1.19 or later
- PowerShell (pre-installed on Windows 10/11)

### Building from Source

```bash
# Clone the repository
git clone https://github.com/arnavpraneet/lyric-clipboard-app.git
cd lyric-clipboard-app

# Install dependencies
go mod download

# Build for your platform
go build -o lyric-clipboard ./cmd/lyric-clipboard

# Or for Windows
go build -o lyric-clipboard.exe ./cmd/lyric-clipboard
```

### Cross-compilation

**Build for Windows from Linux/macOS:**
```bash
GOOS=windows GOARCH=amd64 go build -o lyric-clipboard.exe ./cmd/lyric-clipboard
```

**Build for Linux from Windows:**
```bash
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -o lyric-clipboard ./cmd/lyric-clipboard
```

## Usage

### Basic Usage

Simply run the application while a media player is active:

```bash
# Linux
./lyric-clipboard

# Windows
.\lyric-clipboard.exe
```

The app will:
1. Detect your currently playing song
2. Fetch synchronized lyrics from lrclib.net
3. Update your clipboard with the current lyric line as the song progresses

### Demo Mode

Test the application without a media player:

```bash
# Use default song (Rick Astley - Never Gonna Give You Up)
./lyric-clipboard -demo

# Specify a custom song
./lyric-clipboard -demo -artist "Pink Floyd" -title "Comfortably Numb"
```

### Stopping the Application

Press `Ctrl+C` to gracefully shut down the application.

## How It Works

```
┌─────────────┐      ┌──────────┐      ┌────────────┐      ┌───────────┐
│ Media Player│─────▶│ Detector │─────▶│Orchestrator│─────▶│ Clipboard │
│  (Spotify)  │ D-Bus│  (Linux) │ Poll │            │Update│           │
└─────────────┘  or  │ (Windows)│ Loop │            │      └───────────┘
                 API  └──────────┘      │            │
                                        │            │
                                        ▼            ▼
                                   ┌────────┐   ┌────────┐
                                   │ Lyrics │   │ Parser │
                                   │Fetcher │◀──│  (LRC) │
                                   └────────┘   └────────┘
                                        │
                                        ▼
                                   ┌────────┐
                                   │ Cache  │
                                   └────────┘
```

### Components

1. **Detector** - Platform-specific music detection
   - Linux: Uses D-Bus MPRIS to communicate with media players
   - Windows: Uses PowerShell to access Windows Media Transport Controls

2. **Orchestrator** - Coordinates all components with a polling loop (default: 300ms)

3. **Lyrics Fetcher** - Fetches synced lyrics from lrclib.net with in-memory caching

4. **Parser** - Parses LRC format lyrics and synchronizes with playback position

5. **Clipboard Manager** - Cross-platform clipboard operations

## Configuration

Currently, configuration is done through the code. Future versions will support a configuration file.

**Polling interval** can be modified in [cmd/lyric-clipboard/main.go](cmd/lyric-clipboard/main.go:23):

```go
config := orchestrator.Config{
    PollInterval: 300 * time.Millisecond, // Adjust this value
}
```

## Supported Media Players

### Linux (via D-Bus MPRIS)
- Spotify
- VLC
- Rhythmbox
- Chromium/Chrome
- Any MPRIS-compatible player

### Windows (via Media Transport Controls)
- Spotify
- VLC
- Windows Media Player
- Groove Music
- Chrome/Edge
- Any app that implements Windows Media Transport Controls

## Troubleshooting

### Linux

**No song detected:**
- Ensure your media player is running and playing music
- Check if your player supports MPRIS: `dbus-send --print-reply --dest=org.freedesktop.DBus /org/freedesktop/DBus org.freedesktop.DBus.ListNames`

**Clipboard not working in WSL:**
- Install `xclip`: `sudo apt-get install xclip`

### Windows

**No song detected:**
- Ensure PowerShell execution is enabled
- Verify your media player is playing music and supports Windows Media Transport Controls
- Try running PowerShell as administrator

**Access denied errors:**
- The app cannot run as SYSTEM or as a Windows Service (Windows API limitation)
- Run it as a regular user application

## Development

### Project Structure

```
lyric-clipboard-app/
├── cmd/
│   └── lyric-clipboard/    # Main application entry point
│       └── main.go
├── internal/
│   ├── detector/           # Platform-specific music detection
│   │   ├── detector.go              # Interface definition
│   │   ├── detector_linux.go        # Linux (D-Bus) implementation
│   │   ├── detector_windows.go      # Windows (PowerShell) implementation
│   │   ├── detector_demo.go         # Demo mode implementation
│   │   └── detector_stub.go         # Unsupported platforms
│   ├── orchestrator/       # Main coordinator
│   │   └── orchestrator.go
│   ├── lyrics/            # Lyrics fetching and parsing
│   │   ├── fetcher.go
│   │   └── parser.go
│   └── clipboard/         # Clipboard management
│       └── clipboard.go
├── CLAUDE.md             # AI assistant instructions
├── README.md            # This file
└── go.mod
```

### Running Tests

```bash
go test ./...
```

### Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Future Enhancements

- [ ] macOS support
- [ ] Configuration file support
- [ ] System tray icon and GUI
- [ ] Systemd service for Linux
- [ ] Windows Service support (if API limitations can be overcome)
- [ ] Lyric offset adjustment
- [ ] Multiple lyrics source support
- [ ] Translation support
- [ ] Unit tests
- [ ] CI/CD pipeline

## License

[Add your license here]

## Acknowledgments

- Lyrics provided by [lrclib.net](https://lrclib.net)
- Uses [D-Bus](https://www.freedesktop.org/wiki/Software/dbus/) for Linux media detection
- Uses Windows Media Transport Controls for Windows detection
