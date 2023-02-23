package homepage

type StatusResponse struct {
	BrandName     string  `json:"brand_name,omitempty"`
	BrandImageURL string  `json:"brand_image_url,omitempty"`
	Version       Version `json:"version"`
}

type Version struct {
	Full      string `json:"full"`
	Short     string `json:"short"`
	Release   string `json:"release"`
	GitCommit string `json:"git_commit"`
}
