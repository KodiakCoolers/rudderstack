package backendconfig

// ExtractExcludedBackupSourceIds scans a configuration and extracts
// source ids which have the skipBackup config option set to true
func ExtractExcludedBackupSourceIds(c ConfigT) []string {
	r := make([]string, 0)
	for _, source := range c.Sources {
		if source.Config["skipBackup"] == true {
			r = append(r, source.ID)
		}
	}
	return r
}
