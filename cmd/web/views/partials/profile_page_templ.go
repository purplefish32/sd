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

func ProfilePage(
	instances []types.Instance,
	devices []types.Device,
	profiles []types.Profile,
	pages []types.Page,
	instance types.Instance,
	device types.Device,
	profile types.Profile,
	page types.Page,
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
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 2, "<!-- Right Panel - Device Config --><div class=\"flex-1 bg-sd-darker\" id=\"main-content\"><div class=\"flex flex-row w-full\"><div class=\"w-32 p-6 flex items-center justify-center text-gray-400\"><svg class=\"w-8 h-8\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M15 19l-7-7 7-7\"></path></svg></div><div class=\"flex-grow p-6 text-center text-gray-400\">")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			if device.Type == "pedal" {
				templ_7745c5c3_Err = StreamDeckPedal(instance, device, profile, page).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if device.Type == "xl" {
				templ_7745c5c3_Err = StreamDeckXL(instance, device, profile, page).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			if device.Type == "plus" {
				templ_7745c5c3_Err = StreamDeckPlus(instance, device, profile, page).Render(ctx, templ_7745c5c3_Buffer)
				if templ_7745c5c3_Err != nil {
					return templ_7745c5c3_Err
				}
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 3, "</div><div class=\"w-32 p-6 flex items-center justify-center text-gray-400\"><svg class=\"w-8 h-8\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 5l7 7-7 7\"></path></svg></div></div></div><div class=\"w-64 bg-sd-dark border-r border-sd-darker p-4\"><h2 class=\"text-xl font-semibold mb-4\"></h2><div><ul><li class=\"mb-2\"><div class=\"flex items-center p-2 bg-sd-light rounded cursor-pointer\" onclick=\"this.nextElementSibling.classList.toggle(&#39;hidden&#39;); this.querySelector(&#39;svg&#39;).classList.toggle(&#39;rotate-90&#39;)\"><svg class=\"w-4 h-4 mr-2 transform transition-transform duration-200\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 5l7 7-7 7\"></path></svg> <svg class=\"w-4 h-4 mr-2\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 12h18M3 6h18M3 18h18\"></path></svg> <span>Navigation</span></div><ul class=\"ml-4 mt-1 hidden\"><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Profile</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Page</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Single Action</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Toggle Action</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Multi Action</li></ul></li><li class=\"mb-2\"><div class=\"flex items-center p-2 bg-sd-light rounded cursor-pointer\" onclick=\"this.nextElementSibling.classList.toggle(&#39;hidden&#39;); this.querySelector(&#39;svg&#39;).classList.toggle(&#39;rotate-90&#39;)\"><svg class=\"w-4 h-4 mr-2 transform transition-transform duration-200\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 5l7 7-7 7\"></path></svg> <svg class=\"w-4 h-4 mr-2\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 12h18M3 6h18M3 18h18\"></path></svg> <span>Keyboard</span></div><ul class=\"ml-4 mt-1 hidden\"><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Shortcut</li><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Text</li></ul></li><li class=\"mb-2\"><div class=\"flex items-center p-2 bg-sd-light rounded cursor-pointer\" onclick=\"this.nextElementSibling.classList.toggle(&#39;hidden&#39;); this.querySelector(&#39;svg&#39;).classList.toggle(&#39;rotate-90&#39;)\"><svg class=\"w-4 h-4 mr-2 transform transition-transform duration-200\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M9 5l7 7-7 7\"></path></svg> <svg class=\"w-4 h-4 mr-2\" fill=\"none\" stroke=\"currentColor\" viewBox=\"0 0 24 24\"><path stroke-linecap=\"round\" stroke-linejoin=\"round\" stroke-width=\"2\" d=\"M3 12h18M3 6h18M3 18h18\"></path></svg> <span>Command</span></div><ul class=\"ml-4 mt-1 hidden\"><li class=\"p-2 hover:bg-sd-light rounded cursor-pointer\">Execute</li></ul></li></ul><button class=\"w-full p-3 mt-4 bg-sd-light hover:bg-sd-lighter text-white font-medium rounded transition-colors flex items-center justify-center gap-2\" hx-get=\"/\" hx-target=\"#dialog-container\" hx-trigger=\"click\" hx-swap=\"innerHTML\"><svg xmlns=\"http://www.w3.org/2000/svg\" class=\"h-5 w-5\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z\" clip-rule=\"evenodd\"></path></svg> Use</button> <button class=\"w-full p-3 mt-4 bg-sd-light hover:bg-sd-lighter text-white font-medium rounded transition-colors flex items-center justify-center gap-2\" hx-get=\"/\" hx-target=\"#dialog-container\" hx-trigger=\"click\" hx-swap=\"innerHTML\"><svg xmlns=\"http://www.w3.org/2000/svg\" class=\"h-5 w-5\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M10 3a1 1 0 011 1v5h5a1 1 0 110 2h-5v5a1 1 0 11-2 0v-5H4a1 1 0 110-2h5V4a1 1 0 011-1z\" clip-rule=\"evenodd\"></path></svg> Add Page</button> <button class=\"w-full p-3 mt-4 bg-sd-light hover:bg-sd-lighter text-white font-medium rounded transition-colors flex items-center justify-center gap-2\" hx-get=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var3 string
			templ_7745c5c3_Var3, templ_7745c5c3_Err = templ.JoinStringErrs("/partials/profile/delete?instanceId=" + instance.ID + "&deviceId=" + device.ID)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/partials/profile_page.templ`, Line: 129, Col: 94}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var3))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 4, "\" hx-target=\"#dialog-container\" hx-trigger=\"click\" hx-swap=\"innerHTML\"><svg xmlns=\"http://www.w3.org/2000/svg\" class=\"h-5 w-5\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M3 10a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z\" clip-rule=\"evenodd\"></path></svg> Delete Page</button> <button class=\"w-full p-3 mt-4 bg-red-600 hover:bg-red-700 text-white font-medium rounded transition-colors flex items-center justify-center gap-2\" hx-get=\"")
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			var templ_7745c5c3_Var4 string
			templ_7745c5c3_Var4, templ_7745c5c3_Err = templ.JoinStringErrs("/partials/profile/delete-dialog?instanceId=" + instance.ID + "&deviceId=" + device.ID + "&profileId=" + profile.ID)
			if templ_7745c5c3_Err != nil {
				return templ.Error{Err: templ_7745c5c3_Err, FileName: `views/partials/profile_page.templ`, Line: 141, Col: 130}
			}
			_, templ_7745c5c3_Err = templ_7745c5c3_Buffer.WriteString(templ.EscapeString(templ_7745c5c3_Var4))
			if templ_7745c5c3_Err != nil {
				return templ_7745c5c3_Err
			}
			templ_7745c5c3_Err = templruntime.WriteString(templ_7745c5c3_Buffer, 5, "\" hx-target=\"#dialog-container\" hx-trigger=\"click\" hx-swap=\"innerHTML\"><svg xmlns=\"http://www.w3.org/2000/svg\" class=\"h-5 w-5\" viewBox=\"0 0 20 20\" fill=\"currentColor\"><path fill-rule=\"evenodd\" d=\"M9 2a1 1 0 00-.894.553L7.382 4H4a1 1 0 000 2v10a2 2 0 002 2h8a2 2 0 002-2V6a1 1 0 100-2h-3.382l-.724-1.447A1 1 0 0011 2H9zM7 8a1 1 0 012 0v6a1 1 0 11-2 0V8zm5-1a1 1 0 00-1 1v6a1 1 0 102 0V8a1 1 0 00-1-1z\" clip-rule=\"evenodd\"></path></svg> Delete Profile</button></div></div></div>")
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
