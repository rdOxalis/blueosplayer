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
	LangEnglish Language = "en"
	LangGerman  Language = "de"
	LangSwahili Language = "sw"
)

var currentLanguage = LangEnglish

// Localization texts
var texts = map[Language]map[string]string{
	LangEnglish: {
		"title":                   "🎵 BlueOS Controller",
		"scanning":                "🔍 Scanning network for BlueOS players...",
		"scanning_network":        "   Scanning network: %s",
		"found_player":            "   ✅ Found: %s (%s) at %s",
		"no_players":              "no BlueOS players found",
		"could_not_determine_ip":  "could not determine local IP: %w",
		"available_players":       "📱 Available Players:",
		"select_player":           "Select a player (1-%d): ",
		"invalid_selection":       "❌ Invalid selection",
		"connected_to":            "✅ Connected to: %s (%s)",
		"error_selecting_player":  "Error selecting player: %v",
		"interactive_mode":        "🎵 BlueOS Controller - Interactive Mode",
		"separator":               "=======================================",
		"status_volume":           "📊 Status: %s | Volume: %d%%",
		"error_retrieving_status": "❌ Error retrieving status: %v",
		"available_presets":       "📋 Available Presets:",
		"error_loading_presets":   "❌ Error loading presets: %v",
		"available_commands":      "🎮 Available Commands:",
		"cmd_play_preset":         "  play <preset_id>  - Play preset",
		"cmd_play":                "  play              - Start playback",
		"cmd_pause":               "  pause             - Pause playback",
		"cmd_stop":                "  stop              - Stop playback",
		"cmd_next":                "  next              - Next track",
		"cmd_prev":                "  prev              - Previous track",
		"cmd_volume":              "  volume <0-100>    - Set volume",
		"cmd_vol":                 "  vol <0-100>       - Set volume (short)",
		"cmd_status":              "  status            - Show current status",
		"cmd_presets":             "  presets           - Show presets",
		"cmd_help":                "  help              - Show this help",
		"cmd_lang":                "  lang <en|de|sw>   - Change language",
		"cmd_quit":                "  quit / exit       - Exit program",
		"prompt":                  "Blueos> ",
		"invalid_preset_id":       "❌ Invalid preset ID",
		"error_playing_preset":    "❌ Error playing preset: %v",
		"playing_preset":          "✅ Playing preset %d",
		"error_starting_playback": "❌ Error starting playback: %v",
		"playback_started":        "▶️ Playback started",
		"error_pausing":           "❌ Error pausing: %v",
		"paused":                  "⏸️ Paused",
		"error_stopping":          "❌ Error stopping: %v",
		"stopped":                 "⏹️ Stopped",
		"error_next_track":        "❌ Error skipping to next track: %v",
		"next_track":              "⏭️ Next track",
		"error_prev_track":        "❌ Error going to previous track: %v",
		"prev_track":              "⏮️ Previous track",
		"volume_missing":          "❌ Volume value missing",
		"invalid_volume":          "❌ Invalid volume value",
		"error_setting_volume":    "❌ Error setting volume: %v",
		"volume_set":              "🔊 Volume set to %d%%",
		"language_changed":        "🌍 Language changed to %s",
		"invalid_language":        "❌ Invalid language. Use: en, de, sw",
		"goodbye":                 "👋 Goodbye!",
		"unknown_command":         "❌ Unknown command: %s (Type 'help' for help)",
	},
	LangGerman: {
		"title":                   "🎵 BlueOS Controller",
		"scanning":                "🔍 Suche nach BlueOS Playern im Netzwerk...",
		"scanning_network":        "   Scanne Netzwerk: %s",
		"found_player":            "   ✅ Gefunden: %s (%s) auf %s",
		"no_players":              "keine BlueOS Player gefunden",
		"could_not_determine_ip":  "konnte lokale IP nicht ermitteln: %w",
		"available_players":       "📱 Verfügbare Player:",
		"select_player":           "Wähle einen Player (1-%d): ",
		"invalid_selection":       "❌ Ungültige Auswahl",
		"connected_to":            "✅ Verbunden mit: %s (%s)",
		"error_selecting_player":  "Fehler bei der Player-Auswahl: %v",
		"interactive_mode":        "🎵 BlueOS Controller - Interaktiver Modus",
		"separator":               "==========================================",
		"status_volume":           "📊 Status: %s | Lautstärke: %d%%",
		"error_retrieving_status": "❌ Fehler beim Abrufen des Status: %v",
		"available_presets":       "📋 Verfügbare Presets:",
		"error_loading_presets":   "❌ Fehler beim Laden der Presets: %v",
		"available_commands":      "🎮 Verfügbare Befehle:",
		"cmd_play_preset":         "  play <preset_id>  - Preset abspielen",
		"cmd_play":                "  play              - Wiedergabe starten",
		"cmd_pause":               "  pause             - Pausieren",
		"cmd_stop":                "  stop              - Stoppen",
		"cmd_next":                "  next              - Nächster Titel",
		"cmd_prev":                "  prev              - Vorheriger Titel",
		"cmd_volume":              "  volume <0-100>    - Lautstärke setzen",
		"cmd_vol":                 "  vol <0-100>       - Lautstärke setzen (kurz)",
		"cmd_status":              "  status            - Aktuellen Status anzeigen",
		"cmd_presets":             "  presets           - Presets anzeigen",
		"cmd_help":                "  help              - Diese Hilfe anzeigen",
		"cmd_lang":                "  lang <en|de|sw>   - Sprache ändern",
		"cmd_quit":                "  quit / exit       - Programm beenden",
		"prompt":                  "Blueos> ",
		"invalid_preset_id":       "❌ Ungültige Preset-ID",
		"error_playing_preset":    "❌ Fehler beim Abspielen: %v",
		"playing_preset":          "✅ Preset %d wird abgespielt",
		"error_starting_playback": "❌ Fehler beim Starten: %v",
		"playback_started":        "▶️ Wiedergabe gestartet",
		"error_pausing":           "❌ Fehler beim Pausieren: %v",
		"paused":                  "⏸️ Pausiert",
		"error_stopping":          "❌ Fehler beim Stoppen: %v",
		"stopped":                 "⏹️ Gestoppt",
		"error_next_track":        "❌ Fehler beim Weiterschalten: %v",
		"next_track":              "⏭️ Nächster Titel",
		"error_prev_track":        "❌ Fehler beim Zurückschalten: %v",
		"prev_track":              "⏮️ Vorheriger Titel",
		"volume_missing":          "❌ Lautstärke-Wert fehlt",
		"invalid_volume":          "❌ Ungültiger Lautstärke-Wert",
		"error_setting_volume":    "❌ Fehler beim Setzen der Lautstärke: %v",
		"volume_set":              "🔊 Lautstärke auf %d%% gesetzt",
		"language_changed":        "🌍 Sprache geändert zu %s",
		"invalid_language":        "❌ Ungültige Sprache. Verwende: en, de, sw",
		"goodbye":                 "👋 Auf Wiedersehen!",
		"unknown_command":         "❌ Unbekannter Befehl: %s (Tippe 'help' für Hilfe)",
	},
	LangSwahili: {
		"title":                   "🎵 Kidhibiti cha BlueOS",
		"scanning":                "🔍 Kutafuta vichezaji vya BlueOS kwenye mtandao...",
		"scanning_network":        "   Kutafuta mtandao: %s",
		"found_player":            "   ✅ Kumepatikana: %s (%s) kwa %s",
		"no_players":              "hakuna vichezaji vya BlueOS vilivopatikana",
		"could_not_determine_ip":  "haikuweza kutambua IP ya ndani: %w",
		"available_players":       "📱 Vichezaji Vinavyopatikana:",
		"select_player":           "Chagua kichezaji (1-%d): ",
		"invalid_selection":       "❌ Chaguo batili",
		"connected_to":            "✅ Imeunganishwa na: %s (%s)",
		"error_selecting_player":  "Hitilafu katika kuchagua kichezaji: %v",
		"interactive_mode":        "🎵 Kidhibiti cha BlueOS - Hali ya Maingiliano",
		"separator":               "===========================================",
		"status_volume":           "📊 Hali: %s | Sauti: %d%%",
		"error_retrieving_status": "❌ Hitilafu katika kupata hali: %v",
		"available_presets":       "📋 Mipangilio Inayopatikana:",
		"error_loading_presets":   "❌ Hitilafu katika kupakia mipangilio: %v",
		"available_commands":      "🎮 Amri Zinazopatikana:",
		"cmd_play_preset":         "  play <preset_id>  - Cheza mpangilio",
		"cmd_play":                "  play              - Anza kucheza",
		"cmd_pause":               "  pause             - Simamisha",
		"cmd_stop":                "  stop              - Acha",
		"cmd_next":                "  next              - Wimbo ujao",
		"cmd_prev":                "  prev              - Wimbo uliopita",
		"cmd_volume":              "  volume <0-100>    - Weka sauti",
		"cmd_vol":                 "  vol <0-100>       - Weka sauti (kifupi)",
		"cmd_status":              "  status            - Onyesha hali ya sasa",
		"cmd_presets":             "  presets           - Onyesha mipangilio",
		"cmd_help":                "  help              - Onyesha msaada huu",
		"cmd_lang":                "  lang <en|de|sw>   - Badilisha lugha",
		"cmd_quit":                "  quit / exit       - Toka kwenye programu",
		"prompt":                  "Blueos> ",
		"invalid_preset_id":       "❌ Kitambulisho cha mpangilio si halali",
		"error_playing_preset":    "❌ Hitilafu katika kucheza mpangilio: %v",
		"playing_preset":          "✅ Kucheza mpangilio %d",
		"error_starting_playback": "❌ Hitilafu katika kuanza kucheza: %v",
		"playback_started":        "▶️ Imeanza kucheza",
		"error_pausing":           "❌ Hitilafu katika kusimamisha: %v",
		"paused":                  "⏸️ Imesimamishwa",
		"error_stopping":          "❌ Hitilafu katika kuacha: %v",
		"stopped":                 "⏹️ Imeachwa",
		"error_next_track":        "❌ Hitilafu katika kuruka wimbo ujao: %v",
		"next_track":              "⏭️ Wimbo ujao",
		"error_prev_track":        "❌ Hitilafu katika kurudi wimbo uliopita: %v",
		"prev_track":              "⏮️ Wimbo uliopita",
		"volume_missing":          "❌ Thamani ya sauti inakosekana",
		"invalid_volume":          "❌ Thamani ya sauti si halali",
		"error_setting_volume":    "❌ Hitilafu katika kuweka sauti: %v",
		"volume_set":              "🔊 Sauti imewekwa %d%%",
		"language_changed":        "🌍 Lugha imebadilishwa kuwa %s",
		"invalid_language":        "❌ Lugha si halali. Tumia: en, de, sw",
		"goodbye":                 "👋 Kwaheri!",
		"unknown_command":         "❌ Amri isiyojulikana: %s (Andika 'help' kwa msaada)",
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

func NewBluesoundClient(ip string) *BluesoundClient {
	return &BluesoundClient{
		baseURL: fmt.Sprintf("http://%s:%s", ip, BluesoundPort),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
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

			if player, found := checkForBlueOSPlayer(ip); found {
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

func checkForBlueOSPlayer(ip string) (PlayerInfo, bool) {
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

// Player selection
func selectPlayer() (*BluesoundClient, error) {
	players, err := scanForPlayers()
	if err != nil {
		return nil, err
	}

	if len(players) == 0 {
		return nil, fmt.Errorf(getText("no_players"))
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
		return NewBluesoundClient(selectedPlayer.IP), nil
	}
}

// Show status
func showStatus(client *BluesoundClient) {
	status, err := client.GetStatus()
	if err != nil {
		fmt.Printf(getText("error_retrieving_status")+"\n", err)
		return
	}

	fmt.Printf("\n"+getText("status_volume")+"\n", status.State, status.Volume)
	if status.Song != "" {
		fmt.Printf("🎵 %s", status.Song)
		if status.Artist != "" {
			fmt.Printf(" - %s", status.Artist)
		}
		if status.Album != "" {
			fmt.Printf(" (%s)", status.Album)
		}
		fmt.Println()
	}
}

// Show presets
func showPresets(client *BluesoundClient) {
	presets, err := client.GetPresets()
	if err != nil {
		fmt.Printf(getText("error_loading_presets")+"\n", err)
		return
	}

	fmt.Println("\n" + getText("available_presets"))
	for _, preset := range presets {
		fmt.Printf("  [%d] %s\n", preset.ID, preset.Name)
	}
}

// Show help
func showHelp() {
	fmt.Println("\n" + getText("available_commands"))
	fmt.Println(getText("cmd_play_preset"))
	fmt.Println(getText("cmd_play"))
	fmt.Println(getText("cmd_pause"))
	fmt.Println(getText("cmd_stop"))
	fmt.Println(getText("cmd_next"))
	fmt.Println(getText("cmd_prev"))
	fmt.Println(getText("cmd_volume"))
	fmt.Println(getText("cmd_vol"))
	fmt.Println(getText("cmd_status"))
	fmt.Println(getText("cmd_presets"))
	fmt.Println(getText("cmd_help"))
	fmt.Println(getText("cmd_lang"))
	fmt.Println(getText("cmd_quit"))
}

// Change language
func changeLanguage(lang string) {
	switch strings.ToLower(lang) {
	case "en", "english":
		currentLanguage = LangEnglish
		fmt.Printf(getText("language_changed")+"\n", "English")
	case "de", "german", "deutsch":
		currentLanguage = LangGerman
		fmt.Printf(getText("language_changed")+"\n", "Deutsch")
	case "sw", "swahili", "kiswahili":
		currentLanguage = LangSwahili
		fmt.Printf(getText("language_changed")+"\n", "Kiswahili")
	default:
		fmt.Println(getText("invalid_language"))
	}
}

// Interactive loop
func interactiveMode(client *BluesoundClient) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("\n" + getText("interactive_mode"))
	fmt.Println(getText("separator"))
	showStatus(client)
	showPresets(client)
	showHelp()

	for {
		fmt.Print("\n" + getText("prompt"))
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
					fmt.Println(getText("invalid_preset_id"))
					continue
				}
				if err := client.PlayPreset(presetID); err != nil {
					fmt.Printf(getText("error_playing_preset")+"\n", err)
				} else {
					fmt.Printf(getText("playing_preset")+"\n", presetID)
					time.Sleep(1 * time.Second)
					showStatus(client)
				}
			} else {
				// Start playback
				if err := client.Play(); err != nil {
					fmt.Printf(getText("error_starting_playback")+"\n", err)
				} else {
					fmt.Println(getText("playback_started"))
					time.Sleep(1 * time.Second)
					showStatus(client)
				}
			}

		case "pause":
			if err := client.Pause(); err != nil {
				fmt.Printf(getText("error_pausing")+"\n", err)
			} else {
				fmt.Println(getText("paused"))
				showStatus(client)
			}

		case "stop":
			if err := client.Stop(); err != nil {
				fmt.Printf(getText("error_stopping")+"\n", err)
			} else {
				fmt.Println(getText("stopped"))
				showStatus(client)
			}

		case "next":
			if err := client.Next(); err != nil {
				fmt.Printf(getText("error_next_track")+"\n", err)
			} else {
				fmt.Println(getText("next_track"))
				time.Sleep(1 * time.Second)
				showStatus(client)
			}

		case "prev", "previous":
			if err := client.Previous(); err != nil {
				fmt.Printf(getText("error_prev_track")+"\n", err)
			} else {
				fmt.Println(getText("prev_track"))
				time.Sleep(1 * time.Second)
				showStatus(client)
			}

		case "volume", "vol":
			if len(parts) < 2 {
				fmt.Println(getText("volume_missing"))
				continue
			}
			volume, err := strconv.Atoi(parts[1])
			if err != nil {
				fmt.Println(getText("invalid_volume"))
				continue
			}
			if err := client.SetVolume(volume); err != nil {
				fmt.Printf(getText("error_setting_volume")+"\n", err)
			} else {
				fmt.Printf(getText("volume_set")+"\n", volume)
			}

		case "status":
			showStatus(client)

		case "presets":
			showPresets(client)

		case "help":
			showHelp()

		case "lang", "language":
			if len(parts) < 2 {
				fmt.Println(getText("invalid_language"))
				continue
			}
			changeLanguage(parts[1])

		case "quit", "exit":
			fmt.Println(getText("goodbye"))
			return

		default:
			fmt.Printf(getText("unknown_command")+"\n", command)
		}
	}
}

func main() {
	fmt.Println(getText("title"))
	fmt.Println("====================")

	// Select player
	client, err := selectPlayer()
	if err != nil {
		log.Fatalf(getText("error_selecting_player"), err)
	}

	// Start interactive mode
	interactiveMode(client)
}
