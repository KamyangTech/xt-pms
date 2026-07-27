package dto

type CheckUpdateResponse struct {
	HasUpdate     bool   `json:"has_update"`
	Current       string `json:"current_version"`
	Latest        string `json:"latest_version"`
	Changelog     string `json:"changelog"`
	DownloadURL   string `json:"download_url"`
	SHA256URL     string `json:"sha256_url"`
	SHA256        string `json:"sha256"`
	FileSize      int64  `json:"file_size"`
	PublishedAt   string `json:"published_at"`
}

// DownloadProgress 下载进度（SSE 推送）
type DownloadProgress struct {
	Percent    int    `json:"percent"`
	Downloaded int64  `json:"downloaded"`
	Total      int64  `json:"total"`
	Status     string `json:"status"` // downloading, verifying, completed, failed
	Message    string `json:"message,omitempty"`
}

// UpdateManifest 更新清单（写入临时文件，传给 update-helper 子进程）
type UpdateManifest struct {
	Action          string `json:"action"`            // update / restart / rollback
	FromVersion     string `json:"from_version"`
	ToVersion       string `json:"to_version"`
	PackagePath     string `json:"package_path"`      // tar.gz 本地路径
	SHA256          string `json:"sha256"`
	InstallDir      string `json:"install_dir"`       // /opt/xt-pms 或容器内 /app
	ServiceName     string `json:"service_name"`      // xt-pms
	ParentPID       int    `json:"parent_pid"`
	BackupDir       string `json:"backup_dir"`
	UpdateMode      string `json:"update_mode"`       // systemd / supervisor
}

type ApplyUpdateResponse struct {
	Message string `json:"message"`
	Manual  bool   `json:"manual"` // true=需要手动操作
	Hint    string `json:"hint,omitempty"`
}
