package metadata

import "sync"

// OAuthState holds the OAuth2 client secret, token pair and token callback of a client.
// It is safe for concurrent use; the zero value is ready to use.
// CB is the service's own callback type, so each service keeps its public signature.
type OAuthState[CB any] struct {
	mu           sync.RWMutex // Why: guards every field; Set* and token flows run concurrently with requests.
	clientSecret string
	accessToken  string
	refreshToken string
	onToken      CB
}

// SetClientSecret sets the client secret used by confidential OAuth2 flows.
func (s *OAuthState[CB]) SetClientSecret(secret string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clientSecret = secret
}

// SetAccessToken sets the OAuth2 access token.
func (s *OAuthState[CB]) SetAccessToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = token
}

// SetRefreshToken sets the OAuth2 refresh token.
func (s *OAuthState[CB]) SetRefreshToken(token string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshToken = token
}

// SetTokenCallback sets the callback returned by StoreAccessToken and StoreTokens.
func (s *OAuthState[CB]) SetTokenCallback(cb CB) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onToken = cb
}

// ClientSecret returns the client secret.
func (s *OAuthState[CB]) ClientSecret() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.clientSecret
}

// AccessToken returns the access token.
func (s *OAuthState[CB]) AccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.accessToken
}

// RefreshToken returns the refresh token.
func (s *OAuthState[CB]) RefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refreshToken
}

// RefreshCredentials returns the refresh token and the client secret in one read.
func (s *OAuthState[CB]) RefreshCredentials() (refreshToken, clientSecret string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.refreshToken, s.clientSecret
}

// StoreAccessToken sets the access token and returns the token callback.
// The caller invokes the callback outside the lock, so the callback may call back into the client.
func (s *OAuthState[CB]) StoreAccessToken(accessToken string) CB {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = accessToken
	return s.onToken
}

// StoreTokens sets the access and refresh tokens and returns the token callback.
// The caller invokes the callback outside the lock, so the callback may call back into the client.
func (s *OAuthState[CB]) StoreTokens(accessToken, refreshToken string) CB {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accessToken = accessToken
	s.refreshToken = refreshToken
	return s.onToken
}
