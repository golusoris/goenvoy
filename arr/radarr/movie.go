package radarr

import (
	"context"
	"fmt"
	"net/url"
)

// DeleteMovies performs a batch delete of multiple movies.
func (c *Client) DeleteMovies(ctx context.Context, editor *MovieEditorResource) error {
	return c.base.Delete(ctx, "/api/v3/movie/editor", editor, nil)
}

// GetAlternativeTitles returns all alternative titles.
func (c *Client) GetAlternativeTitles(ctx context.Context) ([]AlternativeTitleResource, error) {
	var out []AlternativeTitleResource
	if err := c.base.Get(ctx, "/api/v3/alttitle", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetAlternativeTitle returns a single alternative title by ID.
func (c *Client) GetAlternativeTitle(ctx context.Context, id int) (*AlternativeTitleResource, error) {
	var out AlternativeTitleResource
	if err := c.base.Get(ctx, fmt.Sprintf("/api/v3/alttitle/%d", id), &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetAllMovies returns every movie configured in Radarr.
func (c *Client) GetAllMovies(ctx context.Context) ([]Movie, error) {
	var out []Movie
	if err := c.base.Get(ctx, "/api/v3/movie", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMovie returns a single movie by its database ID.
func (c *Client) GetMovie(ctx context.Context, id int) (*Movie, error) {
	var out Movie
	path := fmt.Sprintf("/api/v3/movie/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddMovie adds a new movie to Radarr.
func (c *Client) AddMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	var out Movie
	if err := c.base.Post(ctx, "/api/v3/movie", movie, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMovie updates an existing movie. Set moveFiles to true to relocate
// files when the movie path changes.
func (c *Client) UpdateMovie(ctx context.Context, movie *Movie, moveFiles bool) (*Movie, error) {
	var out Movie
	path := fmt.Sprintf("/api/v3/movie/%d?moveFiles=%t", movie.ID, moveFiles)
	if err := c.base.Put(ctx, path, movie, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMovie removes a movie. Set deleteFiles to true to also delete
// downloaded movie files from disk.
func (c *Client) DeleteMovie(ctx context.Context, id int, deleteFiles, addImportExclusion bool) error {
	path := fmt.Sprintf("/api/v3/movie/%d?deleteFiles=%t&addImportExclusion=%t", id, deleteFiles, addImportExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupMovie searches for a movie by term (title).
func (c *Client) LookupMovie(ctx context.Context, term string) ([]Movie, error) {
	var out []Movie
	path := "/api/v3/movie/lookup?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// LookupMovieByTmdbID looks up a movie by its TMDb ID.
func (c *Client) LookupMovieByTmdbID(ctx context.Context, tmdbID int) (*Movie, error) {
	var out Movie
	path := fmt.Sprintf("/api/v3/movie/lookup/tmdb?tmdbId=%d", tmdbID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LookupMovieByImdbID looks up a movie by its IMDb ID.
func (c *Client) LookupMovieByImdbID(ctx context.Context, imdbID string) (*Movie, error) {
	var out Movie
	path := "/api/v3/movie/lookup/imdb?imdbId=" + url.QueryEscape(imdbID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// GetMovieFiles returns all movie files for the given movie IDs.
func (c *Client) GetMovieFiles(ctx context.Context, movieID int) ([]MovieFile, error) {
	var out []MovieFile
	path := fmt.Sprintf("/api/v3/moviefile?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMovieFile returns a single movie file by its database ID.
func (c *Client) GetMovieFile(ctx context.Context, id int) (*MovieFile, error) {
	var out MovieFile
	path := fmt.Sprintf("/api/v3/moviefile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMovieFile removes a single movie file by its database ID.
func (c *Client) DeleteMovieFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v3/moviefile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// DeleteMovieFiles removes multiple movie files by their IDs.
func (c *Client) DeleteMovieFiles(ctx context.Context, ids []int) error {
	body := MovieFileListResource{MovieFileIDs: ids}
	return c.base.Delete(ctx, "/api/v3/moviefile/bulk", &body, nil)
}

// EditMovies performs a batch update on multiple movies.
func (c *Client) EditMovies(ctx context.Context, editor *MovieEditorResource) error {
	return c.base.Put(ctx, "/api/v3/movie/editor", editor, nil)
}

// ImportMovies imports one or more movies in bulk.
func (c *Client) ImportMovies(ctx context.Context, movies []Movie) ([]Movie, error) {
	var out []Movie
	if err := c.base.Post(ctx, "/api/v3/movie/import", movies, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Wanted ----------.

// UpdateMovieFile updates an individual movie file's metadata
// (quality, language, etc.).
func (c *Client) UpdateMovieFile(ctx context.Context, mf *MovieFile) (*MovieFile, error) {
	var out MovieFile
	path := fmt.Sprintf("/api/v3/moviefile/%d", mf.ID)
	if err := c.base.Put(ctx, path, mf, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EditMovieFiles performs a bulk update of movie file metadata
// (quality, language, release group).
func (c *Client) EditMovieFiles(ctx context.Context, editor *MovieFileEditorResource) error {
	return c.base.Put(ctx, "/api/v3/moviefile/editor", editor, nil)
}

// ---------- Custom Format Bulk ----------.

// UpdateMovieFilesBulk performs a bulk update of movie file properties.
func (c *Client) UpdateMovieFilesBulk(ctx context.Context, editor *MovieFileEditorResource) ([]MovieFile, error) {
	var out []MovieFile
	if err := c.base.Put(ctx, "/api/v3/moviefile/bulk", editor, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------- Movie Folder ----------.

// GetMovieFolder returns folder information for a movie.
func (c *Client) GetMovieFolder(ctx context.Context, movieID int) error {
	return c.base.Get(ctx, fmt.Sprintf("/api/v3/movie/%d/folder", movieID), nil)
}

// ---------- Naming Examples ----------.
