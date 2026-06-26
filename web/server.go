package web

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"

	"material/config"
	"material/llm"
	"material/xhs"
)

//go:embed index.html
var indexHTML []byte

// Server 持有 HTTP 服务所需的依赖。
type Server struct {
	cfg       *config.Config
	xhsClient *xhs.Client
	llmClient *llm.Client
}

// NewServer 创建一个 Server 实例。
func NewServer(cfg *config.Config) *Server {
	xhsClient := xhs.NewClient()
	applyXHSConfig(xhsClient, &cfg.XHS)
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)
	return &Server{
		cfg:       cfg,
		xhsClient: xhsClient,
		llmClient: llmClient,
	}
}

// applyXHSConfig 把配置中的签名头 / Cookie 注入到 client。
func applyXHSConfig(c *xhs.Client, x *config.XHSConfig) {
	c.Cookie = x.Cookie
	c.XS = x.XS
	c.XSCommon = x.XSCommon
	c.XT = x.XT
	c.XB3TraceID = x.XB3TraceID
	c.XRapParam = x.XRapParam
	c.XXrayTraceID = x.XXrayTrace
	c.XYDirection = x.XYDirection
}

// Serve 启动 HTTP 服务，监听 addr（如 ":8080"）。
func Serve(addr string, cfg *config.Config) error {
	srv := NewServer(cfg)
	mux := http.NewServeMux()
	mux.HandleFunc("/", srv.handleIndex)
	mux.HandleFunc("/api/generate", srv.handleGenerate)
	log.Printf("listening on %s", addr)
	return http.ListenAndServe(addr, mux)
}

// handleIndex 返回主页面。
func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(indexHTML)
}

// generateRequest 前端请求体。
type generateRequest struct {
	ShareText string `json:"share_text"`
}

// generateResponse 后端响应体。
type generateResponse struct {
	Copy  string `json:"copy,omitempty"`
	Note  string `json:"note,omitempty"`
	Error string `json:"error,omitempty"`
}

// handleGenerate 处理生成文案的 API 请求。
func (s *Server) handleGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, generateResponse{Error: "仅支持 POST 请求"})
		return
	}

	var req generateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, generateResponse{Error: "请求体解析失败: " + err.Error()})
		return
	}
	if req.ShareText == "" {
		writeJSON(w, http.StatusBadRequest, generateResponse{Error: "share_text 不能为空"})
		return
	}

	// 1. 解析链接
	link, err := xhs.ParseShareText(req.ShareText)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, generateResponse{Error: "解析链接失败: " + err.Error()})
		return
	}

	// 2. 请求笔记详情
	body, err := s.xhsClient.GetNoteDetail(r.Context(),
		link.NoteID, link.XSecSource, link.XSecToken)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, generateResponse{Error: "获取笔记详情失败: " + err.Error()})
		return
	}

	// 3. 解析笔记
	feed, err := xhs.ParseFeedResponse(body)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, generateResponse{Error: "解析笔记响应失败: " + err.Error()})
		return
	}
	note := feed.FirstNote()
	if note == nil {
		writeJSON(w, http.StatusBadGateway, generateResponse{Error: "笔记响应中没有数据"})
		return
	}
	noteText := note.ToLLMText()

	// 4. 调用 LLM 生成文案
	copyText, err := s.llmClient.GenerateShareCopy(context.Background(), noteText)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, generateResponse{Error: "生成文案失败: " + err.Error()})
		return
	}

	// 5. 返回结果
	writeJSON(w, http.StatusOK, generateResponse{Copy: copyText, Note: noteText})
}

// writeJSON 写入 JSON 响应。
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
