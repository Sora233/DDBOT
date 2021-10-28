package version

type MigrationFunc func() error

var migrationMap = map[string]MigrationFunc{
	"0": V1,
}

func DoMigration() error {
	// TODO
	panic("not impl")
}
