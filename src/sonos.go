package main

import (
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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

// Sonos API Client
type SonosClient struct {
	baseURL   string
	client    *http.Client
	favorites []SonosFavorite
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
