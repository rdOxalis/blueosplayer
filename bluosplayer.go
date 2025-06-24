package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BluesoundPort = "11000"
	SonosPort     = "1400"
	ScanTimeout   = 3 * time.Second
)

// Device type enumeration
type DeviceType string

const (
	DeviceTypeBluOS DeviceType = "bluos"
	DeviceTypeSonos DeviceType = "sonos"
)

// Language support
type Language string

const (
	LangEnglish Language = "en"
	LangGerman  Language = "de"
	LangSwahili Language = "sw"
)

var currentLanguage = LangEnglish

// Localization texts
var texts = map[Language]map[string]string{
	LangEnglish: {
		"title":                   "🎵 Multi-Room Audio Controller",
		"scanning":                "🔍 Scanning network for audio players...",
		"scanning_network":        "   Scanning network: %s",
		"scanning_interface":      "   Interface %s: %s",
		"found_player":            "   ✅ Found: %s (%s) at %s",
		"no_players":              "no audio players found",
		"could_not_determine_ip":  "could not determine local IP: %w",
		"available_players":       "📱 Available Players:",
		"select_player":           "Select a player (1-%d): ",
		"invalid_selection":       "❌ Invalid selection",
		"connected_to":            "✅ Connected to: %s (%s)",
		"error_selecting_player":  "Error selecting player: %v",
		"interactive_mode":        "🎵 Multi-Room Audio Controller - Interactive Mode",
		"separator":               "=======================================",
		"status_volume":           "📊 Status: %s | Volume: %s",
		"volume_unknown":          "N/A",
		"error_retrieving_status": "❌ Error retrieving status",
		"available_presets":       "📋 Available Presets/Favorites:",
		"error_loading_presets":   "❌ Error loading presets/favorites",
		"available_commands":      "🎮 Available Commands:",
		"cmd_play_preset":         "play <id>   - Play preset/favorite",
		"cmd_play":                "play       - Start playback",
		"cmd_pause":               "pause      - Pause playback",
		"cmd_stop":                "stop       - Stop playback",
		"cmd_next":                "next       - Next track",
		"cmd_prev":                "prev       - Previous track",
		"cmd_volume":              "vol <0-100> - Set volume",
		"cmd_status":              "status     - Refresh status",
		"cmd_presets":             "presets    - Refresh presets/favorites",
		"cmd_help":                "help       - Show help",
		"cmd_lang":                "lang <en|de|sw> - Change language",
		"cmd_output":              "output <id> - Switch to player",
		"cmd_group":               "group <id1+id2> - Group players",
		"cmd_ungroup":             "ungroup - Remove all groups",
		"cmd_debug":               "debug - Show API endpoints",
		"cmd_quit":                "quit/exit  - Exit program",
		"prompt":                  "Command> ",
		"invalid_preset_id":       "❌ Invalid preset/favorite ID",
		"error_playing_preset":    "❌ Error playing preset/favorite",
		"playing_preset":          "✅ Playing preset/favorite %d",
		"error_starting_playback": "❌ Error starting playback",
		"playback_started":        "▶️ Playback started",
		"error_pausing":           "❌ Error pausing",
		"paused":                  "⏸️ Paused",
		"error_stopping":          "❌ Error stopping",
		"stopped":                 "⏹️ Stopped",
		"error_next_track":        "❌ Error skipping to next track",
		"next_track":              "⏭️ Next track",
		"error_prev_track":        "❌ Error going to previous track",
		"prev_track":              "⏮️ Previous track",
		"volume_missing":          "❌ Volume value missing",
		"invalid_volume":          "❌ Invalid volume value",
		"error_setting_volume":    "❌ Error setting volume",
		"volume_set":              "🔊 Volume set to %d%%",
		"language_changed":        "🌍 Language changed to",
		"invalid_language":        "❌ Invalid language. Use: en, de, sw",
		"goodbye":                 "👋 Goodbye!",
		"unknown_command":         "❌ Unknown command: %s (Type 'help' for help)",
		"last_action":             "Last Action:",
		"no_song_playing":         "No song playing",
		"available_outputs":       "📱 Available Players:",
		"current_player":          "Current Player:",
		"switched_to_player":      "🔄 Switched to player %d: %s",
		"invalid_player_id":       "❌ Invalid player ID",
		"error_switching_player":  "❌ Error switching to player",
		"grouped_players":         "🔗 Grouped players: %s as master",
		"invalid_group_format":    "❌ Invalid group format. Use: group <id1+id2>",
		"error_grouping":          "❌ Error grouping players",
		"group_combinations":      "🎵 Group Combinations:",
		"ungrouped_all":           "🔓 All player groups removed",
		"error_ungrouping":        "❌ Error removing groups",
		"scanning_interfaces":     "🔍 Found %d network interfaces to scan",
		"completed_scan":          "✅ Completed scanning %d networks",
	},
	LangGerman: {
		"title":                   "🎵 Multi-Room Audio Controller",
		"scanning":                "🔍 Suche nach Audio-Playern im Netzwerk...",
		"scanning_network":        "   Scanne Netzwerk: %s",
		"scanning_interface":      "   Interface %s: %s",
		"found_player":            "   ✅ Gefunden: %s (%s) auf %s",
		"no_players":              "keine Audio-Player gefunden",
		"could_not_determine_ip":  "konnte lokale IP nicht ermitteln: %w",
		"available_players":       "📱 Verfügbare Player:",
		"select_player":           "Wähle einen Player (1-%d): ",
		"invalid_selection":       "❌ Ungültige Auswahl",
		"connected_to":            "✅ Verbunden mit: %s (%s)",
		"error_selecting_player":  "Fehler bei der Player-Auswahl: %v",
		"interactive_mode":        "🎵 Multi-Room Audio Controller - Interaktiver Modus",
		"separator":               "==========================================",
		"status_volume":           "📊 Status: %s | Lautstärke: %s",
		"volume_unknown":          "N/A",
		"error_retrieving_status": "❌ Fehler beim Abrufen des Status",
		"available_presets":       "📋 Verfügbare Presets/Favoriten:",
		"error_loading_presets":   "❌ Fehler beim Laden der Presets/Favoriten",
		"available_commands":      "🎮 Verfügbare Befehle:",
		"cmd_play_preset":         "play <id>   - Preset/Favorit abspielen",
		"cmd_play":                "play       - Wiedergabe starten",
		"cmd_pause":               "pause      - Pausieren",
		"cmd_stop":                "stop       - Stoppen",
		"cmd_next":                "next       - Nächster Titel",
		"cmd_prev":                "prev       - Vorheriger Titel",
		"cmd_volume":              "vol <0-100> - Lautstärke setzen",
		"cmd_status":              "status     - Status aktualisieren",
		"cmd_presets":             "presets    - Presets/Favoriten aktualisieren",
		"cmd_help":                "help       - Hilfe anzeigen",
		"cmd_lang":                "lang <en|de|sw> - Sprache ändern",
		"cmd_output":              "output <id> - Zu Player wechseln",
		"cmd_group":               "group <id1+id2> - Player gruppieren",
		"cmd_ungroup":             "ungroup - Alle Gruppen auflösen",
		"cmd_debug":               "debug - API-Endpunkte anzeigen",
		"cmd_quit":                "quit/exit  - Programm beenden",
		"prompt":                  "Befehl> ",
		"invalid_preset_id":       "❌ Ungültige Preset/Favoriten-ID",
		"error_playing_preset":    "❌ Fehler beim Abspielen",
		"playing_preset":          "✅ Preset/Favorit %d wird abgespielt",
		"error_starting_playback": "❌ Fehler beim Starten",
		"playback_started":        "▶️ Wiedergabe gestartet",
		"error_pausing":           "❌ Fehler beim Pausieren",
		"paused":                  "⏸️ Pausiert",
		"error_stopping":          "❌ Fehler beim Stoppen",
		"stopped":                 "⏹️ Gestoppt",
		"error_next_track":        "❌ Fehler beim Weiterschalten",
		"next_track":              "⏭️ Nächster Titel",
		"error_prev_track":        "❌ Fehler beim Zurückschalten",
		"prev_track":              "⏮️ Vorheriger Titel",
		"volume_missing":          "❌ Lautstärke-Wert fehlt",
		"invalid_volume":          "❌ Ungültiger Lautstärke-Wert",
		"error_setting_volume":    "❌ Fehler beim Setzen der Lautstärke",
		"volume_set":              "🔊 Lautstärke auf %d%% gesetzt",
		"language_changed":        "🌍 Sprache geändert zu",
		"invalid_language":        "❌ Ungültige Sprache. Verwende: en, de, sw",
		"goodbye":                 "👋 Auf Wiedersehen!",
		"unknown_command":         "❌ Unbekannter Befehl: %s (Tippe 'help' für Hilfe)",
		"last_action":             "Letzte Aktion:",
		"no_song_playing":         "Kein Lied wird abgespielt",
		"available_outputs":       "📱 Verfügbare Player:",
		"current_player":          "Aktueller Player:",
		"switched_to_player":      "🔄 Gewechselt zu Player %d: %s",
		"invalid_player_id":       "❌ Ungültige Player-ID",
		"error_switching_player":  "❌ Fehler beim Wechseln des Players",
		"grouped_players":         "🔗 Player gruppiert: %s als Master",
		"invalid_group_format":    "❌ Ungültiges Gruppen-Format. Verwende: group <id1+id2>",
		"error_grouping":          "❌ Fehler beim Gruppieren",
		"group_combinations":      "🎵 Gruppen-Kombinationen:",
		"ungrouped_all":           "🔓 Alle Player-Gruppen aufgelöst",
		"error_ungrouping":        "❌ Fehler beim Auflösen der Gruppen",
		"scanning_interfaces":     "🔍 %d Netzwerkschnittstellen gefunden zum Scannen",
		"completed_scan":          "✅ Scannen von %d Netzwerken abgeschlossen",
	},
	LangSwahili: {
		"title":                   "🎵 Kidhibiti cha Audio ya Multi-Room",
		"scanning":                "🔍 Kutafuta vichezaji vya audio kwenye mtandao...",
		"scanning_network":        "   Kutafuta mtandao: %s",
		"scanning_interface":      "   Interface %s: %s",
		"found_player":            "   ✅ Kumepatikana: %s (%s) kwa %s",
		"no_players":              "hakuna vichezaji vya audio vilivopatikana",
		"could_not_determine_ip":  "haikuweza kutambua IP ya ndani: %w",
		"available_players":       "📱 Vichezaji Vinavyopatikana:",
		"select_player":           "Chagua kichezaji (1-%d): ",
		"invalid_selection":       "❌ Chaguo batili",
		"connected_to":            "✅ Imeunganishwa na: %s (%s)",
		"error_selecting_player":  "Hitilafu katika kuchagua kichezaji: %v",
		"interactive_mode":        "🎵 Kidhibiti cha Audio ya Multi-Room - Hali ya Maingiliano",
		"separator":               "===========================================",
		"status_volume":           "📊 Hali: %s | Sauti: %s",
		"volume_unknown":          "N/A",
		"error_retrieving_status": "❌ Hitilafu katika kupata hali",
		"available_presets":       "📋 Mipangilio/Vipendwa Vinavyopatikana:",
		"error_loading_presets":   "❌ Hitilafu katika kupakia mipangilio/vipendwa",
		"available_commands":      "🎮 Amri Zinazopatikana:",
		"cmd_play_preset":         "play <id>   - Cheza mpangilio/kipendwa",
		"cmd_play":                "play       - Anza kucheza",
		"cmd_pause":               "pause      - Simamisha",
		"cmd_stop":                "stop       - Acha",
		"cmd_next":                "next       - Wimbo ujao",
		"cmd_prev":                "prev       - Wimbo uliopita",
		"cmd_volume":              "vol <0-100> - Weka sauti",
		"cmd_status":              "status     - Onyesha hali",
		"cmd_presets":             "presets    - Onyesha mipangilio/vipendwa",
		"cmd_help":                "help       - Onyesha msaada",
		"cmd_lang":                "lang <en|de|sw> - Badilisha lugha",
		"cmd_output":              "output <id> - Badili kichezaji",
		"cmd_group":               "group <id1+id2> - Unganisha vichezaji",
		"cmd_ungroup":             "ungroup - Ondoa vikundi vyote",
		"cmd_debug":               "debug - Onyesha API endpoints",
		"cmd_quit":                "quit/exit  - Toka programu",
		"prompt":                  "Amri> ",
		"invalid_preset_id":       "❌ Kitambulisho cha mpangilio/kipendwa si halali",
		"error_playing_preset":    "❌ Hitilafu katika kucheza mpangilio/kipendwa",
		"playing_preset":          "✅ Kucheza mpangilio/kipendwa %d",
		"error_starting_playback": "❌ Hitilafu katika kuanza kucheza",
		"playback_started":        "▶️ Imeanza kucheza",
		"error_pausing":           "❌ Hitilafu katika kusimamisha",
		"paused":                  "⏸️ Imesimamishwa",
		"error_stopping":          "❌ Hitilafu katika kuacha",
		"stopped":                 "⏹️ Imeachwa",
		"error_next_track":        "❌ Hitilafu katika kuruka wimbo ujao",
		"next_track":              "⏭️ Wimbo ujao",
		"error_prev_track":        "❌ Hitilafu katika kurudi wimbo uliopita",
		"prev_track":              "⏮️ Wimbo uliopita",
		"volume_missing":          "❌ Thamani ya sauti inakosekana",
		"invalid_volume":          "❌ Thamani ya sauti si halali",
		"error_setting_volume":    "❌ Hitilafu katika kuweka sauti",
		"volume_set":              "🔊 Sauti imewekwa %d%%",
		"language_changed":        "🌍 Lugha imebadilishwa kuwa",
		"invalid_language":        "❌ Lugha si halali. Tumia: en, de, sw",
		"goodbye":                 "👋 Kwaheri!",
		"unknown_command":         "❌ Amri isiyojulikana: %s (Andika 'help' kwa msaada)",
		"last_action":             "Kitendo cha Mwisho:",
		"no_song_playing":         "Hakuna wimbo unaochezwa",
		"available_outputs":       "📱 Vichezaji Vinavyopatikana:",
		"current_player":          "Kichezaji cha Sasa:",
		"switched_to_player":      "🔄 Imebadilishwa kwa kichezaji %d: %s",
		"invalid_player_id":       "❌ Kitambulisho cha kichezaji si halali",
		"error_switching_player":  "❌ Hitilafu katika kubadili kichezaji",
		"grouped_players":         "🔗 Vichezaji vimeunganishwa: %s kama mkuu",
		"invalid_group_format":    "❌ Muundo wa kikundi si halali. Tumia: group <id1+id2>",
		"error_grouping":          "❌ Hitilafu katika kuunganisha",
		"group_combinations":      "🎵 Miunganiko ya Vikundi:",
		"ungrouped_all":           "🔓 Vikundi vyote vya vichezaji vimeondolewa",
		"error_ungrouping":        "❌ Hitilafu katika kuondoa vikundi",
		"scanning_interfaces":     "🔍 Kumepatikana %d network interfaces za kutafuta",
		"completed_scan":          "✅ Imemaliza kutafuta %d mitandao",
	},
}

