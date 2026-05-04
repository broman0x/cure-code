package version

// [EN] Global versioning for CuRe Code.
// [ID] Penomoran versi global untuk CuRe Code.
const (
	Version    = "1.0.2"
	BuildName  = "Gamba"
	Codename   = "Galileo"
	Author     = "bromanprjkt"
)

// [EN] GetVersion returns the full version string.
// [ID] GetVersion mengembalikan string versi lengkap.
func GetVersion() string {
	return Version + " (" + BuildName + ")"
}
