// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.819
//go:generate templ generate

package layouts

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

func Base(title string) templ.Component {
	return templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
		templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
		if templ_7745c5c3_CtxErr := ctx.Err(); templ_7745c5c3_CtxErr != nil {
			return templ_7745c5c3_CtxErr
		}
		templ_7745c5c3_Buffer, templ_7745c5c3_IsBuffer := templruntime.GetBuffer(templ_7745c5c3_W)
		if !templ_7745c5c3_IsBuffer {
			defer func() {
				templ_7745c5c3_BufErr := templruntime.ReleaseBuffer(templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err == nil {
					templ_7745c5c3_Err = templ_7745c5c3_BufErr
				}
			}()
		}
		ctx = templ.InitializeContext(ctx)
		templ_7745c5c3_Var1 := templ.GetChildren(ctx)
		if templ_7745c5c3_Var1 == nil {
			templ_7745c5c3_Var1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<!doctype html><html lang=\"en\"><head><meta charset=\"UTF-8\"><meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\"><title>Stream Deck - ")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		var templ_7745c5c3_Var2 string
		templ_7745c5c3_Var2, templ_7745c5c3_Err = templ.JoinStringErrs(title)
		if templ_7745c5c3_Err != nil {
			return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/layouts/base.layout.templ`, Line: 10, Col: 31}
		}
		_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var2))
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "</title><script src=\"https://cdn.tailwindcss.com\"></script><script src=\"https://unpkg.com/htmx.org@1.9.10\"></script><script src=\"https://unpkg.com/htmx.org@1.9.12/dist/ext/sse.js\"></script><script src=\"https://unpkg.com/hyperscript.org@0.9.13\"></script><script src=\"https://unpkg.com/htmx.org/dist/ext/debug.js\"></script><script src=\"https://unpkg.com/htmx.org@1.9.10/dist/ext/autofocus.js\"></script><script>\n\t\t\t\ttailwind.config = {\n\t\t\t\t\ttheme: {\n\t\t\t\t\t\textend: {\n\t\t\t\t\t\t\tcolors: {\n\t\t\t\t\t\t\t\t'sd-dark': '#1a1a1a',\n\t\t\t\t\t\t\t\t'sd-darker': '#0f0f0f',\n\t\t\t\t\t\t\t\t'sd-light': '#2d2d2d',\n\t\t\t\t\t\t\t\t'sd-accent': '#00ff00',\n\t\t\t\t\t\t\t},\n\t\t\t\t\t\t},\n\t\t\t\t\t},\n\t\t\t\t}\n\t\t\t</script></head><body class=\"bg-sd-darker text-white\" hx-boost=\"true\" hx-ext=\"debug, sse, autofocus\"><header class=\"bg-sd-dark border-b border-sd-darker p-4\"><div class=\"container mx-auto flex items-center justify-between\"><h1 class=\"text-2xl font-bold text-white tracking-wide\"><span class=\"text-sd-accent\">Open</span>Deck</h1><div class=\"flex items-center space-x-4\"><span class=\"text-sm text-gray-400\">v0.1.0</span></div></div></header><div class=\"flex h-[calc(100vh-4rem)]\"><!-- Sidebar --><aside class=\"w-16 bg-sd-dark border-r border-sd-darker flex flex-col items-center py-4\"><a href=\"/\" class=\"p-2 rounded hover:bg-sd-darker mb-4\"><svg class=\"w-6 h-6\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 12l2-2m0 0l7-7 7 7M5 10v10a1 1 0 001 1h3m10-11l2 2m-2-2v10a1 1 0 01-1 1h-3m-6 0a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1m-6 0h6\"></path></svg></a></aside><!-- Main Content --><main class=\"flex-1\">")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templ_7745c5c3_Var1.Render(ctx, templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</main></div></body></html>")
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
