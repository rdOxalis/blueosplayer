# 🎵 BlueOS Player

A powerful BlueOS preset player for the command line with multi-language support and automatic network scanning.

## ✨ Features

- 🔍 **Automatic Network Scanning** - Finds all BlueOS players on your network
- 🌍 **Multi-Language Support** - English, German, and Swahili
- 🎮 **Interactive Control** - Simple command-line interface
- 📱 **Multiple Player Support** - Choose from detected players
- 🎵 **Full Playback Control** - Play, pause, stop, volume, and preset management

## 📦 Installation

1. **Clone this repository:**
   ```bash
   git clone <repository-url>
   cd blueosplayer
   ```

2. **Build the application:**
   ```bash
   go build blueosplayer.go
   ```

## 🚀 Usage

1. **Start the application:**
   ```bash
   ./blueosplayer
   ```

2. **Select a player:**  
   The app will automatically scan your network and show available BlueOS players:
   ```
   📱 Available Players:
     [1] Living Room Speaker (Bluesound Node) - 192.168.1.100
     [2] Kitchen Speaker (Bluesound Pulse) - 192.168.1.101
   
   Select a player (1-2): 1
   ```

3. **Use interactive commands:**  
   Once connected, you can control your player with simple commands.

## 🎮 Available Commands

| Command | Description |
|---------|-------------|
| `play <preset_id>` | Play a specific preset |
| `play` | Start/resume playback |
| `pause` | Pause playback |
| `stop` | Stop playback |
| `next` | Skip to next track |
| `prev` | Go to previous track |
| `volume <0-100>` | Set volume level |
| `vol <0-100>` | Set volume (short command) |
| `status` | Show current player status |
| `presets` | List all available presets |
| `help` | Show command help |
| `lang <en\|de\|sw>` | Change interface language |
| `quit` / `exit` | Exit the application |

## 🌍 Language Support

Switch between languages anytime during operation:

| Command | Language | Example Output |
|---------|----------|----------------|
| `lang en` | 🇺🇸 English | "✅ Connected to: Living Room Speaker" |
| `lang de` | 🇩🇪 Deutsch | "✅ Verbunden mit: Wohnzimmer Lautsprecher" |
| `lang sw` | 🇹🇿 Kiswahili | "✅ Imeunganishwa na: Kichezaji cha Sebuleni" |

## 📝 Example Session

```
🎵 BlueOS Controller
====================
🔍 Scanning network for BlueOS players...
   Scanning network: 192.168.1
   ✅ Found: Living Room (Node 2i) at 192.168.1.100

📱 Available Players:
  [1] Living Room (Bluesound Node 2i) - 192.168.1.100

Select a player (1-1): 1
✅ Connected to: Living Room (192.168.1.100)

🎵 BlueOS Controller - Interactive Mode
=======================================

📊 Status: stop | Volume: 50%

📋 Available Presets:
  [1] Spotify Daily Mix
  [2] Radio Paradise
  [3] Classical WQXR

🎮 Available Commands:
  play <preset_id>  - Play preset
  play              - Start playback
  ...

Blueos> play 1
✅ Playing preset 1

📊 Status: stream | Volume: 50%
🎵 Great Song - Amazing Artist (Awesome Album)

Blueos> lang de
🌍 Sprache geändert zu Deutsch

Blueos> help
🎮 Verfügbare Befehle:
  play <preset_id>  - Preset abspielen
  ...

Blueos> quit
👋 Auf Wiedersehen!
```

## 🔧 Requirements

- Go 1.19 or higher
- BlueOS-compatible device on the same network
- Network access to scan for devices

## 🤝 Contributing

Feel free to submit issues, feature requests, or pull requests to improve this tool!

## 📄 License

This project is open source. Check the LICENSE file for details.