// Helper function to get localized text
func getText(key string) string {
	if text, exists := texts[currentLanguage][key]; exists {
		return text
	}
	// Fallback to English if key not found
	if text, exists := texts[LangEnglish][key]; exists {
		return text
	}
	return key // Return key as fallback
}

// Structures for BluOS XML parsing
type Presets struct {
	XMLName xml.Name `xml:"presets"`
	Presets []Preset `xml:"preset"`
}

type Preset struct {
	ID    int    `xml:"id,attr"`
	Name  string `xml:"name,attr"`
	URL   string `xml:"url,attr"`
	Image string `xml:"image,attr"`
}

type Status struct {
	XMLName xml.Name `xml:"status"`
	State   string   `xml:"state"`
	Song    string   `xml:"song"`
	Artist  string   `xml:"artist"`
	Album   string   `xml:"album"`
	Volume  int      `xml:"volume"`
}

type SyncStatus struct {
	XMLName xml.Name `xml:"SyncStatus"`
	Name    string   `xml:"name,attr"`
	Brand   string   `xml:"brand,attr"`
	Model   string   `xml:"model,attr"`
}

// Structures for Sonos XML parsing
type SonosGetPositionInfoResponse struct {
	XMLName xml.Name  `xml:"Envelope"`
	Body    SonosBody `xml:"Body"`
}

