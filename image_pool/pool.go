package image_pool

type Image interface {
	Content() ([]byte, error)
}

type Option map[string]interface{}

type OptionFunc func(option Option) Option

type Pool interface {
	Get(...OptionFunc) ([]Image, error)
}
