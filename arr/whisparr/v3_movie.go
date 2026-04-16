package whisparr

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// GetAllMovies returns every movie/scene configured in the instance.
func (c *ClientV3) GetAllMovies(ctx context.Context) ([]Movie, error) {
	var out []Movie
	if err := c.base.Get(ctx, "/api/v3/movie", &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMovie returns a single movie/scene by its database ID.
func (c *ClientV3) GetMovie(ctx context.Context, id int) (*Movie, error) {
	var out Movie
	path := fmt.Sprintf("/api/v3/movie/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// AddMovie adds a new movie/scene.
func (c *ClientV3) AddMovie(ctx context.Context, movie *Movie) (*Movie, error) {
	var out Movie
	if err := c.base.Post(ctx, "/api/v3/movie", movie, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UpdateMovie updates an existing movie/scene. Set moveFiles to true to
// relocate files when the path changes.
func (c *ClientV3) UpdateMovie(ctx context.Context, movie *Movie, moveFiles bool) (*Movie, error) {
	var out Movie
	path := fmt.Sprintf("/api/v3/movie/%d?moveFiles=%t", movie.ID, moveFiles)
	if err := c.base.Put(ctx, path, movie, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMovie removes a movie/scene by ID.
func (c *ClientV3) DeleteMovie(ctx context.Context, id int, deleteFiles, addImportExclusion bool) error {
	path := fmt.Sprintf("/api/v3/movie/%d?deleteFiles=%t&addImportExclusion=%t", id, deleteFiles, addImportExclusion)
	return c.base.Delete(ctx, path, nil, nil)
}

// LookupMovie searches for a movie by term.
func (c *ClientV3) LookupMovie(ctx context.Context, term string) ([]Movie, error) {
	var out []Movie
	path := "/api/v3/lookup/movie?term=" + url.QueryEscape(term)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMoviesByPerformer returns all movies associated with a performer foreign ID.
func (c *ClientV3) GetMoviesByPerformer(ctx context.Context, foreignID string) ([]Movie, error) {
	var out []Movie
	path := "/api/v3/movie/listbyperformerforeignid?performerForeignId=" + url.QueryEscape(foreignID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMoviesByStudio returns all movies associated with a studio foreign ID.
func (c *ClientV3) GetMoviesByStudio(ctx context.Context, foreignID string) ([]Movie, error) {
	var out []Movie
	path := "/api/v3/movie/listbystudioforeignid?studioForeignId=" + url.QueryEscape(foreignID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// GetMovieFile returns a single movie file by ID.
func (c *ClientV3) GetMovieFile(ctx context.Context, id int) (*MovieFile, error) {
	var out MovieFile
	path := fmt.Sprintf("/api/v3/moviefile/%d", id)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DeleteMovieFile deletes a movie file by ID.
func (c *ClientV3) DeleteMovieFile(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v3/moviefile/%d", id)
	return c.base.Delete(ctx, path, nil, nil)
}

// EditMovies applies bulk edits to multiple movies.
func (c *ClientV3) EditMovies(ctx context.Context, editor *MovieEditorResource) error {
	return c.base.Put(ctx, "/api/v3/movie/editor", editor, nil)
}

// DeleteMovies deletes multiple movies according to the editor payload.
func (c *ClientV3) DeleteMovies(ctx context.Context, editor *MovieEditorResource) error {
	return c.base.Delete(ctx, "/api/v3/movie/editor", editor, nil)
}

// ImportMovie imports a movie.
func (c *ClientV3) ImportMovie(ctx context.Context, movies []Movie) error {
	return c.base.Post(ctx, "/api/v3/movie/import", movies, nil)
}

// ---------- Movie File Extended ----------.

// GetMovieFiles returns all movie files.
func (c *ClientV3) GetMovieFiles(ctx context.Context, movieID int) ([]MovieFile, error) {
	var out []MovieFile
	path := fmt.Sprintf("/api/v3/moviefile?movieId=%d", movieID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// UpdateMovieFile updates a movie file.
func (c *ClientV3) UpdateMovieFile(ctx context.Context, mf *MovieFile) (*MovieFile, error) {
	var out MovieFile
	if err := c.base.Put(ctx, fmt.Sprintf("/api/v3/moviefile/%d", mf.ID), mf, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// EditMovieFiles bulk edits movie files.
func (c *ClientV3) EditMovieFiles(ctx context.Context, editor *MovieFileEditorResource) error {
	return c.base.Put(ctx, "/api/v3/moviefile/editor", editor, nil)
}

// LookupMovieByTMDB searches for a movie by TMDb ID.
func (c *ClientV3) LookupMovieByTMDB(ctx context.Context, tmdbID int) (*Movie, error) {
	var out Movie
	path := fmt.Sprintf("/api/v3/lookup/movie/tmdb?tmdbId=%d", tmdbID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// LookupMovieByIMDB searches for a movie by IMDb ID.
func (c *ClientV3) LookupMovieByIMDB(ctx context.Context, imdbID string) (*Movie, error) {
	var out Movie
	path := "/api/v3/lookup/movie/imdb?imdbId=" + url.QueryEscape(imdbID)
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ---------- Naming Config ----------.

// GetMovieList returns all movies matching a list of IDs.
func (c *ClientV3) GetMovieList(ctx context.Context, movieIDs []int) ([]Movie, error) {
	var out []Movie
	vals := make(url.Values)
	for _, id := range movieIDs {
		vals.Add("movieIds", strconv.Itoa(id))
	}
	path := "/api/v3/movie?" + vals.Encode()
	if err := c.base.Get(ctx, path, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// BulkDeleteMovieFiles deletes multiple movie files.
func (c *ClientV3) BulkDeleteMovieFiles(ctx context.Context, ids []int) error {
	return c.base.Delete(ctx, "/api/v3/moviefile/bulk", struct {
		MovieFileIDs []int `json:"movieFileIds"`
	}{MovieFileIDs: ids}, nil)
}
