// Package readarr provides a client for the Readarr v1 API.
//
// Readarr is an ebook and audiobook collection manager for Usenet and
// BitTorrent users that monitors multiple RSS feeds for new books and
// automatically grabs, sorts, and renames them. The v1 API applies to
// current versions of the Readarr application.
//
// The [Client] type wraps [arr.BaseClient] and exposes typed methods
// for every major Readarr resource: authors, books, book files,
// editions, calendar, queue, commands, history, and more.
package readarr
