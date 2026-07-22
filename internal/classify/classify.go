package classify

import (
	"path/filepath"
	"strings"
)

const (
	Video      = "video"
	Audio      = "audio"
	Image      = "image"
	Document   = "document"
	Archive    = "archive"
	Code       = "code"
	Data       = "data"
	Executable = "executable"
	DiskImage  = "disk-image"
	Font       = "font"
	Design3D   = "design-3d"
	Other      = "other"
)

var extensionCategory = buildExtensionMap()

func Category(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if category, ok := extensionCategory[ext]; ok {
		return category
	}
	return Other
}

func buildExtensionMap() map[string]string {
	groups := map[string][]string{
		Video:      {".3g2", ".3gp", ".avi", ".flv", ".m2ts", ".m4v", ".mkv", ".mov", ".mp4", ".mpeg", ".mpg", ".mts", ".ogv", ".ts", ".vob", ".webm", ".wmv"},
		Audio:      {".aac", ".aiff", ".alac", ".ape", ".flac", ".m4a", ".mid", ".midi", ".mp3", ".ogg", ".opus", ".wav", ".wma"},
		Image:      {".avif", ".bmp", ".gif", ".heic", ".heif", ".ico", ".jpeg", ".jpg", ".jxl", ".png", ".psd", ".raw", ".svg", ".tif", ".tiff", ".webp"},
		Document:   {".csv", ".doc", ".docm", ".docx", ".epub", ".key", ".md", ".mobi", ".numbers", ".odp", ".ods", ".odt", ".pages", ".pdf", ".ppt", ".pptx", ".rtf", ".tex", ".txt", ".xls", ".xlsm", ".xlsx"},
		Archive:    {".7z", ".bz2", ".cab", ".gz", ".iso.zip", ".lz", ".lz4", ".rar", ".tar", ".tgz", ".txz", ".xz", ".zip", ".zst"},
		Code:       {".c", ".cc", ".cpp", ".cs", ".css", ".dart", ".go", ".h", ".hpp", ".html", ".java", ".js", ".jsx", ".kt", ".lua", ".php", ".ps1", ".py", ".rb", ".rs", ".scss", ".sh", ".sql", ".swift", ".tsx", ".vue", ".xml", ".yaml", ".yml"},
		Data:       {".arrow", ".avro", ".db", ".db3", ".feather", ".json", ".mdb", ".ndjson", ".orc", ".parquet", ".sqlite", ".sqlite3"},
		Executable: {".apk", ".appimage", ".bat", ".bin", ".cmd", ".com", ".deb", ".dll", ".dmg.exe", ".exe", ".jar", ".msi", ".pkg", ".rpm", ".so"},
		DiskImage:  {".dmg", ".img", ".iso", ".qcow", ".qcow2", ".vdi", ".vhd", ".vhdx", ".vmdk"},
		Font:       {".eot", ".otf", ".ttc", ".ttf", ".woff", ".woff2"},
		Design3D:   {".3ds", ".blend", ".dwg", ".dxf", ".fbx", ".gltf", ".glb", ".obj", ".scad", ".skp", ".step", ".stl"},
	}
	result := make(map[string]string, 180)
	for category, extensions := range groups {
		for _, extension := range extensions {
			result[extension] = category
		}
	}
	return result
}