type SonosBody struct {
	XMLName           xml.Name                  `xml:"Body"`
	GetPositionInfo   SonosGetPositionInfoBody  `xml:"GetPositionInfoResponse"`
	GetTransportInfo  SonosGetTransportInfoBody `xml:"GetTransportInfoResponse"`
	GetVolumeResponse SonosGetVolumeBody        `xml:"GetVolumeResponse"`
	Browse            SonosBrowseBody           `xml:"BrowseResponse"`
}

type SonosGetPositionInfoBody struct {
	XMLName       xml.Name `xml:"GetPositionInfoResponse"`
	Track         string   `xml:"Track"`
	TrackMetaData string   `xml:"TrackMetaData"`
}

type SonosGetTransportInfoBody struct {
	XMLName               xml.Name `xml:"GetTransportInfoResponse"`
	CurrentTransportState string   `xml:"CurrentTransportState"`
}

type SonosGetVolumeBody struct {
	XMLName       xml.Name `xml:"GetVolumeResponse"`
	CurrentVolume string   `xml:"CurrentVolume"`
}

type SonosBrowseBody struct {
	XMLName xml.Name `xml:"BrowseResponse"`
	Result  string   `xml:"Result"`
}

// Sonos favorite item structure
type SonosFavorite struct {
	ID   int
	Name string
	URI  string
	Meta string
}

// Player info for scan results
type PlayerInfo struct {
	IP    string
	Name  string
	Brand string
	Model string
	Type  DeviceType
}

// Network interface info
type NetworkInterface struct {
	Name   string
	IP     string
	Subnet string
}

// Generic client interface
type AudioClient interface {
	GetPresets() ([]Preset, error)
	GetStatus() (*Status, error)
	PlayPreset(id int) error
	Play() error
	Pause() error
	Stop() error
	SetVolume(level int) error
	Next() error
	Previous() error
	AddSlave(slaveIP string) error
	RemoveSlave(slaveIP string) error
	RemoveAllSlaves() error
	LeaveGroup() error
	GetDeviceType() DeviceType
	DebugAPI() string
}

// BluOS API Client
type BluesoundClient struct {
	baseURL string
	client  *http.Client
}

// Sonos API Client
type SonosClient struct {
	baseURL   string
	client    *http.Client
	favorites []SonosFavorite
}

// Global state for TUI
type TUIState struct {
	client           AudioClient
	playerName       string
	status           *Status
	presets          []Preset
	lastAction       string
	statusError      string
	presetsError     string
	availablePlayers []PlayerInfo
}

var tuiState = &TUIState{}

func NewBluesoundClient(ip string) *BluesoundClient {
	return &BluesoundClient{
		baseURL: fmt.Sprintf("http://%s:%s", ip, BluesoundPort),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func NewSonosClient(ip string) *SonosClient {
	return &SonosClient{
		baseURL: fmt.Sprintf("http://%s:%s", ip, SonosPort),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		favorites: make([]SonosFavorite, 0),
	}
}

// Clear screen and move cursor to top
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// Enhanced network scanner that scans all available interfaces
func scanForPlayers() ([]PlayerInfo, error) {
	fmt.Println(getText("scanning"))

	// Get all network interfaces
	interfaces, err := getAllNetworkInterfaces()
	if err != nil {
		return nil, fmt.Errorf(getText("could_not_determine_ip"), err)
	}

	if len(interfaces) == 0 {
		return nil, fmt.Errorf("no network interfaces found")
	}

	fmt.Printf(getText("scanning_interfaces")+"\n", len(interfaces))

	var players []PlayerInfo
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Scan each network interface
	for _, iface := range interfaces {
		fmt.Printf(getText("scanning_interface")+"\n", iface.Name, iface.Subnet)

		// Scan all IPs in this subnet in parallel
		for i := 1; i < 255; i++ {
			wg.Add(1)
			go func(ip string) {
				defer wg.Done()

				// Check for BluOS player
				if player, found := checkForBluOSPlayer(ip); found {
					mu.Lock()
					// Check if we already found this player on another interface
					exists := false
					for _, existingPlayer := range players {
						if existingPlayer.IP == player.IP {
							exists = true
							break
						}
					}
					if !exists {
						players = append(players, player)
						fmt.Printf(getText("found_player")+"\n", player.Name, player.Model, player.IP)
					}
					mu.Unlock()
				}

				// Check for Sonos player
				if player, found := checkForSonosPlayer(ip); found {
					mu.Lock()
					// Check if we already found this player on another interface
					exists := false
					for _, existingPlayer := range players {
						if existingPlayer.IP == player.IP {
							exists = true
							break
						}
					}
					if !exists {
						players = append(players, player)
						fmt.Printf(getText("found_player")+"\n", player.Name, player.Model, player.IP)
					}
					mu.Unlock()
				}
			}(fmt.Sprintf("%s.%d", iface.Subnet, i))
		}
	}

	wg.Wait()
	fmt.Printf(getText("completed_scan")+"\n", len(interfaces))
	return players, nil
}

// Get all network interfaces with their subnets
func getAllNetworkInterfaces() ([]NetworkInterface, error) {
	var interfaces []NetworkInterface

	// Get all network interfaces
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		// Skip down interfaces and loopback
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		// Get addresses for this interface
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			// Only process IPv4 addresses
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					subnet := getSubnet(ipnet.IP.String())

					// Skip common virtual/internal networks unless they're the only option
					ipStr := ipnet.IP.String()
					if !isUsefulNetwork(ipStr) {
						continue
					}

					interfaces = append(interfaces, NetworkInterface{
						Name:   iface.Name,
						IP:     ipStr,
						Subnet: subnet,
					})
				}
			}
		}
	}

	// If no "useful" networks found, include all IPv4 networks
	if len(interfaces) == 0 {
		for _, iface := range ifaces {
			if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
				continue
			}

			addrs, err := iface.Addrs()
			if err != nil {
				continue
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						subnet := getSubnet(ipnet.IP.String())
						interfaces = append(interfaces, NetworkInterface{
							Name:   iface.Name,
							IP:     ipnet.IP.String(),
							Subnet: subnet,
						})
					}
				}
			}
		}
	}

	return interfaces, nil
}

// Check if this is a "useful" network (not VirtualBox NAT, etc.)
func isUsefulNetwork(ip string) bool {
	// Common home/office networks
	if strings.HasPrefix(ip, "192.168.") {
		return true
	}
	if strings.HasPrefix(ip, "10.0.0.") || strings.HasPrefix(ip, "10.0.1.") {
		return true
	}
	if strings.HasPrefix(ip, "172.16.") || strings.HasPrefix(ip, "172.17.") {
		return true
	}

	// VirtualBox NAT network (usually not where real devices are)
	if strings.HasPrefix(ip, "10.0.2.") {
		return false
	}

	// Docker networks
	if strings.HasPrefix(ip, "172.17.") {
		return false
	}

	return true
}

func getSubnet(ip string) string {
	parts := strings.Split(ip, ".")
	return strings.Join(parts[:3], ".")
}

