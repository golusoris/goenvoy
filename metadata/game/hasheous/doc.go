// Package hasheous provides a client for the Hasheous API.
//
// Hasheous is a ROM hash lookup service that identifies games by their file
// hashes (MD5, SHA1, SHA256, CRC). It aggregates signature data from multiple
// sources including TOSEC, No-Intro, Redump, and RetroAchievements.
//
// No authentication is required for hash lookups or platform listing.
//
// Usage:
//
//	c := hasheous.New()
//	result, err := c.LookupBySHA1(context.Background(), "da39a3ee5e6b4b0d3255bfef95601890afd80709")
package hasheous
