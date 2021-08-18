package image_pool

type Image interface {
	Content() ([]byte, error)
}

type Option map[string]interface{}

type OptionFunc func(option Option) Option

func NumOption(num int) OptionFunc {
	return func(option Option) Option {
		option["num"] = num
		return option
	}
}

type Pool interface {
	Get(...OptionFunc) ([]Image, error)
}
