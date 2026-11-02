package main

import (
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

var subtitleLanguages = []string{"English", "French", "Spanish", "Portuguese", "Arabic", "Hindi", "German", "Japanese", "Chinese", "Korean", "Italian", "Dutch"}
var languageCodes = map[string]string{"English": "eng", "French": "fre", "Spanish": "spa", "Portuguese": "por", "Arabic": "ara", "Hindi": "hin", "German": "ger", "Japanese": "jpn", "Chinese": "chi", "Korean": "kor", "Italian": "ita", "Dutch": "dut"}

func sameLanguage(code, language string) bool {
	code = strings.ToLower(code)
	want := languageCodes[language]
	if code == want {
		return true
	}
	aliases := map[string]string{"en": "eng", "fr": "fre", "fra": "fre", "es": "spa", "pt": "por", "pob": "por", "pt-br": "por", "de": "ger", "deu": "ger", "ar": "ara", "hi": "hin", "ja": "jpn", "zh": "chi", "zho": "chi", "ko": "kor", "it": "ita", "nl": "dut", "nld": "dut"}
	return aliases[code] == want && want != ""
}

type videoIdentity struct {
	Query, Type, Year, ID string
	Season, Episode       int
}

var episodePattern = regexp.MustCompile(`(?i)\bS(\d{1,2})[ ._-]*E(\d{1,3})\b`)
var shortEpisodePattern = regexp.MustCompile(`(?i)\b(\d{1,2})x(\d{1,3})\b`)
var yearPattern = regexp.MustCompile(`\b(19\d{2}|20\d{2})\b`)
var imdbPattern = regexp.MustCompile(`\btt\d{6,10}\b`)

func identifyVideo(path, query string) videoIdentity {
	raw := videoSearchTitle(path)
	id := videoIdentity{Type: "movie", ID: imdbPattern.FindString(strings.ToLower(path))}
	episode := episodePattern.FindStringSubmatchIndex(raw)
	pattern := episodePattern
	if episode == nil {
		episode = shortEpisodePattern.FindStringSubmatchIndex(raw)
		pattern = shortEpisodePattern
	}
	if episode != nil {
		parts := pattern.FindStringSubmatch(raw)
		id.Season, _ = strconv.Atoi(parts[1])
		id.Episode, _ = strconv.Atoi(parts[2])
		id.Type = "series"
		raw = raw[:episode[0]]
	}
	if query != "" {
		raw = query
		for _, p := range []*regexp.Regexp{episodePattern, shortEpisodePattern} {
			if match := p.FindStringSubmatchIndex(raw); match != nil {
				parts := p.FindStringSubmatch(raw)
				id.Season, _ = strconv.Atoi(parts[1])
				id.Episode, _ = strconv.Atoi(parts[2])
				id.Type = "series"
				raw = raw[:match[0]]
				break
			}
		}
	}
	years := yearPattern.FindAllStringIndex(raw, -1)
	for i := len(years) - 1; i >= 0; i-- {
		year := years[i]
		prefix := raw[:year[0]]
		suffix := strings.Trim(raw[year[1]:], " ()[]{}")
		if normalizedTitle(prefix) != "" && suffix == "" {
			id.Year = raw[year[0]:year[1]]
			raw = prefix
			break
		}
	}
	raw = imdbPattern.ReplaceAllString(raw, "")
	raw = strings.NewReplacer("(", " ", ")", " ", "[", " ", "]", " ", "_", " ", ".", " ").Replace(raw)
	id.Query = strings.Join(strings.Fields(raw), " ")
	return id
}
func normalizedTitle(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return unicode.ToLower(r)
		}
		return -1
	}, s)
}

type subtitleTitle struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Year string `json:"releaseInfo"`
	Type string `json:"type"`
}

func titleConfidence(title subtitleTitle, id videoIdentity) int {
	score := 0
	if normalizedTitle(title.Name) == normalizedTitle(id.Query) {
		score = 100
	} else {
		a, b := strings.Fields(strings.ToLower(title.Name)), strings.Fields(strings.ToLower(id.Query))
		common := 0
		for _, word := range b {
			for _, other := range a {
				if word == other {
					common++
					break
				}
			}
		}
		score = common * 70 / max(1, max(len(a), len(b)))
	}
	if id.Year != "" {
		if strings.HasPrefix(title.Year, id.Year) {
			score += 30
		} else {
			score -= 60
		}
	}
	return score
}
func certainTitle(titles []subtitleTitle, id videoIdentity) bool {
	return len(titles) > 0 && titleConfidence(titles[0], id) >= 100 && (len(titles) == 1 || titleConfidence(titles[0], id)-titleConfidence(titles[1], id) >= 20)
}

