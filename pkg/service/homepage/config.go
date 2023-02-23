package homepage

type Config struct {
	BrandName     string `yaml:"brand_name"`
	BrandImageURL string `yaml:"brand_image_url"`
}

func (c *Config) Validate() error {
	return nil
}
