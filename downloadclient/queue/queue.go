package queue

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/golusoris/goenvoy/downloadclient/deluge/v2"
	"github.com/golusoris/goenvoy/downloadclient/nzbget/v2"
	"github.com/golusoris/goenvoy/downloadclient/qbit/v2"
	"github.com/golusoris/goenvoy/downloadclient/rtorrent/v2"
	"github.com/golusoris/goenvoy/downloadclient/sabnzbd/v2"
	"github.com/golusoris/goenvoy/downloadclient/transmission/v2"
)

// ErrUnsupportedKind is returned when a Target specifies an unknown download client Kind.
var ErrUnsupportedKind = errors.New("unsupported download client kind")

const (
	transmissionDownloadWait = 3
	transmissionDownloading  = 4
)

// Take takes a point-in-time snapshot of the queue from the specified Target.
func Take(ctx context.Context, t Target, opts ...Option) (Snapshot, error) {
	o := options{
		timeout:    defaultTimeout,
		httpClient: &http.Client{Transport: sharedTransport},
	}
	for _, opt := range opts {
		opt(&o)
	}

	if o.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, o.timeout)
		defer cancel()
	}

	switch t.Kind {
	case KindQBittorrent:
		return takeQBittorrent(ctx, t, &o)
	case KindTransmission:
		return takeTransmission(ctx, t, &o)
	case KindDeluge:
		return takeDeluge(ctx, t, &o)
	case KindRTorrent:
		return takeRTorrent(ctx, t, &o)
	case KindSABnzbd:
		return takeSABnzbd(ctx, t, &o)
	case KindNZBGet:
		return takeNZBGet(ctx, t, &o)
	default:
		return Snapshot{}, fmt.Errorf("queue: %s: %w", t.Kind, ErrUnsupportedKind)
	}
}

func cloneHTTPClient(c *http.Client) *http.Client {
	if c == nil {
		return nil
	}
	cp := *c
	if cp.Jar == nil {
		jar, err := cookiejar.New(nil)
		if err == nil {
			cp.Jar = jar
		}
	}
	return &cp
}

func qbitOptions(o *options) []qbit.Option {
	opts := make([]qbit.Option, 0, 3)
	if o.httpClient != nil {
		opts = append(opts, qbit.WithHTTPClient(cloneHTTPClient(o.httpClient)))
	}
	if o.timeout > 0 {
		opts = append(opts, qbit.WithTimeout(o.timeout))
	}
	if o.userAgent != "" {
		opts = append(opts, qbit.WithUserAgent(o.userAgent))
	}
	return opts
}

func takeQBittorrent(ctx context.Context, t Target, o *options) (Snapshot, error) {
	cl, err := qbit.New(t.BaseURL, qbitOptions(o)...)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: client: %w", t.Kind, err)
	}
	if t.Username != "" || t.Secret != "" {
		if err := cl.Login(ctx, t.Username, t.Secret); err != nil {
			return Snapshot{}, fmt.Errorf("queue: %s: login: %w", t.Kind, err)
		}
	}
	version, err := cl.Version(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: version: %w", t.Kind, err)
	}
	torrents, err := cl.ListTorrents(ctx, &qbit.ListOptions{})
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: torrents: %w", t.Kind, err)
	}

	snap := Snapshot{Kind: t.Kind, Version: version}
	countQBittorrent(torrents, &snap)
	return snap, nil
}

func countQBittorrent(torrents []qbit.Torrent, snap *Snapshot) {
	for _, tr := range torrents {
		snap.DownloadRate += tr.DlSpeed
		snap.UploadRate += tr.UpSpeed
		switch tr.State {
		case "downloading", "metaDL", "forcedDL", "forcedMetaDL", "stalledDL", "checkingDL":
			snap.Active++
		case "queuedDL":
			snap.Queued++
		}
	}
}

func transmissionTarget(baseURL string) (string, string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", "", fmt.Errorf("parse url: %w", err)
	}
	path := strings.Trim(u.Path, "/")
	u.Path, u.RawPath = "", ""
	if path == "" {
		return u.String(), "", nil
	}
	if !strings.HasSuffix(path, "/rpc") && path != "rpc" {
		path += "/rpc"
	}
	return u.String(), "/" + path, nil
}

