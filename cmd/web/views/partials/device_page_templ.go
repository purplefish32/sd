// Code generated by templ - DO NOT EDIT.

// templ: version: v0.3.819
//go:generate templ generate

package partials

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import templruntime "github.com/a-h/templ/runtime"

import (
	"sd/cmd/web/views/layouts"
	"sd/pkg/types"
)

func DevicePage(
	instances []types.Instance,
	devices []types.Device,
	profiles []types.Profile,
	pages []types.Page,
	instance types.Instance,
	device *types.Device,
) templ.Component {
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
		templ_7745c5c3_Var2 := templruntime.GeneratedTemplate(func(templ_7745c5c3_Input templruntime.GeneratedComponentInput) (templ_7745c5c3_Err error) {
			templ_7745c5c3_W, ctx := templ_7745c5c3_Input.Writer, templ_7745c5c3_Input.Context
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
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 1, "<div class=\"flex h-full\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = InstancePanel(instances).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = DevicePanel(instance, devices).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = ProfilePanel(instance, device, profiles).Render(ctx, templ_7745c5c3_Buffer)
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<!-- Right Panel - Device Config --><div class=\"flex-1 bg-sd-darker\" id=\"main-content\"><div class=\"flex flex-row w-full\"></div></div><div class=\"w-64 bg-sd-dark border-r border-sd-darker p-4\"><h2 class=\"text-xl font-semibold mb-4\"></h2><div><ul><li class=\"mb-2\"><div class=\"flex items-center p-2 bg-sd-light rounded cursor-pointer\" onclick=\"this.nextElementSibling.classList.toggle(&#39;hidden&#39;); this.querySelector(&#39;svg&#39;).classList.toggle(&#39;rotate-90&#39;)\"><svg class=\"w-4 h-4 mr-2 transform transition-transform duration-200\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 5l7 7-7 7\"></path></svg> <svg class=\"w-4 h-4 mr-2\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 12h18M3 6h18M3 18h18\"></path></svg> <span>Navigation</span></div><ul class=\"ml-4 mt-1 hidden\"><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Profile</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Page</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Single Action</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Toggle Action</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Multi Action</li></ul></li><li class=\"mb-2\"><div class=\"flex items-center p-2 bg-sd-light rounded cursor-pointer\" onclick=\"this.nextElementSibling.classList.toggle(&#39;hidden&#39;); this.querySelector(&#39;svg&#39;).classList.toggle(&#39;rotate-90&#39;)\"><svg class=\"w-4 h-4 mr-2 transform transition-transform duration-200\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 5l7 7-7 7\"></path></svg> <svg class=\"w-4 h-4 mr-2\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 12h18M3 6h18M3 18h18\"></path></svg> <span>Keyboard</span></div><ul class=\"ml-4 mt-1 hidden\"><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Shortcut</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Text</li></ul></li><li class=\"mb-2\"><div class=\"flex items-center p-2 bg-sd-light rounded cursor-pointer\" onclick=\"this.nextElementSibling.classList.toggle(&#39;hidden&#39;); this.querySelector(&#39;svg&#39;).classList.toggle(&#39;rotate-90&#39;)\"><svg class=\"w-4 h-4 mr-2 transform transition-transform duration-200\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 5l7 7-7 7\"></path></svg> <svg class=\"w-4 h-4 mr-2\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 12h18M3 6h18M3 18h18\"></path></svg> <span>Command</span></div><ul class=\"ml-4 mt-1 hidden\"><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Execute</li></ul></li></ul></div></div></div>")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			return nil
		})
		templ_7745c5c3_Err = layouts.Base("Home").Render(templ.WithChildren(ctx, templ_7745c5c3_Var2), templ_7745c5c3_Buffer)
		if templ_7745c5c3_Err != nil {
			return templ_7745c5c3_Err
		}
		return nil
	})
}

var _ = templruntime.GeneratedTemplate