type onlineSubtitle struct {
	ID       string  `json:"id"`
	URL      string  `json:"url"`
	Language string  `json:"lang"`
	Label    string  `json:"label"`
	Filename string  `json:"subtitleFileName"`
	Release  string  `json:"movieReleaseName"`
	Match    string  `json:"m"`
	FPS      float64 `json:"fpsMilli"`
	Season   int     `json:"season"`
	Episode  int     `json:"episode"`
	Score    int     `json:"-"`
}

func (s onlineSubtitle) name() string {
	if s.Release != "" {
		return s.Release
	}
	if s.Filename != "" {
		return s.Filename
	}
	if s.Label != "" {
		return s.Label
	}
	return "OpenSubtitles " + s.ID
}
func rankSubtitle(s onlineSubtitle, path string, fps float64) int {
	score := 0
	if s.Match == "h" {
		score += 200
	}
	filename := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	release := s.Release + " " + s.Filename
	if normalizedTitle(filename) == normalizedTitle(s.Release) || normalizedTitle(filename) == normalizedTitle(strings.TrimSuffix(s.Filename, filepath.Ext(s.Filename))) {
		score += 100
	}
	words := strings.FieldsFunc(strings.ToLower(filename), func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsNumber(r) })
	other := strings.ToLower(release)
	for _, word := range words {
		if len(word) > 2 && strings.Contains(other, word) {
			score += 3
		}
	}
	for _, quality := range []string{"bluray", "webrip", "webdl", "hdtv", "1080p", "720p", "2160p"} {
		a, b := strings.Contains(normalizedTitle(filename), quality), strings.Contains(normalizedTitle(release), quality)
		if a && b {
			score += 12
		} else if a != b {
			score -= 3
		}
	}
	if strings.Contains(other, "telesync") || strings.Contains(other, ".ts.") || strings.Contains(other, ".cam.") {
		score -= 50
	}
	if fps > 0 && s.FPS > 0 {
		if math.Abs(fps-s.FPS/1000) < .05 {
			score += 15
		} else {
			score -= 20
		}
	}
	if strings.Contains(other, "cd1") || strings.Contains(other, "cd2") || strings.Contains(other, "1 of 2") {
		score -= 60
	}
	return score
}

type subtitleService struct {
	http               *http.Client
	metadata, provider string
	testHTTP           bool
}

var errSubtitleRateLimit = errors.New("Subtitle service is rate-limiting requests. Try again later.")

