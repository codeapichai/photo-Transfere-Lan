package http

import (
	"bytes"
	"encoding/csv"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	"phototransferlan/backend/internal/application/service"
	"phototransferlan/backend/internal/domain/entity"
	wsapi "phototransferlan/backend/internal/presentation/websocket"

	fiberws "github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Dependencies struct {
	Auth      *service.AuthService
	Security  *service.SecurityService
	Settings  *service.SettingsService
	Logs      *service.LogService
	Uploads   *service.UploadService
	Dashboard *service.DashboardService
	Hub       *wsapi.Hub
}

func NewServer(deps Dependencies) *fiber.App {
	app := fiber.New(fiber.Config{BodyLimit: 64 * 1024 * 1024})
	app.Use(recover.New(recover.Config{EnableStackTrace: true, StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
		log.Printf("panic: %v", e)
	}}))
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://127.0.0.1:3000,http://tauri.localhost,https://tauri.localhost,tauri://localhost",
		AllowOriginsFunc: func(origin string) bool { return origin == "null" },
		AllowCredentials: true,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, X-CSRF-Token, X-Upload-Token",
		AllowMethods:     "GET,POST,PUT,OPTIONS",
	}))
	app.Get("/upload", mobileUploadPageHandler())

	api := app.Group("/api")
	api.Post("/setup", limiter.New(limiter.Config{Max: 5}), setupHandler(deps.Auth))
	api.Post("/auth/login", limiter.New(limiter.Config{Max: 10}), loginHandler(deps.Auth, deps.Security))
	api.Post("/auth/logout", requireSession(deps.Security), requireCSRF(deps.Security), logoutHandler(deps.Security))
	api.Get("/dashboard", requireSession(deps.Security), dashboardHandler(deps.Dashboard))
	api.Post("/tokens", requireSession(deps.Security), requireCSRF(deps.Security), createTokenHandler(deps.Security, deps.Dashboard))
	api.Get("/settings", requireSession(deps.Security), settingsGetHandler(deps.Settings))
	api.Put("/settings", requireSession(deps.Security), requireCSRF(deps.Security), settingsPutHandler(deps.Settings))
	api.Get("/logs", requireSession(deps.Security), logsHandler(deps.Logs))
	api.Get("/logs.csv", requireSession(deps.Security), logsCSVHandler(deps.Logs))
	api.Post("/upload-sessions", requireUploadAccess(deps.Security), createUploadSessionHandler(deps.Uploads))
	api.Put("/upload-sessions/:id/chunks/:index", requireUploadAccess(deps.Security), appendChunkHandler(deps.Uploads))
	api.Post("/upload-sessions/:id/complete", requireUploadAccess(deps.Security), completeUploadHandler(deps.Uploads))
	api.Get("/ws", fiberws.New(deps.Hub.Handle))

	return app
}

func mobileUploadPageHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
		return c.SendString(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>PhotoTransfer LAN Upload</title>
  <style>
    body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;margin:0;background:#f5f7f8;color:#17212b}
    main{max-width:560px;margin:0 auto;padding:24px}
    h1{font-size:30px;margin:0 0 8px}
    .drop{margin-top:24px;border:2px dashed #cbd5e1;background:white;border-radius:8px;min-height:220px;display:flex;align-items:center;justify-content:center;text-align:center;padding:24px}
    input{display:none}.bar{height:12px;border-radius:8px;background:#e2e8f0;overflow:hidden;margin-top:18px}.fill{height:100%;width:0;background:#36b37e;transition:width .2s}
    .status{margin-top:16px;color:#475569;word-break:break-word}.button{display:inline-block;margin-top:16px;background:#17212b;color:white;padding:10px 14px;border-radius:6px}
    .stats{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin-top:18px}.stat{background:white;border:1px solid #e2e8f0;border-radius:8px;padding:12px}.stat span{display:block;color:#64748b;font-size:13px}.stat strong{display:block;margin-top:4px;font-size:22px}
    .list{margin-top:18px;background:white;border:1px solid #e2e8f0;border-radius:8px;overflow:hidden}.row{display:flex;justify-content:space-between;gap:12px;padding:10px 12px;border-bottom:1px solid #f1f5f9}.row:last-child{border-bottom:0}.name{overflow:hidden;text-overflow:ellipsis;white-space:nowrap}.state{flex:0 0 auto;color:#64748b}
  </style>
</head>
<body>
<main>
  <h1>Upload to PC</h1>
  <p>Select photos or videos from this phone.</p>
  <label class="drop">
    <input id="file" type="file" multiple accept=".jpg,.jpeg,.png,.heic,.mov,.mp4,image/*,video/*" />
    <span>Tap to choose files</span>
  </label>
  <div class="stats">
    <div class="stat"><span>Total</span><strong id="total">0</strong></div>
    <div class="stat"><span>Uploaded</span><strong id="doneCount">0</strong></div>
    <div class="stat"><span>Remaining</span><strong id="remaining">0</strong></div>
  </div>
  <div class="bar"><div id="fill" class="fill"></div></div>
  <div id="status" class="status">Ready</div>
  <div id="list" class="list" hidden></div>
</main>
<script>
const token = new URLSearchParams(location.search).get("token") || "";
const input = document.getElementById("file");
const fill = document.getElementById("fill");
const status = document.getElementById("status");
const totalEl = document.getElementById("total");
const doneEl = document.getElementById("doneCount");
const remainingEl = document.getElementById("remaining");
const list = document.getElementById("list");
let selectedFiles = [];
let completedFiles = 0;
input.addEventListener("change", async () => {
  if (!token) { status.textContent = "Upload token is missing or expired. Scan the QR code again."; return; }
  selectedFiles = Array.from(input.files);
  completedFiles = 0;
  renderSummary();
  renderList();
  for (let i = 0; i < selectedFiles.length; i++) {
    const file = selectedFiles[i];
    try {
      setFileState(i, "Uploading");
      await upload(file, i);
      completedFiles++;
      setFileState(i, "Done");
    } catch (err) {
      completedFiles++;
      setFileState(i, "Failed");
      status.textContent = err.message || "Upload failed";
    }
    renderSummary();
  }
  status.textContent = "Finished " + completedFiles + " of " + selectedFiles.length + " files";
});
function renderSummary() {
  totalEl.textContent = selectedFiles.length;
  doneEl.textContent = completedFiles;
  remainingEl.textContent = Math.max(selectedFiles.length - completedFiles, 0);
}
function renderList() {
  list.hidden = selectedFiles.length === 0;
  list.innerHTML = selectedFiles.map((file, index) => '<div class="row" data-index="' + index + '"><span class="name">' + escapeHtml(file.name) + '</span><span class="state">Waiting</span></div>').join("");
}
function setFileState(index, text) {
  const row = list.querySelector('[data-index="' + index + '"] .state');
  if (row) row.textContent = text;
}
function escapeHtml(value) {
  return value.replace(/[&<>"']/g, (char) => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#039;'}[char]));
}
async function upload(file, fileIndex) {
  status.textContent = "Starting " + file.name;
  fill.style.width = "0%";
  const sessionRes = await fetch("/api/upload-sessions", {
    method: "POST",
    headers: {"Content-Type":"application/json","X-Upload-Token": token},
    body: JSON.stringify({filename:file.name,filesize:file.size,device_name:navigator.userAgent,duplicate_policy:"skip"})
  });
  if (!sessionRes.ok) throw new Error(await sessionRes.text());
  const session = await sessionRes.json();
  let offset = 0, index = 0;
  while (offset < file.size) {
    const chunk = file.slice(offset, offset + session.chunk_size);
    const chunkRes = await fetch("/api/upload-sessions/" + session.id + "/chunks/" + index, {
      method: "PUT",
      headers: {"X-Upload-Token": token},
      body: chunk
    });
    if (!chunkRes.ok) throw new Error(await chunkRes.text());
    offset += chunk.size; index++;
    const percent = Math.round((offset / file.size) * 100);
    fill.style.width = percent + "%";
    setFileState(fileIndex, percent + "%");
    status.textContent = "Uploading " + (completedFiles + 1) + " of " + selectedFiles.length + ": " + file.name + " " + percent + "%";
  }
  const done = await fetch("/api/upload-sessions/" + session.id + "/complete", {
    method:"POST",
    headers: {"Content-Type":"application/json","X-Upload-Token": token},
    body: JSON.stringify({duplicate_policy:"skip"})
  });
  if (!done.ok) throw new Error(await done.text());
  const result = await done.json();
  status.textContent = file.name + ": " + result.status;
}
</script>
</body>
</html>`)
	}
}

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func setupHandler(auth *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body setupRequest
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		err := auth.Setup(c.UserContext(), body.Username, body.Password)
		if errors.Is(err, service.ErrSetupComplete) {
			return fiber.NewError(fiber.StatusConflict, err.Error())
		}
		if errors.Is(err, service.ErrValidation) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.SendStatus(fiber.StatusCreated)
	}
}

func loginHandler(auth *service.AuthService, security *service.SecurityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body setupRequest
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		user, err := auth.LoginUser(c.UserContext(), body.Username, body.Password)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}
		session, err := security.CreateSession(c.UserContext(), user.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		c.Cookie(&fiber.Cookie{
			Name:     "pt_session",
			Value:    session.SessionID,
			HTTPOnly: true,
			SameSite: "Lax",
			Secure:   c.Protocol() == "https",
			Expires:  session.ExpiresAt,
		})
		c.Cookie(&fiber.Cookie{Name: "pt_csrf", Value: session.CSRFToken, SameSite: "Lax", Secure: c.Protocol() == "https", Expires: session.ExpiresAt})
		return c.JSON(session)
	}
}

func logoutHandler(security *service.SecurityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		_ = security.Logout(c.UserContext(), sessionIDFromRequest(c))
		c.ClearCookie("pt_session", "pt_csrf")
		return c.JSON(fiber.Map{"ok": true})
	}
}

func dashboardHandler(dashboard *service.DashboardService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := dashboard.Get(c.UserContext())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(data)
	}
}

func createTokenHandler(security *service.SecurityService, dashboard *service.DashboardService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := dashboard.Get(c.UserContext())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		token, err := security.CreateTemporaryToken(c.UserContext(), data.UploadURL)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(token)
	}
}

func settingsGetHandler(settings *service.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := settings.Get(c.UserContext())
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(data)
	}
}

func settingsPutHandler(settings *service.SettingsService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body entity.Settings
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		data, err := settings.Save(c.UserContext(), &body)
		if errors.Is(err, service.ErrValidation) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(data)
	}
}

func logsHandler(logs *service.LogService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		limit, _ := strconv.Atoi(c.Query("limit", "500"))
		data, err := logs.List(c.UserContext(), limit)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		return c.JSON(data)
	}
}

func logsCSVHandler(logs *service.LogService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		data, err := logs.List(c.UserContext(), 1000)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		var buf bytes.Buffer
		writer := csv.NewWriter(&buf)
		_ = writer.Write([]string{"id", "event_type", "actor", "message", "metadata_json", "created_at"})
		for _, row := range data {
			_ = writer.Write([]string{
				strconv.FormatUint(uint64(row.ID), 10),
				row.EventType,
				row.Actor,
				row.Message,
				row.MetadataJSON,
				row.CreatedAt.Format(time.RFC3339),
			})
		}
		writer.Flush()
		c.Set(fiber.HeaderContentType, "text/csv")
		c.Set(fiber.HeaderContentDisposition, `attachment; filename="phototransfer-logs.csv"`)
		return c.Send(buf.Bytes())
	}
}

type createUploadSessionRequest struct {
	Filename        string                 `json:"filename"`
	Filesize        int64                  `json:"filesize"`
	DeviceName      string                 `json:"device_name"`
	DuplicatePolicy entity.DuplicatePolicy `json:"duplicate_policy"`
}

func createUploadSessionHandler(uploads *service.UploadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body createUploadSessionRequest
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if body.DuplicatePolicy == "" {
			body.DuplicatePolicy = entity.DuplicateSkip
		}
		session, err := uploads.CreateSession(c.UserContext(), service.CreateUploadSessionInput{
			Filename: body.Filename, Filesize: body.Filesize, DeviceName: body.DeviceName, DuplicatePolicy: body.DuplicatePolicy,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.Status(fiber.StatusCreated).JSON(session)
	}
}

func appendChunkHandler(uploads *service.UploadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		body := c.Context().Request.BodyStream()
		if body == nil {
			body = bytes.NewReader(c.BodyRaw())
		}
		if err := uploads.AppendChunk(c.UserContext(), c.Params("id"), body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}

type completeUploadRequest struct {
	ExpectedSHA256  string                 `json:"expected_sha256"`
	DuplicatePolicy entity.DuplicatePolicy `json:"duplicate_policy"`
}

func completeUploadHandler(uploads *service.UploadService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var body completeUploadRequest
		if err := c.BodyParser(&body); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if body.DuplicatePolicy == "" {
			body.DuplicatePolicy = entity.DuplicateSkip
		}
		upload, err := uploads.Complete(c.UserContext(), c.Params("id"), body.ExpectedSHA256, body.DuplicatePolicy)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return c.JSON(upload)
	}
}

func requireSession(security *service.SecurityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		session, err := security.ValidateSession(c.UserContext(), sessionIDFromRequest(c))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "login required")
		}
		c.Locals("session", session)
		return c.Next()
	}
}

func requireCSRF(security *service.SecurityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		session, ok := c.Locals("session").(*entity.Session)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "login required")
		}
		if err := security.ValidateCSRF(session, c.Get("X-CSRF-Token")); err != nil {
			return fiber.NewError(fiber.StatusForbidden, "invalid csrf token")
		}
		return c.Next()
	}
}

func requireUploadAccess(security *service.SecurityService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if token := c.Get("X-Upload-Token"); token != "" {
			if err := security.ValidateTemporaryToken(c.UserContext(), token); err == nil {
				return c.Next()
			}
		}
		session, err := security.ValidateSession(c.UserContext(), sessionIDFromRequest(c))
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "upload token or login required")
		}
		if c.Method() != fiber.MethodPut {
			if err := security.ValidateCSRF(session, c.Get("X-CSRF-Token")); err != nil {
				return fiber.NewError(fiber.StatusForbidden, "invalid csrf token")
			}
		}
		return c.Next()
	}
}

func sessionIDFromRequest(c *fiber.Ctx) string {
	if value := c.Cookies("pt_session"); value != "" {
		return value
	}
	header := c.Get(fiber.HeaderAuthorization)
	if token, ok := strings.CutPrefix(header, "Bearer "); ok {
		return strings.TrimSpace(token)
	}
	return ""
}
