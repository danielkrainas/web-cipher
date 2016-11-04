package context

func WithVersion(ctx Context, version string) Context {
	return WithValue(ctx, "version", version)
}

func GetVersion(ctx Context) string {
	return GetStringValue(ctx, "version")
}
