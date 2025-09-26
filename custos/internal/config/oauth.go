package config

// OAuthProvider represents OAuth provider configuration
type OAuthProvider struct {
	ClientID     string   `mapstructure:"client_id"`
	ClientSecret string   `mapstructure:"client_secret"`
	RedirectURL  string   `mapstructure:"redirect_url"`
	Scopes       []string `mapstructure:"scopes"`
	AuthURL      string   `mapstructure:"auth_url"`
	TokenURL     string   `mapstructure:"token_url"`
	UserInfoURL  string   `mapstructure:"user_info_url"`
}

// OAuth represents OAuth configuration
type OAuth struct {
	Google    OAuthProvider `mapstructure:"google"`
	GitHub    OAuthProvider `mapstructure:"github"`
	StateKey  string        `mapstructure:"state_key"`  // Secret key for state generation
	StateTTL  int           `mapstructure:"state_ttl"`  // State TTL in seconds
}