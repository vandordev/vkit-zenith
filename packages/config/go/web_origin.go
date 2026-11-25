package config

// LoadWebOrigin selects the Web origin and its environment validation context.
func LoadWebOrigin(options Options) (WebOrigin, error) {
	d, err := loadDocument(options, "webOrigin")
	if err != nil {
		return "", err
	}
	var app AppConfig
	if app.Environment, err = d.stringAt("app", "environment"); err != nil {
		return "", err
	}
	origin, err := d.stringAt("webApi", "identity", "authOrigin")
	if err != nil {
		return "", err
	}
	development, err := d.authBool("webApi", "identity", "development")
	if err != nil {
		return "", err
	}
	if err := d.authOrigins(app, development, origin); err != nil {
		return "", err
	}
	return WebOrigin(origin), nil
}
