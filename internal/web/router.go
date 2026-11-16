package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"beszel-qnap/internal/agentctl"
	"beszel-qnap/internal/auth"
	"beszel-qnap/internal/config"
	"beszel-qnap/internal/updater"
)

type SessionVerifier interface {
	VerifyRequest(ctx context.Context, r *http.Request) (*auth.Session, error)
}

type Server struct {
	Mgr      *agentctl.Manager
	Verifier SessionVerifier
}

func NewServer(mgr *agentctl.Manager, verifier SessionVerifier) *Server {
	return &Server{Mgr: mgr, Verifier: verifier}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	session, err := s.Verifier.VerifyRequest(r.Context(), r)
	if err != nil {
		status := http.StatusUnauthorized
		if errors.Is(err, auth.ErrNotAdmin) {
			status = http.StatusForbidden
		}
		writeJSONError(w, status, err.Error())
		return
	}

	action := strings.TrimSpace(r.URL.Query().Get("action"))
	if action == "" {
		trimmed := strings.Trim(r.URL.Path, "/")
		if trimmed != "" && !strings.HasSuffix(trimmed, ".cgi") {
			parts := strings.Split(trimmed, "/")
			action = parts[len(parts)-1]
		}
	}
	if action == "" {
		action = "status"
	}

	switch {
	case r.Method == http.MethodGet && action == "status":
		s.handleStatus(w, r, session)
	case r.Method == http.MethodGet && action == "config":
		s.handleGetConfig(w)
	case r.Method == http.MethodPost && action == "config":
		s.handleSaveConfig(w, r)
	case r.Method == http.MethodPost && action == "service":
		s.handleService(w, r)
	case r.Method == http.MethodPost && action == "fingerprint":
		s.handleFingerprint(w, r)
	case r.Method == http.MethodGet && action == "releases":
		s.handleReleases(w, r)
	case r.Method == http.MethodPost && action == "download":
		s.handleDownloadStart(w, r)
	case r.Method == http.MethodGet && action == "download_status":
		s.handleDownloadStatus(w)
	case r.Method == http.MethodPost && action == "upload":
		s.handleUpload(w, r)
	case r.Method == http.MethodPost && action == "rollback":
		s.handleRollback(w)
	case r.Method == http.MethodGet && action == "logs":
		s.handleGetLogs(w, r)
	case r.Method == http.MethodPost && action == "logs_clear":
		s.handleClearLogs(w, r)
	default:
		writeJSONError(w, http.StatusNotFound, "unknown action or unsupported method")
	}
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request, sess *auth.Session) {
	includeFP := r.URL.Query().Get("fp") == "1"
	st, err := s.Mgr.Status(includeFP)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  st,
		"session": sess,
	})
}

func (s *Server) handleGetConfig(w http.ResponseWriter) {
	cfg, err := s.Mgr.LoadConfig()
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"config": cfg,
		"env":    cfg.AgentEnvMap(s.Mgr.Paths.AgentDataDir()),
	})
}

func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Config  config.Config `json:"config"`
		Restart bool          `json:"restart"`
	}
	if err := decodeJSONBody(r, &payload, 2<<20); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := s.Mgr.SaveConfig(&payload.Config); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	s.Mgr.LogManager("Saved configuration (restart=%v)", payload.Restart)

	var restartErr string
	if payload.Restart && s.Mgr.FindRunningPID() > 0 {
		if err := s.Mgr.Restart(); err != nil {
			restartErr = err.Error()
		}
	}

	st, _ := s.Mgr.Status(false)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":            true,
		"restart_error": restartErr,
		"config":        &payload.Config,
		"status":        st,
	})
}

func (s *Server) handleService(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Op string `json:"op"`
	}
	if err := decodeJSONBody(r, &payload, 64<<10); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	var err error
	switch payload.Op {
	case "start":
		err = s.Mgr.Start()
	case "stop":
		err = s.Mgr.Stop()
	case "restart":
		err = s.Mgr.Restart()
	default:
		writeJSONError(w, http.StatusBadRequest, "invalid service operation (expected start, stop, or restart)")
		return
	}

	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	st, _ := s.Mgr.Status(false)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"status": st,
	})
}

func (s *Server) handleFingerprint(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Op string `json:"op"`
	}
	if err := decodeJSONBody(r, &payload, 32<<10); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	switch payload.Op {
	case "view", "":
		cfg, err := s.Mgr.LoadConfig()
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		fp, err := s.Mgr.GetFingerprint(cfg)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"fingerprint": fp})
	case "reset":
		msg, err := s.Mgr.ResetFingerprint()
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": msg})
	default:
		writeJSONError(w, http.StatusBadRequest, "invalid fingerprint operation")
	}
}