func transmissionOptions(o *options, username, secret, rpcPath string) []transmission.Option {
	opts := make([]transmission.Option, 0, 5)
	if o.httpClient != nil {
		opts = append(opts, transmission.WithHTTPClient(cloneHTTPClient(o.httpClient)))
	}
	if o.timeout > 0 {
		opts = append(opts, transmission.WithTimeout(o.timeout))
	}
	if o.userAgent != "" {
		opts = append(opts, transmission.WithUserAgent(o.userAgent))
	}
	if username != "" {
		opts = append(opts, transmission.WithAuth(username, secret))
	}
	if rpcPath != "" {
		opts = append(opts, transmission.WithRPCPath(rpcPath))
	}
	return opts
}

func takeTransmission(ctx context.Context, t Target, o *options) (Snapshot, error) {
	base, rpcPath, err := transmissionTarget(t.BaseURL)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: client: %w", t.Kind, err)
	}

	cl, err := transmission.New(base, transmissionOptions(o, t.Username, t.Secret, rpcPath)...)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: client: %w", t.Kind, err)
	}
	session, err := cl.GetSession(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: session: %w", t.Kind, err)
	}
	torrents, err := cl.GetTorrents(ctx, nil)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: torrents: %w", t.Kind, err)
	}

	snap := Snapshot{Kind: t.Kind, Version: session.Version}
	countTransmission(torrents, &snap)
	return snap, nil
}

func countTransmission(torrents []transmission.Torrent, snap *Snapshot) {
	for _, tr := range torrents {
		snap.DownloadRate += tr.RateDownload
		snap.UploadRate += tr.RateUpload
		switch tr.Status {
		case transmissionDownloading:
			snap.Active++
		case transmissionDownloadWait:
			snap.Queued++
		}
	}
}

func delugeOptions(o *options) []deluge.Option {
	opts := make([]deluge.Option, 0, 3)
	if o.httpClient != nil {
		opts = append(opts, deluge.WithHTTPClient(cloneHTTPClient(o.httpClient)))
	}
	if o.timeout > 0 {
		opts = append(opts, deluge.WithTimeout(o.timeout))
	}
	if o.userAgent != "" {
		opts = append(opts, deluge.WithUserAgent(o.userAgent))
	}
	return opts
}

func takeDeluge(ctx context.Context, t Target, o *options) (Snapshot, error) {
	cl, err := deluge.New(t.BaseURL, delugeOptions(o)...)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: client: %w", t.Kind, err)
	}
	if err := cl.Login(ctx, t.Secret); err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: login: %w", t.Kind, err)
	}
	version, err := cl.GetVersion(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: version: %w", t.Kind, err)
	}
	torrents, err := cl.GetTorrentsStatus(ctx, map[string]string{})
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: torrents: %w", t.Kind, err)
	}

	snap := Snapshot{Kind: t.Kind, Version: version}
	countDeluge(torrents, &snap)
	return snap, nil
}

func countDeluge(torrents map[string]*deluge.Torrent, snap *Snapshot) {
	for _, tr := range torrents {
		if tr == nil {
			continue
		}
		snap.DownloadRate += tr.DownloadPayloadRate
		snap.UploadRate += tr.UploadPayloadRate
		switch tr.State {
		case "Downloading":
			snap.Active++
		case "Queued":
			snap.Queued++
		}
	}
}

func rtorrentOptions(o *options, username, secret string) []rtorrent.Option {
	opts := make([]rtorrent.Option, 0, 4)
	if o.httpClient != nil {
		opts = append(opts, rtorrent.WithHTTPClient(cloneHTTPClient(o.httpClient)))
	}
	if o.timeout > 0 {
		opts = append(opts, rtorrent.WithTimeout(o.timeout))
	}
	if o.userAgent != "" {
		opts = append(opts, rtorrent.WithUserAgent(o.userAgent))
	}
	if username != "" {
		opts = append(opts, rtorrent.WithAuth(username, secret))
	}
	return opts
}

func takeRTorrent(ctx context.Context, t Target, o *options) (Snapshot, error) {
	cl, err := rtorrent.New(t.BaseURL, rtorrentOptions(o, t.Username, t.Secret)...)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: client: %w", t.Kind, err)
	}
	info, err := cl.GetSystemInfo(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: system info: %w", t.Kind, err)
	}
	torrents, err := cl.GetTorrents(ctx, "main")
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: torrents: %w", t.Kind, err)
	}

	snap := Snapshot{Kind: t.Kind, Version: info.ClientVersion}
	countRTorrent(torrents, &snap)
	return snap, nil
}