func checkForBluOSPlayer(ip string) (PlayerInfo, bool) {
	client := &http.Client{Timeout: ScanTimeout}
	url := fmt.Sprintf("http://%s:%s/SyncStatus", ip, BluesoundPort)

	resp, err := client.Get(url)
	if err != nil {
		return PlayerInfo{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PlayerInfo{}, false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlayerInfo{}, false
	}

	var syncStatus SyncStatus
	if err := xml.Unmarshal(body, &syncStatus); err != nil {
		return PlayerInfo{}, false
	}

	return PlayerInfo{
		IP:    ip,
		Name:  syncStatus.Name,
		Brand: syncStatus.Brand,
		Model: syncStatus.Model,
		Type:  DeviceTypeBluOS,
	}, true
}

func checkForSonosPlayer(ip string) (PlayerInfo, bool) {
	client := &http.Client{Timeout: ScanTimeout}

	// Try to get device description from Sonos
	url := fmt.Sprintf("http://%s:%s/xml/device_description.xml", ip, SonosPort)
	resp, err := client.Get(url)
	if err != nil {
		return PlayerInfo{}, false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PlayerInfo{}, false
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return PlayerInfo{}, false
	}

	bodyStr := string(body)

	// Check if this is actually a Sonos device
	if !strings.Contains(bodyStr, "Sonos") && !strings.Contains(bodyStr, "RINCON") {
		return PlayerInfo{}, false
	}

	// Extract device name and model using regex
	name := "Sonos Player"
	model := "Sonos"

	// Try to extract friendly name
	if re := regexp.MustCompile(`<friendlyName>(.*?)</friendlyName>`); re != nil {
		if matches := re.FindStringSubmatch(bodyStr); len(matches) > 1 {
			name = strings.TrimSpace(matches[1])
			// Clean up the name - remove IP and RINCON part if present
			if idx := strings.Index(name, " - RINCON"); idx != -1 {
				name = strings.TrimSpace(name[:idx])
			}
			// Remove IP addresses from the name
			ipRegex := regexp.MustCompile(`\d+\.\d+\.\d+\.\d+\s*-?\s*`)
			name = ipRegex.ReplaceAllString(name, "")
			name = strings.TrimSpace(name)
		}
	}

	// Try to extract model name
	if re := regexp.MustCompile(`<modelName>(.*?)</modelName>`); re != nil {
		if matches := re.FindStringSubmatch(bodyStr); len(matches) > 1 {
			model = strings.TrimSpace(matches[1])
		}
	}

	// If name is still too complex or contains IP, use model name
	if len(name) > 50 || strings.Contains(name, ".") || name == "" {
		if model != "Sonos" && model != "" {
			name = model
		} else {
			name = fmt.Sprintf("Sonos-%s", ip[strings.LastIndex(ip, ".")+1:])
		}
	}

	return PlayerInfo{
		IP:    ip,
		Name:  name,
		Brand: "Sonos",
		Model: model,
		Type:  DeviceTypeSonos,
	}, true
}

// BluOS API methods
func (bc *BluesoundClient) makeRequest(endpoint string) ([]byte, error) {
	url := bc.baseURL + endpoint
	resp, err := bc.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return body, nil
}

func (bc *BluesoundClient) GetPresets() ([]Preset, error) {
	data, err := bc.makeRequest("/Presets")
	if err != nil {
		return nil, err
	}

	var presets Presets
	if err := xml.Unmarshal(data, &presets); err != nil {
		return nil, fmt.Errorf("failed to parse presets XML: %w", err)
	}

	return presets.Presets, nil
}

func (bc *BluesoundClient) GetStatus() (*Status, error) {
	data, err := bc.makeRequest("/Status")
	if err != nil {
		return nil, err
	}

	var status Status
	if err := xml.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse status XML: %w", err)
	}

	return &status, nil
}

func (bc *BluesoundClient) PlayPreset(id int) error {
	endpoint := fmt.Sprintf("/Preset?id=%d", id)
	_, err := bc.makeRequest(endpoint)
	return err
}

func (bc *BluesoundClient) Play() error {
	_, err := bc.makeRequest("/Play")
	return err
}

func (bc *BluesoundClient) Pause() error {
	_, err := bc.makeRequest("/Pause")
	return err
}

func (bc *BluesoundClient) Stop() error {
	_, err := bc.makeRequest("/Stop")
	return err
}

func (bc *BluesoundClient) SetVolume(level int) error {
	if level < 0 || level > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}
	endpoint := fmt.Sprintf("/Volume?level=%d", level)
	_, err := bc.makeRequest(endpoint)
	return err
}

func (bc *BluesoundClient) Next() error {
	_, err := bc.makeRequest("/Skip")
	return err
}

func (bc *BluesoundClient) Previous() error {
	_, err := bc.makeRequest("/Back")
	return err
}

func (bc *BluesoundClient) AddSlave(slaveIP string) error {
	endpoint := fmt.Sprintf("/AddSlave?slave=%s", slaveIP)
	_, err := bc.makeRequest(endpoint)
	return err
}

func (bc *BluesoundClient) RemoveSlave(slaveIP string) error {
	endpoint := fmt.Sprintf("/RemoveSlave?slave=%s", slaveIP)
	_, err := bc.makeRequest(endpoint)
	return err
}

func (bc *BluesoundClient) RemoveAllSlaves() error {
	_, err := bc.makeRequest("/RemoveAllSlaves")
	return err
}

func (bc *BluesoundClient) LeaveGroup() error {
	_, err := bc.makeRequest("/LeaveGroup")
	return err
}

func (bc *BluesoundClient) GetDeviceType() DeviceType {
	return DeviceTypeBluOS
}

func (bc *BluesoundClient) DebugAPI() string {
	endpoints := []string{"/Status", "/SyncStatus", "/Presets", "/RemoveSlave", "/AddSlave", "/Slaves"}
	var results []string
	for _, endpoint := range endpoints {
		_, err := bc.makeRequest(endpoint)
		if err != nil {
			results = append(results, fmt.Sprintf("%s: ❌", endpoint))
		} else {
			results = append(results, fmt.Sprintf("%s: ✅", endpoint))
		}
	}
	return fmt.Sprintf("BluOS API Test: %s", strings.Join(results, " | "))
}

// Sonos API methods
func (sc *SonosClient) makeSoapRequest(action, service, body string) ([]byte, error) {
	soapEnvelope := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>%s</s:Body>
</s:Envelope>`, body)

	url := fmt.Sprintf("%s/MediaRenderer/%s/Control", sc.baseURL, service)
	req, err := http.NewRequest("POST", url, strings.NewReader(soapEnvelope))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", fmt.Sprintf(`"urn:schemas-upnp-org:service:%s:1#%s"`, service, action))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(soapEnvelope)))

	resp, err := sc.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("SOAP request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return io.ReadAll(resp.Body)
}

func (sc *SonosClient) loadFavorites() error {
	if len(sc.favorites) > 0 {
		return nil // Already loaded
	}

	// Force clear cache to reload
	sc.favorites = nil

	// Try to get actual Sonos favorites using ContentDirectory with MediaServer path
	body := `<u:Browse xmlns:u="urn:schemas-upnp-org:service:ContentDirectory:1">
		<ObjectID>FV:2</ObjectID>
		<BrowseFlag>BrowseDirectChildren</BrowseFlag>
		<Filter>dc:title,res,dc:creator,upnp:artist,upnp:album</Filter>
		<StartingIndex>0</StartingIndex>
		<RequestedCount>100</RequestedCount>
		<SortCriteria></SortCriteria>
	</u:Browse>`

	soapEnvelope := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>%s</s:Body>
</s:Envelope>`, body)

	// Try MediaServer path first
	url := fmt.Sprintf("%s/MediaServer/ContentDirectory/Control", sc.baseURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(soapEnvelope))

	if err == nil {
		req.Header.Set("Content-Type", "text/xml; charset=utf-8")
		req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:ContentDirectory:1#Browse"`)
		req.Header.Set("Content-Length", fmt.Sprintf("%d", len(soapEnvelope)))

		resp, err := sc.client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			data, err := io.ReadAll(resp.Body)
			if err == nil {
				radioFavorites := sc.parseFavoritesFromResponse(string(data))
				if len(radioFavorites) > 0 {
					// Remove duplicates
					uniqueFavorites := sc.removeDuplicateFavorites(radioFavorites)
					sc.favorites = uniqueFavorites
					return nil
				}
			}
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Fallback: Create informative entries
	sc.favorites = []SonosFavorite{
		{ID: 1, Name: "[INFO] Could not load Sonos radio favorites", URI: "", Meta: ""},
		{ID: 2, Name: "[INFO] Check ContentDirectory service", URI: "", Meta: ""},
	}

	return nil
}

func (sc *SonosClient) removeDuplicateFavorites(favorites []SonosFavorite) []SonosFavorite {
	seen := make(map[string]bool)
	var unique []SonosFavorite

	for _, fav := range favorites {
		key := fav.Name + "|" + fav.URI
		if !seen[key] {
			seen[key] = true
			fav.ID = len(unique) + 1 // Re-number
			unique = append(unique, fav)
		}
	}

	return unique
}

