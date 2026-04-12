// Package screenscraper provides a client for the Screenscraper API v2.
//
// Screenscraper is a database of video game metadata, media, and ROM information
// primarily used by emulation frontends. Authentication requires developer
// credentials (devid/devpassword) and optionally user credentials (ssid/sspassword).
//
// Usage:
//
//	c := screenscraper.New("devid", "devpassword", "myapp",
//		screenscraper.WithUser("user", "pass"))
//	game, err := c.GetGameInfo(context.Background(), &screenscraper.GameInfoOptions{
//		CRC: "ABCD1234",
//	})
package screenscraper
