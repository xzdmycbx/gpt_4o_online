package oauth2

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/oauth2"
)

// TwitterOAuth2Client handles Twitter OAuth2 flow with PKCE
type TwitterOAuth2Client struct {
	config      *oauth2.Config
	redirectURL string
}

// TwitterUserInfo represents Twitter user information
type TwitterUserInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url"`
	Email           string `json:"email,omitempty"`
}

// NewTwitterOAuth2Client creates a new Twitter OAuth2 client
func NewTwitterOAuth2Client(clientID, clientSecret, redirectURL string) *TwitterOAuth2Client {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"tweet.read", "users.read", "offline.access"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://twitter.com/i/oauth2/authorize",
			TokenURL: "https://api.twitter.com/2/oauth2/token",
		},
	}

	return &TwitterOAuth2Client{
		config:      config,
		redirectURL: redirectURL,
	}
}

// GenerateAuthURL generates the OAuth2 authorization URL with PKCE
func (c *TwitterOAuth2Client) GenerateAuthURL(state string) (authURL string, codeVerifier string, err error) {
	// Generate code verifier for PKCE
	codeVerifier, err = generateCodeVerifier()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate code verifier: %w", err)
	}

	// Generate code challenge
	codeChallenge := generateCodeChallenge(codeVerifier)

	// Build authorization URL with PKCE parameters
	authURL = c.config.AuthCodeURL(state,
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	return authURL, codeVerifier, nil
}

// ExchangeCode exchanges authorization code for access token
func (c *TwitterOAuth2Client) ExchangeCode(ctx context.Context, code, codeVerifier string) (*oauth2.Token, error) {
	token, err := c.config.Exchange(ctx, code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	return token, nil
}

// GetUserInfo retrieves user information from Twitter API
func (c *TwitterOAuth2Client) GetUserInfo(ctx context.Context, token *oauth2.Token) (*TwitterUserInfo, error) {
	client := c.config.Client(ctx, token)

	// Twitter API v2 user endpoint
	apiURL := "https://api.twitter.com/2/users/me?user.fields=profile_image_url"
	resp, err := client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("twitter API error: %s", string(body))
	}

	var result struct {
		Data TwitterUserInfo `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}

	return &result.Data, nil
}

// generateCodeVerifier generates a random code verifier for PKCE
func generateCodeVerifier() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generateCodeChallenge generates code challenge from verifier using S256 method
func generateCodeChallenge(verifier string) string {
	// S256: BASE64URL(SHA256(ASCII(code_verifier)))
	h := sha256.New()
	h.Write([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// ValidateState validates OAuth2 state parameter to prevent CSRF
func ValidateState(expectedState, actualState string) bool {
	return expectedState != "" && expectedState == actualState
}

// GenerateState generates a random state parameter for OAuth2
func GenerateState() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// BuildCallbackURL builds the callback URL with error handling
func BuildCallbackURL(baseURL string, params url.Values) string {
	u, _ := url.Parse(baseURL)
	u.RawQuery = params.Encode()
	return u.String()
}