func (sc *SonosClient) parseFavoritesFromResponse(xmlResponse string) []SonosFavorite {
	var favorites []SonosFavorite

	// Look for the Result element in the SOAP response
	resultRegex := regexp.MustCompile(`<Result>(.*?)</Result>`)
	resultMatch := resultRegex.FindStringSubmatch(xmlResponse)

	if len(resultMatch) < 2 {
		return favorites
	}

	// Decode the DIDL-Lite content
	didlContent := html.UnescapeString(resultMatch[1])

	// Parse items from DIDL-Lite
	itemRegex := regexp.MustCompile(`<item[^>]*id="([^"]*)"[^>]*>(.*?)</item>`)
	titleRegex := regexp.MustCompile(`<dc:title[^>]*>(.*?)</dc:title>`)
	resRegex := regexp.MustCompile(`<res[^>]*>(.*?)</res>`)

	items := itemRegex.FindAllStringSubmatch(didlContent, -1)

	for i, item := range items {
		if len(item) > 2 {
			// itemID := item[1]  // commented out - unused variable
			itemContent := item[2]

			var title, uri string

			if titleMatch := titleRegex.FindStringSubmatch(itemContent); len(titleMatch) > 1 {
				title = html.UnescapeString(titleMatch[1])
			}

			if resMatch := resRegex.FindStringSubmatch(itemContent); len(resMatch) > 1 {
				uri = html.UnescapeString(resMatch[1])
			}

			if title != "" {
				favorites = append(favorites, SonosFavorite{
					ID:   i + 1,
					Name: strings.TrimSpace(title),
					URI:  uri,
					Meta: itemContent,
				})
			}
		}
	}

	return favorites
}

func (sc *SonosClient) browseSonosContent(objectID, categoryName string) []SonosFavorite {
	// Browse content using ContentDirectory service
	body := fmt.Sprintf(`<u:Browse xmlns:u="urn:schemas-upnp-org:service:ContentDirectory:1">
		<ObjectID>%s</ObjectID>
		<BrowseFlag>BrowseDirectChildren</BrowseFlag>
		<Filter>*</Filter>
		<StartingIndex>0</StartingIndex>
		<RequestedCount>50</RequestedCount>
		<SortCriteria></SortCriteria>
	</u:Browse>`, objectID)

	data, err := sc.makeSoapRequest("Browse", "ContentDirectory", body)
	if err != nil {
		return []SonosFavorite{}
	}

	var response SonosGetPositionInfoResponse
	if err := xml.Unmarshal(data, &response); err != nil {
		return []SonosFavorite{}
	}

	// Parse the DIDL-Lite XML in the Result field
	resultXML := response.Body.Browse.Result

	// Extract favorites from DIDL-Lite format
	favorites := parseSonosFavorites(resultXML)

	// Add category prefix to names for clarity
	for i := range favorites {
		if categoryName != "" && len(favorites) > 0 {
			// Only add prefix if we found items and it's not the main category
			if objectID != "FV:2" {
				favorites[i].Name = fmt.Sprintf("[%s] %s", categoryName, favorites[i].Name)
			}
		}
	}

	return favorites
}

func parseSonosFavorites(didlXML string) []SonosFavorite {
	var favorites []SonosFavorite

	// Enhanced regex patterns for better DIDL-Lite parsing
	itemRegex := regexp.MustCompile(`<item[^>]*id="([^"]*)"[^>]*>(.*?)</item>`)
	titleRegex := regexp.MustCompile(`<dc:title[^>]*>(.*?)</dc:title>`)
	resRegex := regexp.MustCompile(`<res[^>]*>(.*?)</res>`)

	items := itemRegex.FindAllStringSubmatch(didlXML, -1)

	for i, item := range items {
		if len(item) > 2 {
			itemID := item[1]
			itemContent := item[2]

			var title, uri string

			if titleMatch := titleRegex.FindStringSubmatch(itemContent); len(titleMatch) > 1 {
				title = html.UnescapeString(titleMatch[1])
			}

			if resMatch := resRegex.FindStringSubmatch(itemContent); len(resMatch) > 1 {
				uri = html.UnescapeString(resMatch[1])
			}

			// Skip empty or invalid items
			if title == "" {
				continue
			}

			// Clean up title
			title = strings.TrimSpace(title)

			// Use item ID as URI fallback if no res found
			if uri == "" && itemID != "" {
				uri = itemID
			}

			favorites = append(favorites, SonosFavorite{
				ID:   i + 1,
				Name: title,
				URI:  uri,
				Meta: itemContent,
			})
		}
	}

	return favorites
}

func (sc *SonosClient) GetPresets() ([]Preset, error) {
	if err := sc.loadFavorites(); err != nil {
		return nil, err
	}

	var presets []Preset
	for _, fav := range sc.favorites {
		presets = append(presets, Preset{
			ID:   fav.ID,
			Name: fav.Name,
			URL:  fav.URI,
		})
	}

	return presets, nil
}

func (sc *SonosClient) GetStatus() (*Status, error) {
	// Get transport state
	transportBody := `<u:GetTransportInfo xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:GetTransportInfo>`

	transportData, err := sc.makeSoapRequest("GetTransportInfo", "AVTransport", transportBody)
	if err != nil {
		return &Status{
			State:  "stopped",
			Song:   "",
			Artist: "",
			Album:  "",
			Volume: 0,
		}, nil
	}

	var transportResponse SonosGetPositionInfoResponse
	if err := xml.Unmarshal(transportData, &transportResponse); err != nil {
		return &Status{
			State:  "stopped",
			Song:   "",
			Artist: "",
			Album:  "",
			Volume: 0,
		}, nil
	}

	state := strings.ToLower(transportResponse.Body.GetTransportInfo.CurrentTransportState)

	// Get position info (current track)
	positionBody := `<u:GetPositionInfo xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:GetPositionInfo>`

	positionData, err := sc.makeSoapRequest("GetPositionInfo", "AVTransport", positionBody)
	if err != nil {
		// Continue with basic state info
		return &Status{
			State:  state,
			Song:   "",
			Artist: "",
			Album:  "",
			Volume: 0,
		}, nil
	}

	var positionResponse SonosGetPositionInfoResponse
	if err := xml.Unmarshal(positionData, &positionResponse); err != nil {
		return &Status{
			State:  state,
			Song:   "",
			Artist: "",
			Album:  "",
			Volume: 0,
		}, nil
	}

	// Parse track metadata to extract song, artist, album
	metadata := positionResponse.Body.GetPositionInfo.TrackMetaData
	song, artist, album := parseSonosMetadata(metadata)

	// Get volume
	volumeBody := `<u:GetVolume xmlns:u="urn:schemas-upnp-org:service:RenderingControl:1">
		<InstanceID>0</InstanceID>
		<Channel>Master</Channel>
	</u:GetVolume>`

	volume := 0
	volumeData, err := sc.makeSoapRequest("GetVolume", "RenderingControl", volumeBody)
	if err == nil {
		var volumeResponse SonosGetPositionInfoResponse
		if err := xml.Unmarshal(volumeData, &volumeResponse); err == nil {
			volume, _ = strconv.Atoi(volumeResponse.Body.GetVolumeResponse.CurrentVolume)
		}
	}

	return &Status{
		State:  state,
		Song:   song,
		Artist: artist,
		Album:  album,
		Volume: volume,
	}, nil
}

func parseSonosMetadata(metadata string) (song, artist, album string) {
	titleRegex := regexp.MustCompile(`<dc:title[^>]*>(.*?)</dc:title>`)
	creatorRegex := regexp.MustCompile(`<dc:creator[^>]*>(.*?)</dc:creator>`)
	albumRegex := regexp.MustCompile(`<upnp:album[^>]*>(.*?)</upnp:album>`)

	if match := titleRegex.FindStringSubmatch(metadata); len(match) > 1 {
		song = html.UnescapeString(match[1])
	}
	if match := creatorRegex.FindStringSubmatch(metadata); len(match) > 1 {
		artist = html.UnescapeString(match[1])
	}
	if match := albumRegex.FindStringSubmatch(metadata); len(match) > 1 {
		album = html.UnescapeString(match[1])
	}

	return song, artist, album
}