func (s *Server) handleReleases(w http.ResponseWriter, r *http.Request) {
	cfg, err := s.Mgr.LoadConfig()
	if err != nil {
		cfg = config.Default()
	}

	source := strings.TrimSpace(r.URL.Query().Get("source"))
	if source == "" {
		source = cfg.DownloadSource
	}
	customPrefix := strings.TrimSpace(r.URL.Query().Get("custom_prefix"))
	if customPrefix == "" {
		customPrefix = cfg.CustomMirrorPrefix
	}
	preferGlibc := cfg.PreferGlibc
	if g := r.URL.Query().Get("prefer_glibc"); g != "" {
		preferGlibc = g == "1" || g == "true"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()
	releases, err := updater.FetchReleases(ctx, source, customPrefix, cfg.HTTPProxy, preferGlibc)
	if err != nil {
		writeJSONError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"source":   source,
		"releases": releases,
	})
}

func (s *Server) handleDownloadStart(w http.ResponseWriter, r *http.Request) {
	var req updater.JobRequest
	if err := decodeJSONBody(r, &req, 64<<10); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	job, err := updater.SpawnDownloadJob(s.Mgr, req)
	if err != nil {
		writeJSONError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"ok":  true,
		"job": job,
	})
}

func (s *Server) handleDownloadStatus(w http.ResponseWriter) {
	job := updater.ReadJobStatus(s.Mgr)
	var st *agentctl.Status
	if job.State == "done" || job.State == "error" {
		st, _ = s.Mgr.Status(false)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"job":    job,
		"status": st,
	})
}

func (s *Server) handleUpload(w http.ResponseWriter, r *http.Request) {
	const maxUploadBytes = 128 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

	mr, err := r.MultipartReader()
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("expected multipart upload: %v", err))
		return
	}

	if err := s.Mgr.Paths.EnsureDirs(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tmpDir, err := os.MkdirTemp(s.Mgr.Paths.RunDir(), "upload-*")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)

	var uploadPath, origName string
	for {
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("failed reading upload: %v", err))
			return
		}
		if part.FormName() != "file" {
			_ = part.Close()
			continue
		}
		origName = filepath.Base(part.FileName())
		if origName == "" || origName == "." {
			origName = "beszel-agent-upload"
		}
		uploadPath = filepath.Join(tmpDir, "upload.bin")
		f, err := os.OpenFile(uploadPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			_ = part.Close()
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		_, copyErr := io.Copy(f, part)
		closeErr := f.Close()
		_ = part.Close()
		if copyErr != nil {
			writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("upload interrupted: %v", copyErr))
			return
		}
		if closeErr != nil {
			writeJSONError(w, http.StatusInternalServerError, closeErr.Error())
			return
		}
		break
	}

	if uploadPath == "" {
		writeJSONError(w, http.StatusBadRequest, "missing 'file' form field")
		return
	}

	candidatePath := filepath.Join(s.Mgr.Paths.BinDir(), ".beszel-agent.upload")
	_ = os.Remove(candidatePath)
	if err := updater.ExtractBinary(uploadPath, origName, candidatePath); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer os.Remove(candidatePath)

	meta := &agentctl.BinaryMeta{
		Source:    "upload",
		AssetName: origName,
	}
	if err := updater.InstallPreparedBinary(s.Mgr, candidatePath, meta, true); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	st, _ := s.Mgr.Status(false)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": fmt.Sprintf("Installed beszel-agent %s from %s", meta.Version, origName),
		"status":  st,
	})
}

func (s *Server) handleRollback(w http.ResponseWriter) {
	if err := updater.RollbackBinary(s.Mgr); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}
	st, _ := s.Mgr.Status(false)
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":     true,
		"status": st,
	})
}

func (s *Server) handleGetLogs(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind != "manager" {
		kind = "agent"
	}
	linesCount := 200
	if raw := r.URL.Query().Get("lines"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			linesCount = n
		}
	}
	lines, err := s.Mgr.ReadLogs(kind, linesCount)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"kind":  kind,
		"lines": lines,
	})
}

func (s *Server) handleClearLogs(w http.ResponseWriter, r *http.Request) {
	kind := r.URL.Query().Get("kind")
	if kind != "manager" {
		kind = "agent"
	}
	if err := s.Mgr.ClearLogs(kind); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func decodeJSONBody(r *http.Request, dst any, maxBytes int64) error {
	defer r.Body.Close()
	data, err := io.ReadAll(io.LimitReader(r.Body, maxBytes))
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return errors.New("empty request body")
	}
	return json.Unmarshal(data, dst)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeJSONError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]any{
		"error": msg,
		"code":  code,
	})
}
