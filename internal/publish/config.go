package publish

type Config struct {
	DistsDir      string               `yaml:"distsDir"`
	PoolDir       string               `yaml:"poolDir"`
	Distributions []ConfigDistribution `yaml:"distributions"`
}

type ConfigDistribution struct {
	Name       string            `yaml:"name"`
	Components []ConfigComponent `yaml:"components"`
}

type ConfigComponent struct {
	Name         string   `yaml:"name"`
	FilePatterns []string `yaml:"filePatterns"`
	KeepVersions int      `yaml:"keepVersions"`
}
