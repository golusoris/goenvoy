// Package jackett provides a client for the Jackett API.
//
// Jackett is a Torznab/Newznab proxy that translates queries from
// *arr applications to tracker site queries.
// Authentication uses an API key passed as a query parameter.
//
// Usage:
//
//	c, err := jackett.New("http://localhost:9117", "your-api-key")
//	if err != nil {
//		log.Fatal(err)
//	}
//	results, err := c.Search(context.Background(), "ubuntu", nil)
package jackett
