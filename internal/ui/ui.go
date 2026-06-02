// Package ui is the platform's design system, built once and reused by every
// control-plane surface (admin panel, org console, and future IDE panels).
//
// Front-end decision: htmx over stdlib html/template — no SPA, no build step,
// dependency-free on the server (see docs/ui.md). This package owns the shared
// document shell and stylesheet so each surface only writes its own <body>.
package ui

// CSS is the single shared stylesheet for every htmx surface.
const CSS = `
  body{font:14px/1.5 system-ui,sans-serif;margin:0;background:#0f1117;color:#e6e6e6}
  header{padding:16px 24px;background:#161922;border-bottom:1px solid #262b38}
  h1{font-size:18px;margin:0}
  main{padding:24px;display:grid;gap:16px;grid-template-columns:repeat(auto-fit,minmax(320px,1fr))}
  .card{background:#161922;border:1px solid #262b38;border-radius:10px;padding:16px}
  .card h2{font-size:13px;text-transform:uppercase;letter-spacing:.05em;color:#8b93a7;margin:0 0 10px}
  .span2{grid-column:1/-1}
  .pill{display:inline-block;background:#222838;border:1px solid #2f3750;border-radius:999px;padding:2px 10px;margin:2px;font-size:12px}
  table{width:100%;border-collapse:collapse;font-size:13px}
  td,th{text-align:left;padding:6px 8px;border-bottom:1px solid #232838}
  input,select{background:#0f1320;border:1px solid #2a3346;color:#e6e6e6;border-radius:8px;padding:6px 10px;margin:2px}
  input[type=text]{flex:1}
  button{background:#2f6df6;color:#fff;border:0;border-radius:8px;padding:8px 14px;cursor:pointer}
  a{color:#7fb0ff}
  code{color:#9ad}
  .note{background:#2a1d1d;border:1px solid #5a2d2d;color:#f0bcbc;padding:8px 12px;border-radius:8px}
  #chat-log{height:280px;overflow:auto;display:flex;flex-direction:column;gap:8px;margin-bottom:10px}
  .msg{padding:8px 12px;border-radius:10px;max-width:90%;white-space:pre-wrap}
  .msg.user{align-self:flex-end;background:#2f6df6;color:#fff}
  .msg.bot{align-self:flex-start;background:#1c2230;border:1px solid #2a3346}
  .msg.bot.err{border-color:#5a2d2d;background:#2a1d1d;color:#f0bcbc}
  .trace{margin-top:6px;font-size:11px;color:#7f8aa3}
  form.chat{display:flex;gap:8px}
`

// Head returns the shared document head (doctype, meta, htmx, shared CSS) for the
// given title. The title may contain template actions; the result is concatenated
// with a surface's <body> before parsing. Surfaces append their own body and a
// closing </body></html>.
func Head(title string) string {
	return `<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + title + `</title>
<script src="https://unpkg.com/htmx.org@1.9.12"></script>
<style>` + CSS + `</style></head>`
}
