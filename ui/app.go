package ui

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"material/config"
	"material/llm"
	"material/xhs"
)

// App 持有 GUI 所需的依赖与状态。
type App struct {
	cfg       *config.Config
	xhsClient *xhs.Client
	llmClient *llm.Client
	win       fyne.Window
	mu        sync.Mutex
	running   bool

	// UI 控件引用
	resultEntry *widget.Entry
	copyBtn     *widget.Button
	statusLabel *widget.Label
}

// NewApp 创建一个 App 实例。
func NewApp(cfg *config.Config) *App {
	xhsClient := xhs.NewClient()
	applyXHSConfig(xhsClient, &cfg.XHS)
	llmClient := llm.NewClient(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)
	return &App{
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

// Render 构建主界面并返回根容器。
func (a *App) Render(win fyne.Window) fyne.CanvasObject {
	a.win = win

	// 链接输入框
	linkEntry := widget.NewMultiLineEntry()
	linkEntry.SetPlaceHolder("粘贴小红书分享文本或链接，例如：\n67 【标题】 😆 JD5U1VyvHbeWo1v 😆 https://www.xiaohongshu.com/discovery/item/xxx?xsec_token=...")
	linkEntry.SetMinRowsVisible(4)

	// 生成按钮
	genBtn := widget.NewButtonWithIcon("生成分享文案", theme.DownloadIcon(), func() {
		a.onGenerate(linkEntry.Text)
	})

	// 结果展示区域
	resultEntry := widget.NewMultiLineEntry()
	resultEntry.Wrapping = fyne.TextWrapWord
	resultEntry.SetPlaceHolder("生成的分享文案会显示在这里...")
	resultEntry.SetMinRowsVisible(12)
	a.resultEntry = resultEntry

	// 复制按钮
	copyBtn := widget.NewButtonWithIcon("复制结果", theme.ContentCopyIcon(), func() {
		if resultEntry.Text == "" {
			return
		}
		win.Clipboard().SetContent(resultEntry.Text)
		dialog.ShowInformation("已复制", "文案已复制到剪贴板", win)
	})
	a.copyBtn = copyBtn
	copyBtn.Disable()

	// 状态栏
	statusLabel := widget.NewLabel("就绪")
	a.statusLabel = statusLabel

	// 清空按钮（在 resultEntry / copyBtn / statusLabel 之后定义，闭包才能引用）
	clearBtn := widget.NewButtonWithIcon("清空", theme.DeleteIcon(), func() {
		linkEntry.SetText("")
		resultEntry.SetText("")
		statusLabel.SetText("就绪")
		copyBtn.Disable()
	})

	// 顶部：输入框 + 按钮行
	top := container.NewVBox(
		linkEntry,
		container.NewHBox(genBtn, clearBtn, layout.NewSpacer(), copyBtn),
	)

	// 中部：结果区
	mid := container.NewBorder(nil, nil, nil, nil, resultEntry)

	// 底部：状态栏
	bottom := container.NewHBox(statusLabel, layout.NewSpacer())

	return container.NewBorder(top, bottom, nil, nil, mid)
}

// onGenerate 处理生成按钮点击。
func (a *App) onGenerate(shareText string) {
	a.mu.Lock()
	if a.running {
		a.mu.Unlock()
		return
	}
	a.running = true
	a.mu.Unlock()

	go func() {
		defer func() {
			a.mu.Lock()
			a.running = false
			a.mu.Unlock()
		}()
		a.runGenerate(strings.TrimSpace(shareText))
	}()
}

// runGenerate 执行完整的生成流程，更新 UI。
func (a *App) runGenerate(shareText string) {
	a.setStatus("正在解析链接...")
	a.setResult("")
	a.copyBtn.Disable()

	if shareText == "" {
		a.setError("请输入小红书分享文本或链接")
		return
	}

	// 1. 解析链接
	link, err := xhs.ParseShareText(shareText)
	if err != nil {
		a.setError(fmt.Sprintf("解析链接失败：%v", err))
		return
	}
	a.setStatus(fmt.Sprintf("已解析：note_id=%s", link.NoteID))

	// 2. 请求笔记详情
	a.setStatus("正在获取笔记详情...")
	body, err := a.xhsClient.GetNoteDetail(context.Background(),
		link.NoteID, link.XSecSource, link.XSecToken)
	if err != nil {
		a.setError(fmt.Sprintf("获取笔记详情失败：%v", err))
		return
	}

	// 3. 解析笔记
	feed, err := xhs.ParseFeedResponse(body)
	if err != nil {
		a.setError(fmt.Sprintf("解析笔记响应失败：%v", err))
		return
	}
	note := feed.FirstNote()
	if note == nil {
		a.setError("笔记响应中没有数据")
		return
	}
	noteText := note.ToLLMText()
	a.setStatus("已获取笔记内容，正在生成文案...")

	// 4. 调用 LLM
	copyText, err := a.llmClient.GenerateShareCopy(context.Background(), noteText)
	if err != nil {
		a.setError(fmt.Sprintf("生成文案失败：%v", err))
		return
	}

	// 5. 展示结果
	a.setResult(copyText)
	a.setStatus("完成")
	a.copyBtn.Enable()
}

// --- UI helper setters ---
// Fyne v2.5 中 widget 的 SetText 等方法可在 goroutine 中直接调用。

func (a *App) setStatus(text string) {
	a.statusLabel.SetText(text)
}

func (a *App) setResult(text string) {
	a.resultEntry.SetText(text)
}

func (a *App) setError(msg string) {
	a.resultEntry.SetText("错误：" + msg)
	a.statusLabel.SetText("出错")
}
