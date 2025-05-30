package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	BluesoundPort = "11000"
	ScanTimeout   = 3 * time.Second
)

// Language support
type Language string

const (
	LangEnglish  Language = "en"
	LangGerman   Language = "de"
	LangSwahili  Language = "sw"
)

var currentLanguage = LangEnglish

// Localization texts
var texts = map[Language]map[string]string{
	LangEnglish: {
		"title":                    "🎵 BluOS Controller",
		"scanning":                 "🔍 Scanning network for BluOS players...",
		"scanning_network":         "   Scanning network: %s",
		"found_player":             "   ✅ Found: %s (%s) at %s",
		"no_players":               "no BluOS players found",
		"could_not_determine_ip":   "could not determine local IP: %w",
		"available_players":        "📱 Available Players:",
		"select_player":            "Select a player (1-%d): ",
		"invalid_selection":        "❌ Invalid selection",
		"connected_to":             "✅ Connected to: %s (%s)",
		"error_selecting_player":   "Error selecting player: %v",
		"interactive_mode":         "🎵 BluOS Controller - Interactive Mode",
		"separator":                "=======================================",
		"status_volume":            "📊 Status: %s | Volume: %s",
		"volume_unknown":           "N/A",
		"error_retrieving_status":  "❌ Error retrieving status",
		"available_presets":        "📋 Available Presets:",
		"error_loading_presets":    "❌ Error loading presets",
		"available_commands":       "🎮 Available Commands:",
		"cmd_play_preset":          "play <id>   - Play preset",
		"cmd_play":                 "play       - Start playback",
		"cmd_pause":                "pause      - Pause playback",
		"cmd_stop":                 "stop       - Stop playback",
		"cmd_next":                 "next       - Next track",
		"cmd_prev":                 "prev       - Previous track",
		"cmd_volume":               "vol <0-100> - Set volume",
		"cmd_status":               "status     - Refresh status",
		"cmd_presets":              "presets    - Refresh presets",
		"cmd_help":                 "help       - Show help",
		"cmd_lang":                 "lang <en|de|sw> - Change language",
		"cmd_output":               "output <id> - Switch to player",
		"cmd_group":                "group <id1+id2> - Group players",
		"cmd_ungroup":              "ungroup - Remove all groups",
		"cmd_debug":                "debug - Show API endpoints",
		"cmd_quit":                 "quit/exit  - Exit program",
		"prompt":                   "Command> ",
		"invalid_preset_id":        "❌ Invalid preset ID",
		"error_playing_preset":     "❌ Error playing preset",
		"playing_preset":           "✅ Playing preset %d",
		"error_starting_playback":  "❌ Error starting playback",
		"playback_started":         "▶️ Playback started",
		"error_pausing":            "❌ Error pausing",
		"paused":                   "⏸️ Paused",
		"error_stopping":           "❌ Error stopping",
		"stopped":                  "⏹️ Stopped",
		"error_next_track":         "❌ Error skipping to next track",
		"next_track":               "⏭️ Next track",
		"error_prev_track":         "❌ Error going to previous track",
		"prev_track":               "⏮️ Previous track",
		"volume_missing":           "❌ Volume value missing",
		"invalid_volume":           "❌ Invalid volume value",
		"error_setting_volume":     "❌ Error setting volume",
		"volume_set":               "🔊 Volume set to %d%%",
		"language_changed":         "🌍 Language changed to",
		"invalid_language":         "❌ Invalid language. Use: en, de, sw",
		"goodbye":                  "👋 Goodbye!",
		"unknown_command":          "❌ Unknown command: %s (Type 'help' for help)",
		"last_action":              "Last Action:",
		"no_song_playing":          "No song playing",
		"available_outputs":        "📱 Available Players:",
		"current_player":           "Current Player:",
		"switched_to_player":       "🔄 Switched to player %d: %s",
		"invalid_player_id":        "❌ Invalid player ID",
		"error_switching_player":   "❌ Error switching to player",
		"grouped_players":          "🔗 Grouped players: %s as master",
		"invalid_group_format":     "❌ Invalid group format. Use: group <id1+id2>",
		"error_grouping":           "❌ Error grouping players",
		"group_combinations":       "🎵 Group Combinations:",
		"ungrouped_all":            "🔓 All player groups removed",
		"error_ungrouping":         "❌ Error removing groups",
	},
	LangGerman: {
		"title":                    "🎵 BluOS Controller",
		"scanning":                 "🔍 Suche nach BluOS Playern im Netzwerk...",
		"scanning_network":         "   Scanne Netzwerk: %s",
		"found_player":             "   ✅ Gefunden: %s (%s) auf %s",
		"no_players":               "keine BluOS Player gefunden",
		"could_not_determine_ip":   "konnte lokale IP nicht ermitteln: %w",
		"available_players":        "📱 Verfügbare Player:",
		"select_player":            "Wähle einen Player (1-%d): ",
		"invalid_selection":        "❌ Ungültige Auswahl",
		"connected_to":             "✅ Verbunden mit: %s (%s)",
		"error_selecting_player":   "Fehler bei der Player-Auswahl: %v",
		"interactive_mode":         "🎵 BluOS Controller - Interaktiver Modus",
		"separator":                "==========================================",
		"status_volume":            "📊 Status: %s | Lautstärke: %s",
		"volume_unknown":           "N/A",
		"error_retrieving_status":  "❌ Fehler beim Abrufen des Status",
		"available_presets":        "📋 Verfügbare Presets:",
		"error_loading_presets":    "❌ Fehler beim Laden der Presets",
		"available_commands":       "🎮 Verfügbare Befehle:",
		"cmd_play_preset":          "play <id>   - Preset abspielen",
		"cmd_play":                 "play       - Wiedergabe starten",
		"cmd_pause":                "pause      - Pausieren",
		"cmd_stop":                 "stop       - Stoppen",
		"cmd_next":                 "next       - Nächster Titel",
		"cmd_prev":                 "prev       - Vorheriger Titel",
		"cmd_volume":               "vol <0-100> - Lautstärke setzen",
		"cmd_status":               "status     - Status aktualisieren",
		"cmd_presets":              "presets    - Presets aktualisieren",
		"cmd_help":                 "help       - Hilfe anzeigen",
		"cmd_lang":                 "lang <en|de|sw> - Sprache ändern",
		"cmd_output":               "output <id> - Zu Player wechseln",
		"cmd_group":                "group <id1+id2> - Player gruppieren",
		"cmd_ungroup":              "ungroup - Alle Gruppen auflösen",
		"cmd_debug":                "debug - API-Endpunkte anzeigen",
		"cmd_quit":                 "quit/exit  - Programm beenden",
		"prompt":                   "Befehl> ",
		"invalid_preset_id":        "❌ Ungültige Preset-ID",
		"error_playing_preset":     "❌ Fehler beim Abspielen",
		"playing_preset":           "✅ Preset %d wird abgespielt",
		"error_starting_playback":  "❌ Fehler beim Starten",
		"playback_started":         "▶️ Wiedergabe gestartet",
		"error_pausing":            "❌ Fehler beim Pausieren",
		"paused":                   "⏸️ Pausiert",
		"error_stopping":           "❌ Fehler beim Stoppen",
		"stopped":                  "⏹️ Gestoppt",
		"error_next_track":         "❌ Fehler beim Weiterschalten",
		"next_track":               "⏭️ Nächster Titel",
		"error_prev_track":         "❌ Fehler beim Zurückschalten",
		"prev_track":               "⏮️ Vorheriger Titel",
		"volume_missing":           "❌ Lautstärke-Wert fehlt",
		"invalid_volume":           "❌ Ungültiger Lautstärke-Wert",
		"error_setting_volume":     "❌ Fehler beim Setzen der Lautstärke",
		"volume_set":               "🔊 Lautstärke auf %d%% gesetzt",
		"language_changed":         "🌍 Sprache geändert zu",
		"invalid_language":         "❌ Ungültige Sprache. Verwende: en, de, sw",
		"goodbye":                  "👋 Auf Wiedersehen!",
		"unknown_command":          "❌ Unbekannter Befehl: %s (Tippe 'help' für Hilfe)",
		"last_action":              "Letzte Aktion:",
		"no_song_playing":          "Kein Lied wird abgespielt",
		"available_outputs":        "📱 Verfügbare Player:",
		"current_player":           "Aktueller Player:",
		"switched_to_player":       "🔄 Gewechselt zu Player %d: %s",
		"invalid_player_id":        "❌ Ungültige Player-ID",
		"error_switching_player":   "❌ Fehler beim Wechseln des Players",
		"grouped_players":          "🔗 Player gruppiert: %s als Master",
		"invalid_group_format":     "❌ Ungültiges Gruppen-Format. Verwende: group <id1+id2>",
		"error_grouping":           "❌ Fehler beim Gruppieren",
		"group_combinations":       "🎵 Gruppen-Kombinationen:",
		"ungrouped_all":            "🔓 Alle Player-Gruppen aufgelöst",
		"error_ungrouping":         "❌ Fehler beim Auflösen der Gruppen",
	},
	LangSwahili: {
		"title":                    "🎵 Kidhibiti cha BluOS",
		"scanning":                 "🔍 Kutafuta vichezaji vya BluOS kwenye mtandao...",
		"scanning_network":         "   Kutafuta mtandao: %s",
		"found_player":             "   ✅ Kumepatikana: %s (%s) kwa %s",
		"no_players":               "hakuna vichezaji vya BluOS vilivopatikana",
		"could_not_determine_ip":   "haikuweza kutambua IP ya ndani: %w",
		"available_players":        "📱 Vichezaji Vinavyopatikana:",
		"select_player":            "Chagua kichezaji (1-%d): ",
		"invalid_selection":        "❌ Chaguo batili",
		"connected_to":             "✅ Imeunganishwa na: %s (%s)",
		"error_selecting_player":   "Hitilafu katika kuchagua kichezaji: %v",
		"interactive_mode":         "🎵 Kidhibiti cha BluOS - Hali ya Maingiliano",
		"separator":                "===========================================",
		"status_volume":            "📊 Hali: %s | Sauti: %s",
		"volume_unknown":           "N/A",
		"error_retrieving_status":  "❌ Hitilafu katika kupata hali",
		"available_presets":        "📋 Mipangilio Inayopatikana:",
		"error_loading_presets":    "❌ Hitilafu katika kupakia mipangilio",
		"available_commands":       "🎮 Amri Zinazopatikana:",
		"cmd_play_preset":          "play <id>   - Cheza mpangilio",
		"cmd_play":                 "play       - Anza kucheza",
		"cmd_pause":                "pause      - Simamisha",
		"cmd_stop":                 "stop       - Acha",
		"cmd_next":                 "next       - Wimbo ujao",
		"cmd_prev":                 "prev       - Wimbo uliopita",
		"cmd_volume":               "vol <0-100> - Weka sauti",
		"cmd_status":               "status     - Onyesha hali",
		"cmd_presets":              "presets    - Onyesha mipangilio",
		"cmd_help":                 "help       - Onyesha msaada",
		"cmd_lang":                 "lang <en|de|sw> - Badilisha lugha",
		"cmd_output":               "output <id> - Badili kichezaji",
		"cmd_group":                "group <id1+id2> - Unganisha vichezaji",
		"cmd_ungroup":              "ungroup - Ondoa vikundi vyote",
		"cmd_debug":                "debug - Onyesha API endpoints",
		"cmd_quit":                 "quit/exit  - Toka programu",
		"prompt":                   "Amri> ",
		"invalid_preset_id":        "❌ Kitambulisho cha mpangilio si halali",
		"error_playing_preset":     "❌ Hitilafu katika kucheza mpangilio",
		"playing_preset":           "✅ Kucheza mpangilio %d",
		"error_starting_playback":  "❌ Hitilafu katika kuanza kucheza",
		"playback_started":         "▶️ Imeanza kucheza",
		"error_pausing":            "❌ Hitilafu katika kusimamisha",
		"paused":                   "⏸️ Imesimamishwa",
		"error_stopping":           "❌ Hitilafu katika kuacha",
		"stopped":                  "⏹️ Imeachwa",
		"error_next_track":         "❌ Hitilafu katika kuruka wimbo ujao",
		"next_track":               "⏭️ Wimbo ujao",
		"error_prev_track":         "❌ Hitilafu katika kurudi wimbo uliopita",
		"prev_track":               "⏮️ Wimbo uliopita",
		"volume_missing":           "❌ Thamani ya sauti inakosekana",
		"invalid_volume":           "❌ Thamani ya sauti si halali",
		"error_setting_volume":     "❌ Hitilafu katika kuweka sauti",
		"volume_set":               "🔊 Sauti imewekwa %d%%",
		"language_changed":         "🌍 Lugha imebadilishwa kuwa",
		"invalid_language":         "❌ Lugha si halali. Tumia: en, de, sw",
		"goodbye":                  "👋 Kwaheri!",
		"unknown_command":          "❌ Amri isiyojulikana: %s (Andika 'help' kwa msaada)",
		"last_action":              "Kitendo cha Mwisho:",
		"no_song_playing":          "Hakuna wimbo unaochezwa",
		"available_outputs":        "📱 Vichezaji Vinavyopatikana:",
		"current_player":           "Kichezaji cha Sasa:",
		"switched_to_player":       "🔄 Imebadilishwa kwa kichezaji %d: %s",
		"invalid_player_id":        "❌ Kitambulisho cha kichezaji si halali",
		"error_switching_player":   "❌ Hitilafu katika kubadili kichezaji",
		"grouped_players":          "🔗 Vichezaji vimeunganishwa: %s kama mkuu",
		"invalid_group_format":     "❌ Muundo wa kikundi si halali. Tumia: group <id1+id2>",
		"error_grouping":           "❌ Hitilafu katika kuunganisha",
		"group_combinations":       "🎵 Miunganiko ya Vikundi:",
		"ungrouped_all":            "🔓 Vikundi vyote vya vichezaji vimeondolewa",
		"error_ungrouping":         "❌ Hitilafu katika kuondoa vikundi",
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

// Structures for XML parsing
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

// Player info for scan results
type PlayerInfo struct {
	IP    string
	Name  string
	Brand string
	Model string
}

// Bluesound API Client
type BluesoundClient struct {
	baseURL string
	client  *http.Client
}

// Global state for TUI
type TUIState struct {
	client           *BluesoundClient
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

// Clear screen and move cursor to top
func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

// Network scanner
func scanForPlayers() ([]PlayerInfo, error) {
	fmt.Println(getText("scanning"))
	
	// Get local IP
	localIP, err := getLocalIP()
	if err != nil {
		return nil, fmt.Errorf(getText("could_not_determine_ip"), err)
	}
	
	// Calculate network range
	subnet := getSubnet(localIP)
	fmt.Printf(getText("scanning_network")+"\n", subnet)
	
	var players []PlayerInfo
	var mu sync.Mutex
	var wg sync.WaitGroup
	
	// Scan all IPs in subnet in parallel
	for i := 1; i < 255; i++ {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			
			if player, found := checkForBluOSPlayer(ip); found {
				mu.Lock()
				players = append(players, player)
				mu.Unlock()
				fmt.Printf(getText("found_player")+"\n", player.Name, player.Model, player.IP)
			}
		}(fmt.Sprintf("%s.%d", subnet, i))
	}
	
	wg.Wait()
	return players, nil
}

func getLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()
	
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
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
	}, true
}