func newSubtitleService() *subtitleService {
	service := &subtitleService{metadata: "https://v3-cinemeta.strem.io", provider: "https://opensubtitles-v3.strem.io"}
	service.http = &http.Client{Timeout: 25 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
			return errors.New("too many subtitle redirects")
		}
		return service.checkURL(req.URL)
	}}
	return service
}
func (s *subtitleService) checkURL(u *url.URL) error {
	if s.testHTTP && u.Scheme == "http" {
		return nil
	}
	host := strings.ToLower(u.Hostname())
	if u.Scheme != "https" || u.User != nil || u.Port() != "" && u.Port() != "443" {
		return errors.New("unsupported subtitle download address")
	}
	for _, domain := range []string{"strem.io", "opensubtitles.org", "opensubtitles.com"} {
		if host == domain || strings.HasSuffix(host, "."+domain) {
			return nil
		}
	}
	return errors.New("unrecognized subtitle download host")
}
func (s *subtitleService) get(ctx context.Context, target string, limit int64) ([]byte, error) {
	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}
	if err = s.checkURL(u); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "mswplayr/0.4.0")
	resp, err := s.http.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.New("Subtitle service could not be reached. Try again.")
	}
	defer resp.Body.Close()
	if resp.StatusCode == 429 {
		return nil, errSubtitleRateLimit
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Subtitle service returned HTTP %d.", resp.StatusCode)
	}
	return readBounded(resp.Body, limit)
}
func readBounded(reader io.Reader, limit int64) ([]byte, error) {
	b, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err == nil && int64(len(b)) > limit {
		return nil, errors.New("Subtitle response is too large.")
	}
	return b, err
}
func (s *subtitleService) titles(ctx context.Context, id videoIdentity) ([]subtitleTitle, error) {
	if id.ID != "" {
		return []subtitleTitle{{ID: id.ID, Name: id.Query, Year: id.Year, Type: id.Type}}, nil
	}
	if len(id.Query) < 2 {
		return nil, errors.New("Enter a movie or series title.")
	}
	target := s.metadata + "/catalog/" + id.Type + "/top/search=" + url.PathEscape(id.Query) + ".json"
	b, err := s.get(ctx, target, 4<<20)
	if err != nil {
		return nil, err
	}
	var response struct {
		Metas []subtitleTitle `json:"metas"`
	}
	if json.Unmarshal(b, &response) != nil {
		return nil, errors.New("Subtitle title search returned an invalid response.")
	}
	var titles []subtitleTitle
	for _, title := range response.Metas {
		if imdbPattern.MatchString(title.ID) {
			title.Type = id.Type
			titles = append(titles, title)
		}
	}
	sort.SliceStable(titles, func(i, j int) bool { return titleConfidence(titles[i], id) > titleConfidence(titles[j], id) })
	if len(titles) > 12 {
		titles = titles[:12]
	}
	return titles, nil
}
func videoFingerprint(path string) (string, int64) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || info.Size() < 131072 {
		return "", 0
	}
	sum := uint64(info.Size())
	buf := make([]byte, 65536)
	for _, offset := range []int64{0, info.Size() - 65536} {
		if _, err = f.ReadAt(buf, offset); err != nil {
			return "", 0
		}
		for i := 0; i < len(buf); i += 8 {
			sum += binary.LittleEndian.Uint64(buf[i : i+8])
		}
	}
	return fmt.Sprintf("%016x", sum), info.Size()
}
func (s *subtitleService) search(ctx context.Context, title subtitleTitle, id videoIdentity, path, language string, fps float64) ([]onlineSubtitle, error) {
	videoID := title.ID
	if id.Type == "series" {
		if id.Episode == 0 {
			return nil, errors.New("Include the season and episode, for example S01E03.")
		}
		videoID += fmt.Sprintf(":%d:%d", id.Season, id.Episode)
	}
	extra := url.Values{"filename": {filepath.Base(path)}}
	if hash, size := videoFingerprint(path); hash != "" {
		extra.Set("videoHash", hash)
		extra.Set("videoSize", strconv.FormatInt(size, 10))
	}
	target := s.provider + "/subtitles/" + id.Type + "/" + videoID + "/" + extra.Encode() + ".json"
	b, err := s.get(ctx, target, 4<<20)
	if err != nil {
		return nil, err
	}
	var response struct {
		Subtitles []onlineSubtitle `json:"subtitles"`
	}
	if json.Unmarshal(b, &response) != nil {
		return nil, errors.New("Subtitle search returned an invalid response.")
	}
	var matches []onlineSubtitle
	seen := map[string]bool{}
	for _, sub := range response.Subtitles {
		if !sameLanguage(sub.Language, language) || sub.URL == "" || seen[sub.URL] {
			continue
		}
		if id.Type == "series" && (sub.Season > 0 && sub.Season != id.Season || sub.Episode > 0 && sub.Episode != id.Episode) {
			continue
		}
		u, err := url.Parse(sub.URL)
		if err != nil || s.checkURL(u) != nil {
			continue
		}
		sub.Score = rankSubtitle(sub, path, fps)
		matches = append(matches, sub)
		seen[sub.URL] = true
	}
	sort.SliceStable(matches, func(i, j int) bool { return matches[i].Score > matches[j].Score })
	if len(matches) > 40 {
		matches = matches[:40]
	}
	return matches, nil
}

type cachedSubtitle struct{ Path, Title, Language string }

