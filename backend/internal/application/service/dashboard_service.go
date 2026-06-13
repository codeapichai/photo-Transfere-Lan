package service

import (
	"context"
	"net"

	"phototransferlan/backend/internal/domain/repository"
)

type DashboardService struct {
	uploads  repository.UploadRepository
	settings repository.SettingsRepository
}

type Dashboard struct {
	LocalIP             string `json:"local_ip"`
	ServiceStatus       string `json:"service_status"`
	StorageLocation     string `json:"storage_location"`
	AvailableSpaceBytes int64  `json:"available_space_bytes"`
	UploadURL           string `json:"upload_url"`
	TodayFiles          int64  `json:"today_files"`
	TodayBytes          int64  `json:"today_bytes"`
}

func NewDashboardService(uploads repository.UploadRepository, settings repository.SettingsRepository) *DashboardService {
	return &DashboardService{uploads: uploads, settings: settings}
}

func (s *DashboardService) Get(ctx context.Context) (*Dashboard, error) {
	settings, err := s.settings.Get(ctx)
	if err != nil {
		return nil, err
	}
	files, bytes, err := s.uploads.StatsForToday(ctx)
	if err != nil {
		return nil, err
	}
	ip := localIP()
	return &Dashboard{
		LocalIP:             ip,
		ServiceStatus:       "running",
		StorageLocation:     settings.UploadDirectory,
		AvailableSpaceBytes: 0,
		UploadURL:           "http://" + ip + ":8080/upload",
		TodayFiles:          files,
		TodayBytes:          bytes,
	}, nil
}

func localIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok || ipNet.IP.IsLoopback() {
			continue
		}
		if ip := ipNet.IP.To4(); ip != nil {
			return ip.String()
		}
	}
	return "127.0.0.1"
}