// API methods
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
	fmt.Printf("🔗 %s %s\n", getText("current_player"), tuiState.playerName)
	fmt.Println()
	
	// Available Players Section
	if len(tuiState.availablePlayers) > 1 {
		fmt.Println(getText("available_outputs"))
		for i, player := range tuiState.availablePlayers {
			activeMarker := ""
			if player.Name == tuiState.playerName {
				activeMarker = " ✅"
			}
			fmt.Printf("  [%d] %s (%s)%s\n", i+1, player.Name, player.IP, activeMarker)
		}
		
		// Show possible group combinations
		if len(tuiState.availablePlayers) > 1 {
			fmt.Println(getText("group_combinations"))
			for i, master := range tuiState.availablePlayers {
				for j, slave := range tuiState.availablePlayers {
					if i != j {
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
	
	// Commands Section - Display in compact rows (hide utility commands)
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
func selectPlayer() (*BluesoundClient, string, []PlayerInfo, error) {
	players, err := scanForPlayers()
	if err != nil {
		return nil, "", nil, err
	}
	
	if len(players) == 0 {
		return nil, "", nil, fmt.Errorf(getText("no_players"))
	}
	
	fmt.Println("\n" + getText("available_players"))
	for i, player := range players {
		fmt.Printf("  [%d] %s (%s %s) - %s\n", i+1, player.Name, player.Brand, player.Model, player.IP)
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
		return NewBluesoundClient(selectedPlayer.IP), selectedPlayer.Name, players, nil
	}
}

// Switch to different player
func switchToPlayer(playerID int) {
	if playerID < 1 || playerID > len(tuiState.availablePlayers) {
		tuiState.lastAction = getText("invalid_player_id")
		return
	}
	
	selectedPlayer := tuiState.availablePlayers[playerID-1]
	tuiState.client = NewBluesoundClient(selectedPlayer.IP)
	tuiState.playerName = selectedPlayer.Name
	tuiState.lastAction = fmt.Sprintf(getText("switched_to_player"), playerID, selectedPlayer.Name)
	
	// Update status and presets for new player
	updateStatus()
	updatePresets()
}

// Group players
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
	tuiState.lastAction = "Testing API endpoints..."
	
	// Test various common BluOS endpoints
	endpoints := []string{
		"/Status",
		"/SyncStatus", 
		"/Presets",
		"/RemoveSlave",
		"/AddSlave",
		"/Slaves",
		"/Standalone",
		"/Reset",
		"/ClearSlaves",
	}
	
	var results []string
	for _, endpoint := range endpoints {
		_, err := tuiState.client.makeRequest(endpoint)
		if err != nil {
			results = append(results, fmt.Sprintf("%s: ❌", endpoint))
		} else {
			results = append(results, fmt.Sprintf("%s: ✅", endpoint))
		}
	}
	
	tuiState.lastAction = fmt.Sprintf("API Test: %s", strings.Join(results, " | "))
}

// Ungroup all players
func ungroupAll() {
	var successCount int
	var errorMessages []string
	
	// Try removing slaves one by one using RemoveSlave
	for _, player := range tuiState.availablePlayers {
		if player.Name != tuiState.playerName {
			// Try to remove this player as a slave from current master
			if _, err := tuiState.client.makeRequest(fmt.Sprintf("/RemoveSlave?slave=%s", player.IP)); err == nil {
				successCount++
			} else {
				errorMessages = append(errorMessages, fmt.Sprintf("Remove %s: %v", player.Name, err))
			}
			
			// Also try the reverse - remove current player as slave from this one
			otherClient := NewBluesoundClient(player.IP)
			currentPlayerIP := strings.Split(tuiState.client.baseURL, "://")[1]
			currentPlayerIP = strings.Split(currentPlayerIP, ":")[0]
			
			if _, err := otherClient.makeRequest(fmt.Sprintf("/RemoveSlave?slave=%s", currentPlayerIP)); err == nil {
				successCount++
			}
		}
	}
	
	// Try various standalone/reset approaches on all players
	for _, player := range tuiState.availablePlayers {
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
				// Play preset
				presetID, err := strconv.Atoi(parts[1])
				if err != nil {
					tuiState.lastAction = getText("invalid_preset_id")
					continue
				}
				if err := tuiState.client.PlayPreset(presetID); err != nil {
					tuiState.lastAction = getText("error_playing_preset")
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
			tuiState.lastAction = "Presets refreshed"
			
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