func subtitleCacheKey(path, language string) string {
	info, err := os.Stat(path)
	signature := path
	if err == nil {
		signature = fmt.Sprintf("%s:%d:%d", historyKey(path), info.Size(), info.ModTime().UnixNano())
	}
	return fmt.Sprintf("%x", sha256.Sum256([]byte(signature+":"+language)))[:32]
}
func subtitleCacheRead(dir, path, language string) (cachedSubtitle, bool) {
	var cached cachedSubtitle
	data, err := os.ReadFile(filepath.Join(dir, subtitleCacheKey(path, language)+".json"))
	if err != nil || json.Unmarshal(data, &cached) != nil {
		return cached, false
	}
	if filepath.Base(cached.Path) != cached.Path {
		return cached, false
	}
	full := filepath.Join(dir, cached.Path)
	info, err := os.Stat(full)
	if err != nil || info.Size() == 0 {
		return cached, false
	}
	cached.Path = full
	return cached, true
}
func (s *subtitleService) download(ctx context.Context, dir, path, language string, sub onlineSubtitle) (cachedSubtitle, error) {
	data, err := s.get(ctx, sub.URL, 8<<20)
	if err != nil {
		return cachedSubtitle{}, err
	}
	data, ext, err := unpackSubtitle(data, path)
	if err != nil {
		return cachedSubtitle{}, err
	}
	if ctx.Err() != nil {
		return cachedSubtitle{}, ctx.Err()
	}
	if err = os.MkdirAll(dir, 0700); err != nil {
		return cachedSubtitle{}, err
	}
	key := subtitleCacheKey(path, language)
	name := fmt.Sprintf("%s-%x%s", key, sha256.Sum256([]byte(sub.URL)), ext)
	cached := cachedSubtitle{Path: name, Title: sub.name(), Language: language}
	if err = atomicSubtitleWrite(filepath.Join(dir, name), data); err != nil {
		return cached, err
	}
	meta, _ := json.Marshal(cached)
	if err = atomicSubtitleWrite(filepath.Join(dir, key+".json"), meta); err != nil {
		return cached, err
	}
	cached.Path = filepath.Join(dir, name)
	return cached, nil
}
func atomicSubtitleWrite(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), "subtitle-")
	if err != nil {
		return err
	}
	temp := f.Name()
	defer os.Remove(temp)
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(temp, path)
}

var timecodePattern = regexp.MustCompile(`(?m)\d{1,2}:\d{2}:\d{2}[,.]\d{2,3}\s*-->`)

func unpackSubtitle(data []byte, video string) ([]byte, string, error) {
	if len(data) > 2 && data[0] == 0x1f && data[1] == 0x8b {
		reader, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, "", err
		}
		data, err = readBounded(reader, 4<<20)
		reader.Close()
		if err != nil {
			return nil, "", err
		}
	}
	if bytes.HasPrefix(data, []byte("PK\x03\x04")) {
		archive, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
		if err != nil {
			return nil, "", err
		}
		if len(archive.File) > 100 {
			return nil, "", errors.New("Subtitle archive has too many files.")
		}
		id := identifyVideo(video, "")
		var chosen *zip.File
		best := -1
		for _, file := range archive.File {
			ext := strings.ToLower(filepath.Ext(file.Name))
			if file.FileInfo().IsDir() || ext != ".srt" && ext != ".vtt" && ext != ".ass" && ext != ".ssa" {
				continue
			}
			score := 0
			if normalizedTitle(strings.TrimSuffix(filepath.Base(video), filepath.Ext(video))) == normalizedTitle(strings.TrimSuffix(filepath.Base(file.Name), ext)) {
				score += 100
			}
			if id.Episode > 0 {
				entryID := identifyVideo(file.Name, "")
				if entryID.Episode != id.Episode || entryID.Season != id.Season {
					continue
				}
				score += 50
			}
			if score > best {
				chosen, best = file, score
			}
		}
		if chosen == nil || chosen.UncompressedSize64 > 4<<20 {
			return nil, "", errors.New("No matching subtitle file in the download.")
		}
		reader, err := chosen.Open()
		if err != nil {
			return nil, "", err
		}
		data, err = readBounded(reader, 4<<20)
		reader.Close()
		if err != nil {
			return nil, "", err
		}
	}
	if len(data) > 4<<20 {
		return nil, "", errors.New("Subtitle file is too large.")
	}
	if len(data) >= 2 && (bytes.HasPrefix(data, []byte{0xff, 0xfe}) || bytes.HasPrefix(data, []byte{0xfe, 0xff})) {
		var order binary.ByteOrder = binary.LittleEndian
		if data[0] == 0xfe {
			order = binary.BigEndian
		}
		units := []uint16{}
		for i := 2; i+1 < len(data); i += 2 {
			units = append(units, order.Uint16(data[i:i+2]))
		}
		data = []byte(string(utf16.Decode(units)))
	} else if !utf8.Valid(data) {
		data, _ = charmap.Windows1252.NewDecoder().Bytes(data)
	}
	text := strings.TrimSpace(strings.TrimPrefix(string(data), "\ufeff"))
	if strings.Contains(strings.ToLower(text), "<html") {
		return nil, "", errors.New("Provider returned a web page instead of subtitles.")
	}
	if strings.HasPrefix(text, "WEBVTT") && strings.Contains(text, "-->") {
		return []byte(text), ".vtt", nil
	}
	if strings.Contains(text, "[Events]") && strings.Contains(text, "Dialogue:") {
		return []byte(text), ".ass", nil
	}
	if timecodePattern.MatchString(text) {
		return []byte(text), ".srt", nil
	}
	return nil, "", errors.New("The download did not contain valid subtitle cues.")
}