func countRTorrent(torrents []rtorrent.Torrent, snap *Snapshot) {
	for _, tr := range torrents {
		snap.DownloadRate += tr.DownRate
		snap.UploadRate += tr.UpRate
		if tr.IsOpen && tr.IsActive && !tr.IsComplete {
			snap.Active++
		}
	}
}

func sabnzbdOptions(o *options) []sabnzbd.Option {
	opts := make([]sabnzbd.Option, 0, 3)
	if o.httpClient != nil {
		opts = append(opts, sabnzbd.WithHTTPClient(cloneHTTPClient(o.httpClient)))
	}
	if o.timeout > 0 {
		opts = append(opts, sabnzbd.WithTimeout(o.timeout))
	}
	if o.userAgent != "" {
		opts = append(opts, sabnzbd.WithUserAgent(o.userAgent))
	}
	return opts
}

func takeSABnzbd(ctx context.Context, t Target, o *options) (Snapshot, error) {
	cl, err := sabnzbd.New(t.BaseURL, t.Secret, sabnzbdOptions(o)...)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: client: %w", t.Kind, err)
	}
	version, err := cl.GetVersion(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: version: %w", t.Kind, err)
	}
	q, err := cl.GetQueue(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: queue: %w", t.Kind, err)
	}

	snap := Snapshot{
		Kind:         t.Kind,
		Version:      version,
		Paused:       q.Paused || q.PausedAll,
		DownloadRate: ParseSABSpeed(q.Speed),
	}
	countSABnzbd(q.Slots, &snap)
	return snap, nil
}

func countSABnzbd(slots []sabnzbd.Slot, snap *Snapshot) {
	for _, s := range slots {
		switch s.Status {
		case "Downloading":
			snap.Active++
		case "Queued":
			snap.Queued++
		}
	}
}

func nzbgetOptions(o *options) []nzbget.Option {
	opts := make([]nzbget.Option, 0, 3)
	if o.httpClient != nil {
		opts = append(opts, nzbget.WithHTTPClient(cloneHTTPClient(o.httpClient)))
	}
	if o.timeout > 0 {
		opts = append(opts, nzbget.WithTimeout(o.timeout))
	}
	if o.userAgent != "" {
		opts = append(opts, nzbget.WithUserAgent(o.userAgent))
	}
	return opts
}

func takeNZBGet(ctx context.Context, t Target, o *options) (Snapshot, error) {
	cl, err := nzbget.New(t.BaseURL, t.Username, t.Secret, nzbgetOptions(o)...)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: client: %w", t.Kind, err)
	}
	version, err := cl.GetVersion(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: version: %w", t.Kind, err)
	}
	st, err := cl.GetStatus(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: status: %w", t.Kind, err)
	}
	groups, err := cl.ListGroups(ctx)
	if err != nil {
		return Snapshot{}, fmt.Errorf("queue: %s: groups: %w", t.Kind, err)
	}

	snap := Snapshot{
		Kind:         t.Kind,
		Version:      version,
		Paused:       st.DownloadPaused,
		DownloadRate: st.DownloadRate,
	}
	countNZBGet(groups, &snap)
	return snap, nil
}

func countNZBGet(groups []nzbget.Group, snap *Snapshot) {
	for _, grp := range groups {
		switch {
		case grp.ActiveDownloads > 0 || strings.HasPrefix(grp.Status, "DOWNLOADING"):
			snap.Active++
		case grp.Status == "QUEUED":
			snap.Queued++
		}
	}
}

// ParseSABSpeed converts SABnzbd's "1.2 M" style speed (per second, binary prefixes) to
// bytes per second. Unparseable input yields 0, which never causes backpressure.
func ParseSABSpeed(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	num, unit := s, ""
	if i := strings.LastIndexAny(s, "KMGTkmgt"); i >= 0 && i == len(s)-1 {
		num, unit = strings.TrimSpace(s[:i]), strings.ToUpper(s[i:])
	}
	f, err := strconv.ParseFloat(num, 64)
	if err != nil || f < 0 || math.IsInf(f, 0) || math.IsNaN(f) {
		return 0
	}
	return int64(math.Round(f * sabMultiplier(unit)))
}

func sabMultiplier(unit string) float64 {
	switch unit {
	case "K":
		return 1 << 10
	case "M":
		return 1 << 20
	case "G":
		return 1 << 30
	case "T":
		return 1 << 40
	default:
		return 1
	}
}