func (sc *SonosClient) PlayPreset(id int) error {
	if err := sc.loadFavorites(); err != nil {
		return err
	}

	// Find the favorite
	var favorite *SonosFavorite
	for _, fav := range sc.favorites {
		if fav.ID == id {
			favorite = &fav
			break
		}
	}

	if favorite == nil {
		return fmt.Errorf("favorite not found")
	}

	// Skip INFO entries
	if strings.HasPrefix(favorite.Name, "[INFO]") {
		return fmt.Errorf("this is an info entry, not playable")
	}

	if favorite.URI == "" {
		return fmt.Errorf("no URI available for this favorite")
	}

	// For Sonos favorites, we need to use the original metadata from the browse response
	// The key is to preserve the exact metadata structure that Sonos expects
	var metadata string
	if favorite.Meta != "" {
		// Use the original metadata wrapped in DIDL-Lite envelope
		metadata = fmt.Sprintf(`&lt;DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"&gt;&lt;item id="FAVORITE"&gt;%s&lt;/item&gt;&lt;/DIDL-Lite&gt;`,
			strings.ReplaceAll(strings.ReplaceAll(favorite.Meta, "<", "&lt;"), ">", "&gt;"))
	} else {
		// Create minimal valid metadata for radio stations
		metadata = `&lt;DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"&gt;&lt;item id="R:0/0"&gt;&lt;dc:title&gt;` + html.EscapeString(favorite.Name) + `&lt;/dc:title&gt;&lt;upnp:class&gt;object.item.audioItem.audioBroadcast&lt;/upnp:class&gt;&lt;/item&gt;&lt;/DIDL-Lite&gt;`
	}

	// Create SOAP request with proper XML escaping
	body := fmt.Sprintf(`<u:SetAVTransportURI xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
		<CurrentURI>%s</CurrentURI>
		<CurrentURIMetaData>%s</CurrentURIMetaData>
	</u:SetAVTransportURI>`, html.EscapeString(favorite.URI), metadata)

	soapEnvelope := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>%s</s:Body>
</s:Envelope>`, body)

	url := fmt.Sprintf("%s/MediaRenderer/AVTransport/Control", sc.baseURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(soapEnvelope))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:AVTransport:1#SetAVTransportURI"`)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(soapEnvelope)))

	resp, err := sc.client.Do(req)
	if err != nil {
		return fmt.Errorf("SOAP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		// Try alternative approach for radio streams
		if strings.Contains(favorite.URI, "x-sonosapi") || strings.Contains(favorite.URI, "radio") {
			return sc.playRadioStation(favorite)
		}

		return fmt.Errorf("SetAVTransportURI failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Start playback
	return sc.Play()
}

func (sc *SonosClient) playRadioStation(favorite *SonosFavorite) error {
	// Clear the queue first
	clearBody := `<u:RemoveAllTracksFromQueue xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:RemoveAllTracksFromQueue>`

	_, err := sc.makeSoapRequest("RemoveAllTracksFromQueue", "AVTransport", clearBody)
	if err != nil {
		// Continue anyway
	}

	// Set the queue mode to play from queue
	setPlayModeBody := `<u:SetPlayMode xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
		<NewPlayMode>NORMAL</NewPlayMode>
	</u:SetPlayMode>`

	_, err = sc.makeSoapRequest("SetPlayMode", "AVTransport", setPlayModeBody)
	if err != nil {
		// Continue anyway
	}

	// Add the radio station to queue
	addBody := fmt.Sprintf(`<u:AddURIToQueue xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
		<EnqueuedURI>%s</EnqueuedURI>
		<EnqueuedURIMetaData>&lt;DIDL-Lite xmlns:dc="http://purl.org/dc/elements/1.1/" xmlns:upnp="urn:schemas-upnp-org:metadata-1-0/upnp/" xmlns="urn:schemas-upnp-org:metadata-1-0/DIDL-Lite/"&gt;&lt;item id="R:0/0"&gt;&lt;dc:title&gt;%s&lt;/dc:title&gt;&lt;upnp:class&gt;object.item.audioItem.audioBroadcast&lt;/upnp:class&gt;&lt;/item&gt;&lt;/DIDL-Lite&gt;</EnqueuedURIMetaData>
		<DesiredFirstTrackNumberEnqueued>1</DesiredFirstTrackNumberEnqueued>
		<EnqueueAsNext>0</EnqueueAsNext>
	</u:AddURIToQueue>`, html.EscapeString(favorite.URI), html.EscapeString(favorite.Name))

	_, err = sc.makeSoapRequest("AddURIToQueue", "AVTransport", addBody)
	if err != nil {
		return fmt.Errorf("failed to add radio to queue: %w", err)
	}

	// Seek to the first track in queue
	seekBody := `<u:Seek xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
		<Unit>TRACK_NR</Unit>
		<Target>1</Target>
	</u:Seek>`

	_, err = sc.makeSoapRequest("Seek", "AVTransport", seekBody)
	if err != nil {
		// Continue anyway
	}

	// Start playback
	return sc.Play()
}

func (sc *SonosClient) Play() error {
	body := `<u:Play xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
		<Speed>1</Speed>
	</u:Play>`

	_, err := sc.makeSoapRequest("Play", "AVTransport", body)
	return err
}

func (sc *SonosClient) Pause() error {
	body := `<u:Pause xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:Pause>`

	_, err := sc.makeSoapRequest("Pause", "AVTransport", body)
	return err
}

func (sc *SonosClient) Stop() error {
	body := `<u:Stop xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:Stop>`

	_, err := sc.makeSoapRequest("Stop", "AVTransport", body)
	return err
}

func (sc *SonosClient) SetVolume(level int) error {
	if level < 0 || level > 100 {
		return fmt.Errorf("volume must be between 0 and 100")
	}

	body := fmt.Sprintf(`<u:SetVolume xmlns:u="urn:schemas-upnp-org:service:RenderingControl:1">
		<InstanceID>0</InstanceID>
		<Channel>Master</Channel>
		<DesiredVolume>%d</DesiredVolume>
	</u:SetVolume>`, level)

	_, err := sc.makeSoapRequest("SetVolume", "RenderingControl", body)
	return err
}

func (sc *SonosClient) Next() error {
	body := `<u:Next xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:Next>`

	_, err := sc.makeSoapRequest("Next", "AVTransport", body)
	return err
}

func (sc *SonosClient) Previous() error {
	body := `<u:Previous xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:Previous>`

	_, err := sc.makeSoapRequest("Previous", "AVTransport", body)
	return err
}

func (sc *SonosClient) AddSlave(slaveIP string) error {
	// Sonos grouping is more complex - for now, return not implemented
	return fmt.Errorf("Sonos grouping not yet implemented")
}

func (sc *SonosClient) RemoveSlave(slaveIP string) error {
	return fmt.Errorf("Sonos grouping not yet implemented")
}

func (sc *SonosClient) RemoveAllSlaves() error {
	return fmt.Errorf("Sonos grouping not yet implemented")
}

func (sc *SonosClient) LeaveGroup() error {
	return fmt.Errorf("Sonos grouping not yet implemented")
}

func (sc *SonosClient) GetDeviceType() DeviceType {
	return DeviceTypeSonos
}

func (sc *SonosClient) DebugAPI() string {
	// Test basic HTTP connectivity first
	resp, err := sc.client.Get(sc.baseURL + "/xml/device_description.xml")
	if err != nil {
		return fmt.Sprintf("Sonos Debug: Device not reachable: %v", err)
	}
	resp.Body.Close()

	// Test SOAP services with correct actions
	var results []string

	// Test AVTransport
	if sc.testAVTransport() {
		results = append(results, "AVTransport: ✅")
	} else {
		results = append(results, "AVTransport: ❌")
	}

	// Test RenderingControl
	if sc.testRenderingControl() {
		results = append(results, "RenderingControl: ✅")
	} else {
		results = append(results, "RenderingControl: ❌")
	}

	// Test ContentDirectory
	if sc.testContentDirectory() {
		results = append(results, "ContentDirectory: ✅")
	} else {
		results = append(results, "ContentDirectory: ❌")
	}

	// Add favorite discovery debug info
	sc.favorites = nil // Clear cache to force reload
	sc.loadFavorites()
	results = append(results, fmt.Sprintf("Favorites: %d found", len(sc.favorites)))

	return fmt.Sprintf("Sonos Debug: %s", strings.Join(results, " | "))
}

func (sc *SonosClient) testAVTransport() bool {
	body := `<u:GetTransportInfo xmlns:u="urn:schemas-upnp-org:service:AVTransport:1">
		<InstanceID>0</InstanceID>
	</u:GetTransportInfo>`

	_, err := sc.makeSoapRequest("GetTransportInfo", "AVTransport", body)
	return err == nil
}

func (sc *SonosClient) testRenderingControl() bool {
	body := `<u:GetVolume xmlns:u="urn:schemas-upnp-org:service:RenderingControl:1">
		<InstanceID>0</InstanceID>
		<Channel>Master</Channel>
	</u:GetVolume>`

	_, err := sc.makeSoapRequest("GetVolume", "RenderingControl", body)
	return err == nil
}

func (sc *SonosClient) testContentDirectory() bool {
	// Try MediaServer path first
	body := `<u:Browse xmlns:u="urn:schemas-upnp-org:service:ContentDirectory:1">
		<ObjectID>0</ObjectID>
		<BrowseFlag>BrowseMetadata</BrowseFlag>
		<Filter>*</Filter>
		<StartingIndex>0</StartingIndex>
		<RequestedCount>1</RequestedCount>
		<SortCriteria></SortCriteria>
	</u:Browse>`

	soapEnvelope := fmt.Sprintf(`<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/" s:encodingStyle="http://schemas.xmlsoap.org/soap/encoding/">
<s:Body>%s</s:Body>
</s:Envelope>`, body)

	url := fmt.Sprintf("%s/MediaServer/ContentDirectory/Control", sc.baseURL)
	req, err := http.NewRequest("POST", url, strings.NewReader(soapEnvelope))
	if err != nil {
		return false
	}

	req.Header.Set("Content-Type", "text/xml; charset=utf-8")
	req.Header.Set("SOAPAction", `"urn:schemas-upnp-org:service:ContentDirectory:1#Browse"`)
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(soapEnvelope)))

	resp, err := sc.client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Update TUI state
func updateStatus() {
	status, err := tuiState.client.GetStatus()
	if err != nil {
		tuiState.statusError = getText("error_retrieving_status")
		tuiState.status = nil
	} else {
		tuiState.status = status
		tuiState.statusError = ""
	}
}

func updatePresets() {
	presets, err := tuiState.client.GetPresets()
	if err != nil {
		tuiState.presetsError = getText("error_loading_presets")
		tuiState.presets = nil
	} else {
		tuiState.presets = presets
		tuiState.presetsError = ""
	}
}

// Render the complete TUI
func renderTUI() {
	clearScreen()

	// Header
	fmt.Println(getText("title"))
	fmt.Println(strings.Repeat("=", 70))
	deviceTypeIndicator := ""
	if tuiState.client != nil {
		switch tuiState.client.GetDeviceType() {
		case DeviceTypeBluOS:
			deviceTypeIndicator = " [BluOS]"
		case DeviceTypeSonos:
			deviceTypeIndicator = " [Sonos]"
		}
	}
	fmt.Printf("🔗 %s %s%s\n", getText("current_player"), tuiState.playerName, deviceTypeIndicator)
	fmt.Println()

	// Available Players Section
	if len(tuiState.availablePlayers) > 1 {
		fmt.Println(getText("available_outputs"))
		for i, player := range tuiState.availablePlayers {
			activeMarker := ""
			if player.Name == tuiState.playerName {
				activeMarker = " ✅"
			}
			typeIndicator := ""
			switch player.Type {
			case DeviceTypeBluOS:
				typeIndicator = " [BluOS]"
			case DeviceTypeSonos:
				typeIndicator = " [Sonos]"
			}
			fmt.Printf("  [%d] %s (%s)%s%s\n", i+1, player.Name, player.IP, typeIndicator, activeMarker)
		}

		// Show possible group combinations (only for compatible devices)
		if len(tuiState.availablePlayers) > 1 {
			fmt.Println(getText("group_combinations"))
			for i, master := range tuiState.availablePlayers {
				for j, slave := range tuiState.availablePlayers {
					if i != j && master.Type == slave.Type && master.Type == DeviceTypeBluOS {
						fmt.Printf("  group %d+%d - %s + %s\n", i+1, j+1, master.Name, slave.Name)
					}
				}
			}
		}
		fmt.Println()
	}

	// Status Section
	if tuiState.statusError != "" {
		fmt.Println(tuiState.statusError)
	} else if tuiState.status != nil {
		volumeStr := getText("volume_unknown")
		if tuiState.status.Volume >= 0 {
			volumeStr = fmt.Sprintf("%d%%", tuiState.status.Volume)
		}
		fmt.Printf(getText("status_volume")+"\n", tuiState.status.State, volumeStr)
		if tuiState.status.Song != "" {
			fmt.Printf("🎵 %s", tuiState.status.Song)
			if tuiState.status.Artist != "" {
				fmt.Printf(" - %s", tuiState.status.Artist)
			}
			if tuiState.status.Album != "" {
				fmt.Printf(" (%s)", tuiState.status.Album)
			}
			fmt.Println()
		} else {
			fmt.Printf("🎵 %s\n", getText("no_song_playing"))
		}
	}
	fmt.Println()

	// Presets Section
	fmt.Println(getText("available_presets"))
	if tuiState.presetsError != "" {
		fmt.Println(tuiState.presetsError)
	} else if tuiState.presets != nil {
		for _, preset := range tuiState.presets {
			fmt.Printf("  [%d] %s\n", preset.ID, preset.Name)
		}
	}
	fmt.Println()

	// Commands Section - Display in compact rows
	fmt.Println(getText("available_commands"))
	fmt.Println("  play <id> | play | pause | stop | next | prev | vol <0-100>")
	fmt.Println("  output <id> | group <id1+id2> | ungroup | lang <en|de|sw> | quit")
	fmt.Println()

	// Last Action
	if tuiState.lastAction != "" {
		fmt.Printf("%s %s\n", getText("last_action"), tuiState.lastAction)
		fmt.Println()
	}

	// Separator line
	fmt.Println(strings.Repeat("=", 70))
}

// Player selection
func selectPlayer() (AudioClient, string, []PlayerInfo, error) {
	players, err := scanForPlayers()
	if err != nil {
		return nil, "", nil, err
	}

	if len(players) == 0 {
		return nil, "", nil, fmt.Errorf(getText("no_players"))
	}

	fmt.Println("\n" + getText("available_players"))
	for i, player := range players {
		typeIndicator := ""
		switch player.Type {
		case DeviceTypeBluOS:
			typeIndicator = " [BluOS]"
		case DeviceTypeSonos:
			typeIndicator = " [Sonos]"
		}
		fmt.Printf("  [%d] %s (%s %s) - %s%s\n", i+1, player.Name, player.Brand, player.Model, player.IP, typeIndicator)
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("\n"+getText("select_player"), len(players))
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > len(players) {
			fmt.Println(getText("invalid_selection"))
			continue
		}

		selectedPlayer := players[choice-1]
		fmt.Printf(getText("connected_to")+"\n", selectedPlayer.Name, selectedPlayer.IP)

		var client AudioClient
		switch selectedPlayer.Type {
		case DeviceTypeBluOS:
			client = NewBluesoundClient(selectedPlayer.IP)
		case DeviceTypeSonos:
			client = NewSonosClient(selectedPlayer.IP)
		default:
			return nil, "", nil, fmt.Errorf("unsupported device type")
		}

		return client, selectedPlayer.Name, players, nil
	}
}

// Switch to different player
func switchToPlayer(playerID int) {
	if playerID < 1 || playerID > len(tuiState.availablePlayers) {
		tuiState.lastAction = getText("invalid_player_id")
		return
	}

	selectedPlayer := tuiState.availablePlayers[playerID-1]

	switch selectedPlayer.Type {
	case DeviceTypeBluOS:
		tuiState.client = NewBluesoundClient(selectedPlayer.IP)
	case DeviceTypeSonos:
		tuiState.client = NewSonosClient(selectedPlayer.IP)
	default:
		tuiState.lastAction = getText("error_switching_player")
		return
	}

	tuiState.playerName = selectedPlayer.Name
	tuiState.lastAction = fmt.Sprintf(getText("switched_to_player"), playerID, selectedPlayer.Name)

	// Update status and presets for new player
	updateStatus()
	updatePresets()
}

// Group players (only works for BluOS devices)
func groupPlayers(groupSpec string) {
	parts := strings.Split(groupSpec, "+")
	if len(parts) != 2 {
		tuiState.lastAction = getText("invalid_group_format")
		return
	}

	masterID, err1 := strconv.Atoi(parts[0])
	slaveID, err2 := strconv.Atoi(parts[1])

	if err1 != nil || err2 != nil || masterID < 1 || slaveID < 1 ||
		masterID > len(tuiState.availablePlayers) || slaveID > len(tuiState.availablePlayers) {
		tuiState.lastAction = getText("invalid_group_format")
		return
	}

	if masterID == slaveID {
		tuiState.lastAction = getText("invalid_group_format")
		return
	}

	masterPlayer := tuiState.availablePlayers[masterID-1]
	slavePlayer := tuiState.availablePlayers[slaveID-1]

	// Check if both are BluOS devices
	if masterPlayer.Type != DeviceTypeBluOS || slavePlayer.Type != DeviceTypeBluOS {
		tuiState.lastAction = "❌ Grouping only supported for BluOS devices"
		return
	}

	// Switch to master player
	tuiState.client = NewBluesoundClient(masterPlayer.IP)
	tuiState.playerName = masterPlayer.Name

	// Add slave to master
	if err := tuiState.client.AddSlave(slavePlayer.IP); err != nil {
		tuiState.lastAction = getText("error_grouping")
		return
	}

	tuiState.lastAction = fmt.Sprintf(getText("grouped_players"), masterPlayer.Name)
	updateStatus()
}

// Debug function to test API endpoints
func debugAPI() {
	if tuiState.client != nil {
		tuiState.lastAction = tuiState.client.DebugAPI()
	} else {
		tuiState.lastAction = "No client connected"
	}
}

// Ungroup all players (only works for BluOS devices)
func ungroupAll() {
	if tuiState.client == nil {
		tuiState.lastAction = "No client connected"
		return
	}

	if tuiState.client.GetDeviceType() != DeviceTypeBluOS {
		tuiState.lastAction = "❌ Ungrouping only supported for BluOS devices"
		return
	}

	var successCount int

	// Try removing slaves one by one using RemoveSlave
	for _, player := range tuiState.availablePlayers {
		if player.Name != tuiState.playerName && player.Type == DeviceTypeBluOS {
			if _, err := tuiState.client.(*BluesoundClient).makeRequest(fmt.Sprintf("/RemoveSlave?slave=%s", player.IP)); err == nil {
				successCount++
			}

			// Also try the reverse
			otherClient := NewBluesoundClient(player.IP)
			currentPlayerIP := strings.Split(tuiState.client.(*BluesoundClient).baseURL, "://")[1]
			currentPlayerIP = strings.Split(currentPlayerIP, ":")[0]

			if _, err := otherClient.makeRequest(fmt.Sprintf("/RemoveSlave?slave=%s", currentPlayerIP)); err == nil {
				successCount++
			}
		}
	}

	// Try various standalone/reset approaches on all BluOS players
	for _, player := range tuiState.availablePlayers {
		if player.Type == DeviceTypeBluOS {
			client := NewBluesoundClient(player.IP)

			// Try various standalone/reset approaches
			standaloneMethods := []string{
				"/Standalone",
				"/Reset",
				"/ClearSlaves",
			}

			for _, method := range standaloneMethods {
				if _, err := client.makeRequest(method); err == nil {
					successCount++
					break
				}
			}
		}
	}

	if successCount > 0 {
		tuiState.lastAction = getText("ungrouped_all")
	} else {
		tuiState.lastAction = fmt.Sprintf("%s (RemoveSlave approach failed)", getText("error_ungrouping"))
	}

	updateStatus()
}

// Change language
func changeLanguage(lang string) {
	switch strings.ToLower(lang) {
	case "en", "english":
		currentLanguage = LangEnglish
		tuiState.lastAction = getText("language_changed") + " English"
	case "de", "german", "deutsch":
		currentLanguage = LangGerman
		tuiState.lastAction = getText("language_changed") + " Deutsch"
	case "sw", "swahili", "kiswahili":
		currentLanguage = LangSwahili
		tuiState.lastAction = getText("language_changed") + " Kiswahili"
	default:
		tuiState.lastAction = getText("invalid_language")
	}
}

// Interactive loop
func interactiveMode() {
	reader := bufio.NewReader(os.Stdin)

	// Initial data load
	updateStatus()
	updatePresets()

	for {
		renderTUI()
		fmt.Print(getText("prompt"))

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		parts := strings.Fields(input)
		command := strings.ToLower(parts[0])

		switch command {
		case "play":
			if len(parts) > 1 {
				// Play preset/favorite
				presetID, err := strconv.Atoi(parts[1])
				if err != nil {
					tuiState.lastAction = getText("invalid_preset_id")
					continue
				}
				if err := tuiState.client.PlayPreset(presetID); err != nil {
					tuiState.lastAction = fmt.Sprintf("%s: %v", getText("error_playing_preset"), err)
				} else {
					tuiState.lastAction = fmt.Sprintf(getText("playing_preset"), presetID)
					time.Sleep(500 * time.Millisecond)
					updateStatus()
				}
			} else {
				// Start playback
				if err := tuiState.client.Play(); err != nil {
					tuiState.lastAction = getText("error_starting_playback")
				} else {
					tuiState.lastAction = getText("playback_started")
					time.Sleep(500 * time.Millisecond)
					updateStatus()
				}
			}

		case "pause":
			if err := tuiState.client.Pause(); err != nil {
				tuiState.lastAction = getText("error_pausing")
			} else {
				tuiState.lastAction = getText("paused")
				updateStatus()
			}

		case "stop":
			if err := tuiState.client.Stop(); err != nil {
				tuiState.lastAction = getText("error_stopping")
			} else {
				tuiState.lastAction = getText("stopped")
				updateStatus()
			}

		case "next":
			if err := tuiState.client.Next(); err != nil {
				tuiState.lastAction = getText("error_next_track")
			} else {
				tuiState.lastAction = getText("next_track")
				time.Sleep(500 * time.Millisecond)
				updateStatus()
			}

		case "prev", "previous":
			if err := tuiState.client.Previous(); err != nil {
				tuiState.lastAction = getText("error_prev_track")
			} else {
				tuiState.lastAction = getText("prev_track")
				time.Sleep(500 * time.Millisecond)
				updateStatus()
			}

		case "vol", "volume":
			if len(parts) < 2 {
				tuiState.lastAction = getText("volume_missing")
				continue
			}
			volume, err := strconv.Atoi(parts[1])
			if err != nil {
				tuiState.lastAction = getText("invalid_volume")
				continue
			}
			if err := tuiState.client.SetVolume(volume); err != nil {
				tuiState.lastAction = getText("error_setting_volume")
			} else {
				tuiState.lastAction = fmt.Sprintf(getText("volume_set"), volume)
				updateStatus()
			}

		case "status":
			updateStatus()
			tuiState.lastAction = "Status refreshed"

		case "presets":
			updatePresets()
			tuiState.lastAction = "Presets/Favorites refreshed"

		case "help":
			tuiState.lastAction = "Help displayed above"

		case "output":
			if len(parts) < 2 {
				tuiState.lastAction = getText("invalid_player_id")
				continue
			}
			playerID, err := strconv.Atoi(parts[1])
			if err != nil {
				tuiState.lastAction = getText("invalid_player_id")
				continue
			}
			switchToPlayer(playerID)

		case "group":
			if len(parts) < 2 {
				tuiState.lastAction = getText("invalid_group_format")
				continue
			}
			groupPlayers(parts[1])

		case "ungroup":
			ungroupAll()

		case "debug":
			debugAPI()

		case "lang", "language":
			if len(parts) < 2 {
				tuiState.lastAction = getText("invalid_language")
				continue
			}
			changeLanguage(parts[1])

		case "quit", "exit":
			clearScreen()
			fmt.Println(getText("goodbye"))
			return

		default:
			tuiState.lastAction = fmt.Sprintf(getText("unknown_command"), command)
		}
	}
}

func main() {
	fmt.Println(getText("title"))
	fmt.Println(strings.Repeat("=", 70))

	// Select player
	client, playerName, availablePlayers, err := selectPlayer()
	if err != nil {
		log.Fatalf(getText("error_selecting_player"), err)
	}

	// Initialize TUI state
	tuiState.client = client
	tuiState.playerName = playerName
	tuiState.availablePlayers = availablePlayers

	// Start interactive mode
	interactiveMode()
}
